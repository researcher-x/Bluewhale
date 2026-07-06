package sources

import "context"

// Assetfinder wraps the "assetfinder" CLI tool.
type Assetfinder struct{}

// NewAssetfinder constructs an Assetfinder source.
func NewAssetfinder() *Assetfinder { return &Assetfinder{} }

// Name implements Source.
func (a *Assetfinder) Name() string { return "assetfinder" }

// Run implements Source by invoking: assetfinder --subs-only <domain>
func (a *Assetfinder) Run(ctx context.Context, domain string) ([]string, error) {
	return runCommandLines(ctx, "assetfinder", "--subs-only", domain)
}
