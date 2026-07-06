// Package sources defines the common interface implemented by every
// passive subdomain enumeration backend, plus the built-in implementations
// (Subfinder, Assetfinder, Subdominator, crt.sh).
//
// Adding a new source (AlienVault, SecurityTrails, VirusTotal, Censys,
// Chaos, BufferOver, RapidDNS, Anubis, ...) only requires implementing this
// interface and registering it where sources are wired up in the scanner.
package sources

import "context"

// Source represents a single passive subdomain data provider.
type Source interface {
	// Name returns a short, human-readable identifier for the source
	// (e.g. "subfinder"). Used in logs and summaries.
	Name() string

	// Run executes the source against the given domain and returns the
	// list of discovered subdomains. Implementations should respect
	// ctx cancellation/timeout and never panic; all failures must be
	// returned as an error so the scanner can continue with other
	// sources.
	Run(ctx context.Context, domain string) ([]string, error)
}
