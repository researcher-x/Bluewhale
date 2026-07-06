// Package banner prints the BlueWhale startup ASCII art and version line.
package banner

import (
	"fmt"

	"github.com/bluewhale/bluewhale/internal/config"
)

const art = `
   ______  _                _       __ _           __
  / __/ / (_)__ _  _____ __| | ____/ /(_)__  ___   / /
 _\ \/ / / / _ \ |/ / -_)___/ / __/ _ \/ / _ \/ -_) /_/
/___/_/_/_/\___/___/\__/    /_/_/ /_//_/_//_/\__/(_)
              B  L  U  E  W  H  A  L  E
`

// Print writes the ASCII banner and version information to stdout.
func Print() {
	fmt.Print(art)
	fmt.Printf("  BlueWhale v%s — Passive Subdomain Recon Framework\n\n", config.Version)
}
