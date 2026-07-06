package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// domainPattern is a pragmatic (not RFC-exhaustive) validator for hostnames
// such as "example.com" or "sub.example.co.uk".
var domainPattern = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$`)

// ValidateDomain checks that the supplied string looks like a syntactically
// valid domain name. It does not perform any DNS resolution.
func ValidateDomain(domain string) error {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	if len(domain) > 253 {
		return fmt.Errorf("domain %q is too long", domain)
	}
	if !domainPattern.MatchString(domain) {
		return fmt.Errorf("domain %q does not look like a valid domain name", domain)
	}
	return nil
}

// NormalizeDomain lowercases and trims a domain string.
func NormalizeDomain(domain string) string {
	return strings.ToLower(strings.TrimSpace(domain))
}

// StripWildcard converts a wildcard host such as "*.dev.example.com" into
// "dev.example.com". If no wildcard prefix is present, the input is
// returned unchanged (trimmed).
func StripWildcard(host string) string {
	host = strings.TrimSpace(host)
	return strings.TrimPrefix(host, "*.")
}

// IsWildcard reports whether a host string starts with the "*." prefix
// commonly seen in certificate transparency log entries.
func IsWildcard(host string) bool {
	return strings.HasPrefix(strings.TrimSpace(host), "*.")
}
