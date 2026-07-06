package sources

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// runCommandLines executes name with args under ctx and returns each
// non-empty, trimmed line of stdout. Stderr is captured separately and
// included in the returned error (if any) for easier debugging, but a
// non-zero exit code alone does not panic — it is surfaced as a normal
// error so the caller (scanner) can continue with other sources.
func runCommandLines(ctx context.Context, name string, args ...string) ([]string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %s: %w (stderr: %s)", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	var lines []string
	scanner := bufio.NewScanner(&stdout)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading output of %s: %w", name, err)
	}
	return lines, nil
}
