package config

import "path/filepath"

// BuildScanPaths derives every file/directory path used by a scan from the
// target domain and the user-supplied (or default) output root directory.
func BuildScanPaths(outputRoot, domain string) ScanPaths {
	root := filepath.Join(outputRoot, domain)
	raw := filepath.Join(root, "raw")
	final := filepath.Join(root, "final")
	logs := filepath.Join(root, "logs")

	return ScanPaths{
		Root:     root,
		RawDir:   raw,
		FinalDir: final,
		LogsDir:  logs,

		LogFile:   filepath.Join(logs, "bluewhale.log"),
		FinalFile: filepath.Join(final, "subdomains.txt"),

		Subfinder:    filepath.Join(raw, "subfinder.txt"),
		Assetfinder:  filepath.Join(raw, "assetfinder.txt"),
		Subdominator: filepath.Join(raw, "subdominator.txt"),
		Crtsh:        filepath.Join(raw, "crtsh.txt"),
		CrtshDomains: filepath.Join(raw, "crtsh_domains.txt"),
		Wildcard:     filepath.Join(raw, "wildcard.txt"),
		WildcardScan: filepath.Join(raw, "wildcard_scan.txt"),
	}
}
