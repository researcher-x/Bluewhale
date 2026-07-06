<img src="/bluewhale/blue-whale.jpg">

# BlueWhale

Passive subdomain reconnaissance framework written in Go. BlueWhale
orchestrates existing passive OSINT tools — **Subfinder**, **Assetfinder**,
**Subdominator**, and **crt.sh** (certificate transparency logs) — runs them
concurrently, and merges the results into a single deduplicated subdomain
list.

BlueWhale performs no active scanning of its own: it only calls out to
already-installed third-party tools and public certificate-transparency
data, then organizes and deduplicates what they return.

## Requirements

BlueWhale expects the following tools to be installed and available on
`PATH`:

| Tool | Install |
|---|---|
| [subfinder](https://github.com/projectdiscovery/subfinder) | `go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest` |
| [assetfinder](https://github.com/tomnomnom/assetfinder) | `go install github.com/tomnomnom/assetfinder@latest` |
| [subdominator](https://github.com/RevoltSecurities/Subdominator) | `pip install subdominator` |

BlueWhale checks for these on startup and prints an installation hint for
anything missing.

## Build

```bash
go build -o bluewhale ./cmd/bluewhale
```

## Usage

```bash
bluewhale -d example.com              # run a scan
bluewhale -d example.com -v           # verbose, step-by-step logging
bluewhale -d example.com -o output/   # custom output directory
bluewhale --version
bluewhale --help
```

## Output layout

```
output/
└── example.com/
    ├── raw/
    │   ├── subfinder.txt
    │   ├── assetfinder.txt
    │   ├── subdominator.txt
    │   ├── crtsh.txt
    │   ├── crtsh_domains.txt
    │   ├── wildcard.txt
    │   └── wildcard_scan.txt
    ├── final/
    │   └── subdomains.txt
    └── logs/
        └── bluewhale.log
```

If a scan directory for the same domain already exists, BlueWhale renames it
aside (e.g. `example.com.bak-20260706-020000`) rather than silently
overwriting it, and prints a warning to that effect.

## How a scan works

1. Validate the target domain.
2. Create the output/raw/final/logs folder structure (backing up any prior
   scan first).
3. Verify required external tools are installed.
4. Run Subfinder, Assetfinder, Subdominator, and a crt.sh query
   concurrently, each under its own timeout; a failure in one source never
   stops the others.
5. Save each source's raw output to its own file under `raw/`.
6. Extract wildcard entries from crt.sh (e.g. `*.dev.example.com` →
   `dev.example.com`) and re-run Subfinder against that list
   (`subfinder -dL wildcard.txt`).
7. Read every raw file back from disk and merge them using Go maps (no
   shell `cat`), deduplicate case-insensitively, and sort alphabetically.
8. Write the final merged list to `final/subdomains.txt` and print a
   summary.

## Architecture

```
cmd/bluewhale/        CLI entry point (flag parsing, wiring)
internal/banner/      ASCII startup banner
internal/config/      Config + output-path construction
internal/deps/        External tool dependency checks
internal/logger/      File + console structured logging
internal/scanner/     Orchestration: concurrency, merge, summary
internal/ui/          Terminal spinner for non-verbose mode
internal/utils/       Domain validation, file I/O, merge/dedupe helpers
pkg/sources/          Source interface + Subfinder/Assetfinder/Subdominator/crt.sh
```

## Adding a new source

Every passive data provider implements a single interface:

```go
type Source interface {
    Name() string
    Run(ctx context.Context, domain string) ([]string, error)
}
```

To add a new source (AlienVault, SecurityTrails, VirusTotal, Censys, Chaos,
BufferOver, RapidDNS, Anubis, ...):

1. Create a new file under `pkg/sources/` implementing the `Source`
   interface (see `subfinder.go` or `assetfinder.go` for a minimal
   example).
2. Wire it into the concurrent execution block in
   `internal/scanner/scanner.go` following the existing goroutine pattern.
3. Add its raw output path to `internal/config/config.go` /
   `paths.go` and include that path in `mergeAllRawFiles`.

No other code needs to change — the dependency checker, logger, spinner,
merge logic, and summary report all operate generically over source
results.

## Notes

- No external Go modules are required — BlueWhale is built entirely on the
  standard library, so `go build` works offline once the module is
  vendored/cloned.
- Never panics: every external command and network call is wrapped in
  proper error handling, logged, and reported in the final summary without
  aborting the rest of the scan.
