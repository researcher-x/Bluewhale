package scanner

import (
	"fmt"
	"strings"
)

// displayName maps internal source identifiers to the labels used in the
// final human-readable summary.
var displayName = map[string]string{
	"subfinder":     "Subfinder",
	"assetfinder":   "Assetfinder",
	"subdominator":  "Subdominator",
	"crt.sh":        "crt.sh",
	"wildcard-scan": "Wildcard Scan",
}

// FormatSummary renders the scan Summary in the CLI's final-report layout.
func FormatSummary(sum *Summary) string {
	var b strings.Builder

	fmt.Fprintln(&b, "BlueWhale Scan Complete")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Target:")
	fmt.Fprintln(&b, sum.Domain)
	fmt.Fprintln(&b)

	for _, r := range sum.Results {
		label := displayName[r.Name]
		if label == "" {
			label = r.Name
		}
		fmt.Fprintf(&b, "%s:\n", label)
		switch {
		case r.Err != nil:
			fmt.Fprintln(&b, "failed")
		case r.Skipped:
			fmt.Fprintln(&b, "skipped")
		default:
			fmt.Fprintf(&b, "%d\n", r.Count)
		}
		fmt.Fprintln(&b)
	}

	fmt.Fprintln(&b, "Unique Subdomains:")
	fmt.Fprintln(&b, sum.UniqueSubdomains)
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "Duration:")
	fmt.Fprintln(&b, sum.Duration)
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "Saved:")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, sum.FinalFilePath)

	return b.String()
}
