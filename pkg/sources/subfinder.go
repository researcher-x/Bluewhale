package sources

import "context"

// Subfinder wraps the "subfinder" CLI tool.
type Subfinder struct{}

// NewSubfinder constructs a Subfinder source.
func NewSubfinder() *Subfinder { return &Subfinder{} }

// Name implements Source.
func (s *Subfinder) Name() string { return "subfinder" }

// Run implements Source by invoking: subfinder -d <domain> -silent
func (s *Subfinder) Run(ctx context.Context, domain string) ([]string, error) {
	return runCommandLines(ctx, "subfinder", "-d", domain, "-silent")
}

// RunList invokes subfinder against a list file of domains rather than a
// single domain, used for the wildcard re-enumeration step:
//
//	subfinder -dL <listFile> -silent
func (s *Subfinder) RunList(ctx context.Context, listFile string) ([]string, error) {
	return runCommandLines(ctx, "subfinder", "-dL", listFile, "-silent")
}
