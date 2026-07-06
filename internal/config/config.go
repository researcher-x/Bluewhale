package config

// Version is the current BlueWhale release version.
const Version = "1.0.0"

// Config holds all runtime configuration derived from CLI flags.
type Config struct {
	// Domain is the target domain to enumerate subdomains for.
	Domain string

	// Verbose enables detailed step-by-step logging to stdout.
	Verbose bool

	// OutputDir is the root directory under which scan results are stored.
	OutputDir string
}

// ScanPaths holds every directory/file path used during a single scan.
type ScanPaths struct {
	Root      string // output/<domain>
	RawDir    string // output/<domain>/raw
	FinalDir  string // output/<domain>/final
	LogsDir   string // output/<domain>/logs
	LogFile   string // output/<domain>/logs/bluewhale.log
	FinalFile string // output/<domain>/final/subdomains.txt

	Subfinder    string // raw/subfinder.txt
	Assetfinder  string // raw/assetfinder.txt
	Subdominator string // raw/subdominator.txt
	Crtsh        string // raw/crtsh.txt
	CrtshDomains string // raw/crtsh_domains.txt
	Wildcard     string // raw/wildcard.txt
	WildcardScan string // raw/wildcard_scan.txt
}
