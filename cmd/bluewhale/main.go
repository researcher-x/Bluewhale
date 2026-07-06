// Command bluewhale is a passive subdomain reconnaissance CLI that
// orchestrates Subfinder, Assetfinder, Subdominator, and crt.sh to produce
// a single deduplicated list of subdomains for a target domain.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bluewhale/bluewhale/internal/banner"
	"github.com/bluewhale/bluewhale/internal/config"
	"github.com/bluewhale/bluewhale/internal/deps"
	"github.com/bluewhale/bluewhale/internal/logger"
	"github.com/bluewhale/bluewhale/internal/scanner"
	"github.com/bluewhale/bluewhale/internal/ui"
	"github.com/bluewhale/bluewhale/internal/utils"
)

const defaultOutputDir = "output"

func main() {
	os.Exit(run())
}

// run contains the CLI logic and returns a process exit code. Keeping this
// separate from main() makes it possible to return non-zero codes without
// scattering os.Exit calls throughout the logic.
func run() int {
	var (
		domain    string
		verbose   bool
		outputDir string
		showVer   bool
	)

	fs := flag.NewFlagSet("bluewhale", flag.ContinueOnError)
	fs.StringVar(&domain, "d", "", "target domain to enumerate (required)")
	fs.BoolVar(&verbose, "v", false, "enable verbose output")
	fs.StringVar(&outputDir, "o", defaultOutputDir, "custom output directory")
	fs.BoolVar(&showVer, "version", false, "print version and exit")
	fs.Usage = printUsage

	if err := fs.Parse(os.Args[1:]); err != nil {
		return 2
	}

	if showVer {
		fmt.Printf("BlueWhale v%s\n", config.Version)
		return 0
	}

	if domain == "" {
		printUsage()
		return 2
	}

	domain = utils.NormalizeDomain(domain)
	if err := utils.ValidateDomain(domain); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	cfg := config.Config{Domain: domain, Verbose: verbose, OutputDir: outputDir}

	banner.Print()

	// 1 & 2: validate domain (done above) and prepare output directory,
	// warning rather than silently clobbering any previous scan.
	paths := config.BuildScanPaths(cfg.OutputDir, cfg.Domain)
	if utils.PathExists(paths.Root) {
		if err := backupExistingScan(paths.Root); err != nil {
			fmt.Fprintf(os.Stderr, "error: could not preserve previous scan: %v\n", err)
			return 1
		}
	}

	for _, dir := range []string{paths.RawDir, paths.FinalDir, paths.LogsDir} {
		if err := utils.EnsureDir(dir); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
	}

	log, err := logger.New(paths.LogFile, verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not open log file: %v\n", err)
		return 1
	}
	defer log.Close()

	log.Info("BlueWhale scan started for %s", domain)

	// 4. Dependency check.
	log.Info("Checking dependencies...")
	if missing := deps.Check(); len(missing) > 0 {
		fmt.Println("BlueWhale is missing required dependencies:")
		fmt.Println()
		for _, m := range missing {
			fmt.Printf("  - %v\n", m)
			log.Error("missing dependency: %v", m)
		}
		fmt.Println()
		fmt.Println("Install the tools above and re-run BlueWhale.")
		return 1
	}
	log.OK("all dependencies satisfied")

	// Spinner is only shown when verbose mode is off, per spec: verbose
	// mode replaces the spinner with detailed step-by-step log lines.
	var spinner *ui.Spinner
	if !verbose {
		spinner = ui.NewSpinner("Starting scan...")
		spinner.Start()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handleInterrupts(cancel)

	s := scanner.New(cfg, paths, log, spinner)
	summary, err := s.Run(ctx)

	if spinner != nil {
		spinner.Stop()
	}

	if err != nil {
		log.Error("scan failed: %v", err)
		fmt.Fprintf(os.Stderr, "error: scan failed: %v\n", err)
		return 1
	}

	log.OK("scan complete: %d unique subdomains", summary.UniqueSubdomains)

	fmt.Println()
	fmt.Print(scanner.FormatSummary(summary))

	return 0
}

// backupExistingScan renames a pre-existing scan directory out of the way
// (appending a timestamp suffix) and warns the user, rather than silently
// overwriting previous results.
func backupExistingScan(root string) error {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.bak-%s", root, timestamp)

	fmt.Printf("Warning: previous scan found at %q\n", root)
	fmt.Printf("         preserving it as %q\n\n", backupPath)

	return os.Rename(root, backupPath)
}

// handleInterrupts cancels ctx when the process receives SIGINT/SIGTERM,
// allowing in-flight external commands to be terminated gracefully instead
// of leaving orphaned processes or partially-written files.
func handleInterrupts(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nreceived interrupt, shutting down...")
		cancel()
	}()
}

func printUsage() {
	exe := filepath.Base(os.Args[0])
	fmt.Printf(`BlueWhale — Passive Subdomain Recon Framework

Usage:
  %s -d <domain> [flags]

Examples:
  %s -d example.com
  %s -d example.com -v
  %s -d example.com -o output/
  %s --version

Flags:
  -d string    Target domain (required)
  -v           Verbose mode
  -o string    Custom output directory (default %q)
  --version    Print version and exit
  --help       Show this help message
`, exe, exe, exe, exe, exe, defaultOutputDir)
}
