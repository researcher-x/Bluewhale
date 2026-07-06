package utils

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// EnsureDir creates a directory (and any necessary parents) if it does not
// already exist.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("creating directory %q: %w", path, err)
	}
	return nil
}

// PathExists reports whether a file or directory exists at the given path.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// WriteLines writes a slice of strings to a file, one per line, overwriting
// any existing content. Empty slices still produce an (empty) file so that
// downstream tooling can rely on the file's existence.
func WriteLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", path, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range lines {
		if _, err := w.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("writing to file %q: %w", path, err)
		}
	}
	return w.Flush()
}

// ReadLines reads a file and returns its non-empty, trimmed lines. Missing
// files are treated as an empty result rather than an error, since a source
// may legitimately produce no output.
func ReadLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("opening file %q: %w", path, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	// Increase buffer size to handle unusually long lines safely.
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading file %q: %w", path, err)
	}
	return lines, nil
}

// MergeUnique merges multiple slices of strings into a single sorted slice
// with duplicates removed. Comparison is case-insensitive; the
// first-encountered casing is preserved in the output.
func MergeUnique(sets ...[]string) []string {
	seen := make(map[string]string) // lowercase -> original casing
	for _, set := range sets {
		for _, item := range set {
			trimmed := strings.TrimSpace(item)
			if trimmed == "" {
				continue
			}
			key := strings.ToLower(trimmed)
			if _, exists := seen[key]; !exists {
				seen[key] = trimmed
			}
		}
	}

	result := make([]string, 0, len(seen))
	for _, v := range seen {
		result = append(result, v)
	}
	sort.Strings(result)
	return result
}
