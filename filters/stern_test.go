package filters

import (
	"fmt"
	"strings"
	"testing"
)

func TestFilterSternEmpty(t *testing.T) {
	got, err := filterStern("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestFilterSternShort(t *testing.T) {
	var lines []string
	for i := 0; i < 10; i++ {
		lines = append(lines, fmt.Sprintf("myapp-abc%03d myapp 2024-01-15T10:23:%02dZ INFO Request GET /health 200 2ms", i, i))
	}
	raw := strings.Join(lines, "\n")

	got, err := filterStern(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ≤20 lines — should passthrough unchanged
	if got != raw {
		t.Errorf("expected passthrough for short input (≤20 lines), got:\n%s", got)
	}
}

func TestFilterSternDeduplication(t *testing.T) {
	// Build 20+ lines where the same INFO message appears from 3 pods
	// plus filler to reach the threshold
	var lines []string
	// Same message from 3 pods
	lines = append(lines, "myapp-abc123 myapp 2024-01-15T10:23:45Z INFO Starting server on :8080")
	lines = append(lines, "myapp-def456 myapp 2024-01-15T10:23:46Z INFO Starting server on :8080")
	lines = append(lines, "myapp-ghi789 myapp 2024-01-15T10:23:47Z INFO Starting server on :8080")
	// Filler lines to exceed the 20-line threshold
	for i := 0; i < 20; i++ {
		lines = append(lines, fmt.Sprintf("myapp-abc123 myapp 2024-01-15T10:24:%02dZ INFO Health check passed run=%d", i, i))
	}
	raw := strings.Join(lines, "\n")

	got, err := filterStern(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The repeated message should be deduplicated with a pod count
	if !strings.Contains(got, "[3 pods]") {
		t.Errorf("expected '[3 pods]' deduplication prefix, got:\n%s", got)
	}
	// The raw message should appear only once as a grouped line
	occurrences := strings.Count(got, "Starting server on :8080")
	if occurrences != 1 {
		t.Errorf("expected deduplicated message to appear once, got %d occurrences in:\n%s", occurrences, got)
	}

	// Token savings >= 50%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterSternAlwaysShowErrors(t *testing.T) {
	// Build 20+ lines with an ERROR that must be preserved
	var lines []string
	for i := 0; i < 18; i++ {
		lines = append(lines, fmt.Sprintf("myapp-abc123 myapp 2024-01-15T10:23:%02dZ INFO Request GET /health 200 2ms", i))
	}
	lines = append(lines, "myapp-abc123 myapp 2024-01-15T10:23:49Z ERROR Failed to process job: timeout")
	lines = append(lines, "myapp-def456 myapp 2024-01-15T10:23:50Z INFO Request GET /health 200 3ms")
	lines = append(lines, "myapp-ghi789 myapp 2024-01-15T10:23:51Z INFO Request GET /health 200 4ms")
	raw := strings.Join(lines, "\n")

	got, err := filterStern(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "ERROR") {
		t.Errorf("expected ERROR line preserved in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Failed to process job") {
		t.Errorf("expected error message in output, got:\n%s", got)
	}
}

func TestFilterSternMixedPods(t *testing.T) {
	// 3 pods repeating the same startup message + unique messages + filler
	var lines []string
	lines = append(lines, "myapp-abc123 myapp 2024-01-15T10:23:45Z INFO Starting server on :8080")
	lines = append(lines, "myapp-def456 myapp 2024-01-15T10:23:46Z INFO Starting server on :8080")
	lines = append(lines, "myapp-ghi789 myapp 2024-01-15T10:23:47Z INFO Starting server on :8080")
	lines = append(lines, "myapp-abc123 myapp 2024-01-15T10:23:48Z INFO Request GET /api/users 200 45ms")
	lines = append(lines, "myapp-def456 myapp 2024-01-15T10:23:50Z INFO Request GET /api/orders 200 12ms")
	// Filler to exceed 20 lines
	for i := 0; i < 18; i++ {
		lines = append(lines, fmt.Sprintf("myapp-abc123 myapp 2024-01-15T10:24:%02dZ INFO Health check ok seq=%d", i, i))
	}
	raw := strings.Join(lines, "\n")

	got, err := filterStern(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Repeated startup message should be grouped
	if !strings.Contains(got, "[3 pods]") {
		t.Errorf("expected '[3 pods]' for repeated startup message, got:\n%s", got)
	}
	// Unique messages: each pod had a different API request, so they should appear separately (not grouped as 3 pods)
	if strings.Count(got, "[3 pods]") > 1 {
		// Only the identical message should be grouped under 3 pods
		t.Errorf("only identical messages should be grouped as 3 pods, got:\n%s", got)
	}
}

func TestFilterSternSanityCheck(t *testing.T) {
	var lines []string
	for i := 0; i < 25; i++ {
		lines = append(lines, fmt.Sprintf("myapp-abc%03d myapp 2024-01-15T10:23:%02dZ INFO Request GET /health 200 %dms", i%3, i%60, i))
	}
	raw := strings.Join(lines, "\n")

	got, err := filterStern(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	if filteredTokens > rawTokens {
		t.Errorf("filter expanded output: raw=%d tokens, filtered=%d tokens", rawTokens, filteredTokens)
	}
}
