package sources

import "context"

// Subdominator wraps the "subdominator" CLI tool.
//
// Subdominator's actual CLI flags vary by version/installation method, so
// the invocation below targets its common "-d <domain>" passive-enumeration
// form. If your installed version uses different flags, adjust the args
// here — the rest of BlueWhale is unaffected since it only depends on the
// Source interface.
type Subdominator struct{}

// NewSubdominator constructs a Subdominator source.
func NewSubdominator() *Subdominator { return &Subdominator{} }

// Name implements Source.
func (s *Subdominator) Name() string { return "subdominator" }

// Run implements Source by invoking: subdominator -d <domain>
func (s *Subdominator) Run(ctx context.Context, domain string) ([]string, error) {
	return runCommandLines(ctx, "subdominator", "-d", domain)
}
