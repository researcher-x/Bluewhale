// Package deps verifies that external command-line tools BlueWhale relies
// on are installed and reachable via PATH.
package deps

import (
	"fmt"
	"os/exec"
)

// Requirement describes an external tool BlueWhale depends on, along with
// a human-friendly installation hint shown when it is missing.
type Requirement struct {
	Binary     string
	InstallMsg string
}

// Required lists every external tool BlueWhale expects to find on PATH.
var Required = []Requirement{
	{
		Binary:     "subfinder",
		InstallMsg: "install via: go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest",
	},
	{
		Binary:     "assetfinder",
		InstallMsg: "install via: go install github.com/tomnomnom/assetfinder@latest",
	},
	{
		Binary:     "subdominator",
		InstallMsg: "install via: pip install subdominator (or see https://github.com/RevoltSecurities/Subdominator)",
	},
}

// MissingError describes a single missing dependency and how to fix it.
type MissingError struct {
	Binary     string
	InstallMsg string
}

func (e MissingError) Error() string {
	return fmt.Sprintf("required tool %q not found in PATH — %s", e.Binary, e.InstallMsg)
}

// Check verifies that every required binary is available on PATH. It
// returns one MissingError per missing tool; an empty slice means all
// dependencies are satisfied.
func Check() []error {
	var missing []error
	for _, req := range Required {
		if _, err := exec.LookPath(req.Binary); err != nil {
			missing = append(missing, MissingError{Binary: req.Binary, InstallMsg: req.InstallMsg})
		}
	}
	return missing
}
