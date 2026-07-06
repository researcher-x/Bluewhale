// Package scanner orchestrates a full BlueWhale scan: dependency checks,
// concurrent execution of passive sources, crt.sh wildcard follow-up,
// merging/deduplication, and final result persistence.
package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bluewhale/bluewhale/internal/config"
	"github.com/bluewhale/bluewhale/internal/logger"
	"github.com/bluewhale/bluewhale/internal/ui"
	"github.com/bluewhale/bluewhale/internal/utils"
	"github.com/bluewhale/bluewhale/pkg/sources"
)

// sourceTimeout bounds how long any single external tool invocation may
// run before it is cancelled. Individual source failures never abort the
// overall scan.
const sourceTimeout = 5 * time.Minute

// Scanner coordinates a single end-to-end BlueWhale scan.
type Scanner struct {
	cfg     config.Config
	paths   config.ScanPaths
	log     *logger.Logger
	spinner *ui.Spinner
}

// New creates a Scanner ready to run against the given configuration.
func New(cfg config.Config, paths config.ScanPaths, log *logger.Logger, spinner *ui.Spinner) *Scanner {
	return &Scanner{cfg: cfg, paths: paths, log: log, spinner: spinner}
}

// rawResult carries the outcome of one concurrently-executed source.
type rawResult struct {
	name  string
	lines []string
	err   error
}

// Run executes the full BlueWhale workflow and returns a Summary describing
// the outcome. It never panics; individual source failures are recorded in
// the summary rather than aborting the scan.
func (s *Scanner) Run(ctx context.Context) (*Summary, error) {
	start := time.Now()

	if err := s.prepareDirectories(); err != nil {
		return nil, err
	}

	results := make([]SourceResult, 0, 5)
	var resultsMu sync.Mutex
	addResult := func(r SourceResult) {
		resultsMu.Lock()
		defer resultsMu.Unlock()
		results = append(results, r)
	}

	s.setStatus("Running passive sources...")

	var wg sync.WaitGroup
	resultCh := make(chan rawResult, 4)

	// 1. Subfinder
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.log.Info("Running Subfinder...")
		sfCtx, cancel := context.WithTimeout(ctx, sourceTimeout)
		defer cancel()
		lines, err := sources.NewSubfinder().Run(sfCtx, s.cfg.Domain)
		resultCh <- rawResult{name: "subfinder", lines: lines, err: err}
	}()

	// 2. Assetfinder
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.log.Info("Running Assetfinder...")
		afCtx, cancel := context.WithTimeout(ctx, sourceTimeout)
		defer cancel()
		lines, err := sources.NewAssetfinder().Run(afCtx, s.cfg.Domain)
		resultCh <- rawResult{name: "assetfinder", lines: lines, err: err}
	}()

	// 3. Subdominator
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.log.Info("Running Subdominator...")
		sdCtx, cancel := context.WithTimeout(ctx, sourceTimeout)
		defer cancel()
		lines, err := sources.NewSubdominator().Run(sdCtx, s.cfg.Domain)
		resultCh <- rawResult{name: "subdominator", lines: lines, err: err}
	}()

	// 4. crt.sh (also captures wildcard + domain views for the follow-up
	// wildcard re-enumeration step).
	var crtAll, crtDomains, crtWildcards []string
	var crtErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.log.Info("Running crt.sh...")
		chCtx, cancel := context.WithTimeout(ctx, sourceTimeout)
		defer cancel()
		all, domains, wildcards, err := sources.NewCrtSh().Fetch(chCtx, s.cfg.Domain)
		crtAll, crtDomains, crtWildcards, crtErr = all, domains, wildcards, err
		resultCh <- rawResult{name: "crt.sh", lines: domains, err: err}
	}()

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	rawByName := make(map[string][]string)
	for r := range resultCh {
		rawByName[r.name] = r.lines
		if r.err != nil {
			s.log.Error("%s failed: %v", r.name, r.err)
			addResult(SourceResult{Name: r.name, Err: r.err})
		} else {
			s.log.OK("%s found %d subdomains", r.name, len(r.lines))
			addResult(SourceResult{Name: r.name, Count: len(r.lines)})
		}
	}

	// Persist raw outputs for each source (even empty, so folder layout is
	// always complete and predictable).
	if err := utils.WriteLines(s.paths.Subfinder, rawByName["subfinder"]); err != nil {
		s.log.Error("saving subfinder output: %v", err)
	}
	if err := utils.WriteLines(s.paths.Assetfinder, rawByName["assetfinder"]); err != nil {
		s.log.Error("saving assetfinder output: %v", err)
	}
	if err := utils.WriteLines(s.paths.Subdominator, rawByName["subdominator"]); err != nil {
		s.log.Error("saving subdominator output: %v", err)
	}

	// crt.sh specific outputs.
	s.log.Info("Extracting wildcard entries...")
	if crtErr != nil {
		s.log.Error("crt.sh failed: %v", crtErr)
	}
	if err := utils.WriteLines(s.paths.Crtsh, crtAll); err != nil {
		s.log.Error("saving crtsh.txt: %v", err)
	}
	if err := utils.WriteLines(s.paths.CrtshDomains, crtDomains); err != nil {
		s.log.Error("saving crtsh_domains.txt: %v", err)
	}
	if err := utils.WriteLines(s.paths.Wildcard, crtWildcards); err != nil {
		s.log.Error("saving wildcard.txt: %v", err)
	}

	// 5. Wildcard re-enumeration via subfinder -dL wildcard.txt
	s.setStatus("Enumerating wildcard domains...")
	var wildcardScan []string
	if len(crtWildcards) > 0 {
		s.log.Info("Running Subfinder against wildcard list...")
		wcCtx, cancel := context.WithTimeout(ctx, sourceTimeout)
		lines, err := sources.NewSubfinder().RunList(wcCtx, s.paths.Wildcard)
		cancel()
		if err != nil {
			s.log.Error("wildcard subfinder run failed: %v", err)
			addResult(SourceResult{Name: "wildcard-scan", Err: err})
		} else {
			wildcardScan = lines
			s.log.OK("wildcard scan found %d subdomains", len(lines))
			addResult(SourceResult{Name: "wildcard-scan", Count: len(lines)})
		}
	} else {
		s.log.Info("No wildcard entries found; skipping wildcard re-enumeration.")
		addResult(SourceResult{Name: "wildcard-scan", Skipped: true})
	}
	if err := utils.WriteLines(s.paths.WildcardScan, wildcardScan); err != nil {
		s.log.Error("saving wildcard_scan.txt: %v", err)
	}

	// 6-10. Merge everything natively in Go (no shell "cat"), dedupe, sort.
	s.setStatus("Merging and deduplicating results...")
	s.log.Info("Merging outputs...")
	s.log.Info("Removing duplicates...")

	merged, err := s.mergeAllRawFiles()
	if err != nil {
		return nil, fmt.Errorf("merging results: %w", err)
	}

	// 11. Save final subdomains.
	s.log.Info("Saving final results...")
	if err := utils.WriteLines(s.paths.FinalFile, merged); err != nil {
		return nil, fmt.Errorf("writing final results: %w", err)
	}

	summary := &Summary{
		Domain:           s.cfg.Domain,
		Results:          results,
		UniqueSubdomains: len(merged),
		FinalFilePath:    s.paths.FinalFile,
		Duration:         time.Since(start).Round(time.Millisecond).String(),
	}

	return summary, nil
}

// prepareDirectories creates the raw/final/logs folder structure for this
// scan. Warning-on-overwrite (moving aside a prior scan) is handled earlier
// in main, before the logger/scanner are constructed, so this is purely
// idempotent directory creation.
func (s *Scanner) prepareDirectories() error {
	for _, dir := range []string{s.paths.RawDir, s.paths.FinalDir, s.paths.LogsDir} {
		if err := utils.EnsureDir(dir); err != nil {
			return err
		}
	}
	return nil
}

// mergeAllRawFiles reads every raw/*.txt file back from disk and merges
// them into a single deduplicated, alphabetically sorted slice. Reading
// from disk (rather than reusing in-memory slices) satisfies the
// requirement that merging happens over "every txt file generated" and
// keeps the merge step independent/idempotent.
func (s *Scanner) mergeAllRawFiles() ([]string, error) {
	rawFiles := []string{
		s.paths.Subfinder,
		s.paths.Assetfinder,
		s.paths.Subdominator,
		s.paths.Crtsh,
		s.paths.CrtshDomains,
		s.paths.Wildcard,
		s.paths.WildcardScan,
	}

	sets := make([][]string, 0, len(rawFiles))
	for _, path := range rawFiles {
		lines, err := utils.ReadLines(path)
		if err != nil {
			return nil, err
		}
		sets = append(sets, lines)
	}

	return utils.MergeUnique(sets...), nil
}

func (s *Scanner) setStatus(msg string) {
	if s.spinner != nil {
		s.spinner.UpdateMessage(msg)
	}
}
