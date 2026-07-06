package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// crtShEntry mirrors the subset of fields BlueWhale cares about from
// crt.sh's JSON API response.
type crtShEntry struct {
	NameValue string `json:"name_value"`
}

// CrtSh queries the crt.sh certificate transparency log search API.
type CrtSh struct {
	// HTTPClient allows tests/customization to override the client used
	// for requests. If nil, http.DefaultClient is used.
	HTTPClient *http.Client
}

// NewCrtSh constructs a CrtSh source.
func NewCrtSh() *CrtSh {
	return &CrtSh{HTTPClient: http.DefaultClient}
}

// Name implements Source.
func (c *CrtSh) Name() string { return "crt.sh" }

// Run implements Source. For a plain Source consumer this returns only the
// non-wildcard subdomains discovered. Callers that also need the wildcard
// entries (to drive the wildcard re-enumeration step) should use Fetch
// directly instead.
func (c *CrtSh) Run(ctx context.Context, domain string) ([]string, error) {
	_, domains, _, err := c.Fetch(ctx, domain)
	return domains, err
}

// Fetch downloads and parses crt.sh results for domain, returning three
// views of the data:
//
//   - all:       every unique name found in the certificate log (as-is)
//   - domains:   unique non-wildcard subdomains
//   - wildcards: unique wildcard entries with the "*." prefix stripped
//     (e.g. "*.dev.example.com" -> "dev.example.com")
func (c *CrtSh) Fetch(ctx context.Context, domain string) (all, domains, wildcards []string, err error) {
	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("building crt.sh request: %w", err)
	}

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("querying crt.sh: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, nil, fmt.Errorf("crt.sh returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("reading crt.sh response: %w", err)
	}

	var entries []crtShEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, nil, nil, fmt.Errorf("parsing crt.sh JSON: %w", err)
	}

	allSet := make(map[string]struct{})
	domainSet := make(map[string]struct{})
	wildcardSet := make(map[string]struct{})

	for _, entry := range entries {
		// name_value can contain multiple newline-separated hostnames.
		for _, raw := range strings.Split(entry.NameValue, "\n") {
			name := strings.ToLower(strings.TrimSpace(raw))
			if name == "" {
				continue
			}
			allSet[name] = struct{}{}

			if strings.HasPrefix(name, "*.") {
				wildcardSet[strings.TrimPrefix(name, "*.")] = struct{}{}
			} else {
				domainSet[name] = struct{}{}
			}
		}
	}

	all = setToSortedSlice(allSet)
	domains = setToSortedSlice(domainSet)
	wildcards = setToSortedSlice(wildcardSet)
	return all, domains, wildcards, nil
}

func setToSortedSlice(set map[string]struct{}) []string {
	result := make([]string, 0, len(set))
	for k := range set {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
