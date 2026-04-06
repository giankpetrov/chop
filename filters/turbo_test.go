package filters

import (
	"fmt"
	"strings"
	"testing"
)

func TestFilterTurboEmpty(t *testing.T) {
	got, err := filterTurbo("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty passthrough, got %q", got)
	}
}

func TestFilterTurboAllCached(t *testing.T) {
	// The compact "all cached" path requires the summary line to not contain the
	// word "failed" at all. Use a summary with no failure count field to trigger it.
	raw := `• Packages in scope: api, web, docs
• Running build in 3 packages
• Remote caching disabled

api:build: cache hit, replaying output abc123def
api:build: > next build
api:build: ✓ Compiled successfully

web:build: cache hit, replaying output def456ghi
web:build: > next build
web:build: ✓ Compiled successfully

docs:build: cache hit, replaying output ghi789jkl
docs:build: > next build
docs:build: ✓ Compiled successfully

 Tasks:    3 successful
 Cached:   3 cached, 0 not cached
 Time:     4.567s`

	got, err := filterTurbo(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All cached should produce compact output
	if !strings.Contains(got, "all cached") {
		t.Errorf("expected 'all cached' in compact output, got:\n%s", got)
	}
	// Should mention task count (3 tasks)
	if !strings.Contains(got, "3") {
		t.Errorf("expected task count in output, got:\n%s", got)
	}
	// Should mention time
	if !strings.Contains(got, "4.567s") {
		t.Errorf("expected elapsed time in output, got:\n%s", got)
	}
}

func TestFilterTurboWithFailure(t *testing.T) {
	raw := `• Packages in scope: api, web
• Running build in 2 packages
• Remote caching disabled

api:build: cache miss, executing abc123def
api:build: > next build
api:build:    Creating an optimized production build...
api:build: ✓ Compiled successfully

web:build: cache miss, executing def456ghi
web:build: > next build
web:build: ERROR: failed to compile
web:build: Error: Cannot find module './missing'

 Tasks:    1 successful, 1 failed
 Cached:   0 cached, 2 not cached
 Time:     8.910s`

	got, err := filterTurbo(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show the failed task
	if !strings.Contains(got, "web:build") {
		t.Errorf("expected failed task 'web:build' in output, got:\n%s", got)
	}

	// Should mark it as FAILED
	if !strings.Contains(got, "FAILED") {
		t.Errorf("expected FAILED status for web:build, got:\n%s", got)
	}

	// Error message should be present
	if !strings.Contains(got, "ERROR") && !strings.Contains(got, "Error:") {
		t.Errorf("expected error message in output, got:\n%s", got)
	}
}

func TestFilterTurboMixedCacheStatus(t *testing.T) {
	raw := `• Packages in scope: api, web, docs
• Running build in 3 packages
• Remote caching disabled

api:build: cache miss, executing abc123def
api:build: > next build
api:build: ✓ Compiled successfully

web:build: cache hit, replaying output def456ghi
web:build: > next build
web:build: ✓ Compiled successfully

docs:build: cache miss, executing ghi789jkl
docs:build: > next build
docs:build: ✓ Compiled successfully

 Tasks:    3 successful, 0 failed
 Cached:   1 cached, 2 not cached
 Time:     9.876s`

	got, err := filterTurbo(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Each task should appear with its cache status
	if !strings.Contains(got, "api:build") {
		t.Errorf("expected api:build in output, got:\n%s", got)
	}
	if !strings.Contains(got, "web:build") {
		t.Errorf("expected web:build in output, got:\n%s", got)
	}
	if !strings.Contains(got, "docs:build") {
		t.Errorf("expected docs:build in output, got:\n%s", got)
	}

	// Cache status labels
	if !strings.Contains(got, "cache hit") {
		t.Errorf("expected 'cache hit' label in output, got:\n%s", got)
	}
	if !strings.Contains(got, "cache miss") {
		t.Errorf("expected 'cache miss' label in output, got:\n%s", got)
	}

	// Should NOT be the all-cached compact path
	if strings.Contains(got, "all cached") {
		t.Errorf("should not use 'all cached' compact output when there are misses, got:\n%s", got)
	}
}

func TestFilterTurboTokenSavings(t *testing.T) {
	// 5 packages, each with 12+ lines of output
	var lines []string
	packages := []string{"api", "web", "docs", "admin", "shared"}

	lines = append(lines, "• Packages in scope: "+strings.Join(packages, ", "))
	lines = append(lines, fmt.Sprintf("• Running build in %d packages", len(packages)))
	lines = append(lines, "• Remote caching disabled")
	lines = append(lines, "")

	for i, pkg := range packages {
		task := pkg + ":build"
		lines = append(lines, fmt.Sprintf("%s: cache miss, executing hash%03d", task, i))
		lines = append(lines, fmt.Sprintf("%s: > next build", task))
		lines = append(lines, fmt.Sprintf("%s:    Creating an optimized production build...", task))
		lines = append(lines, fmt.Sprintf("%s:    Compiled all source files", task))
		lines = append(lines, fmt.Sprintf("%s:    Running type checks", task))
		lines = append(lines, fmt.Sprintf("%s:    Checking for circular imports", task))
		lines = append(lines, fmt.Sprintf("%s:    Bundling output files", task))
		lines = append(lines, fmt.Sprintf("%s:    Generating source maps", task))
		lines = append(lines, fmt.Sprintf("%s:    Writing build artifacts", task))
		lines = append(lines, fmt.Sprintf("%s:    Optimizing assets", task))
		lines = append(lines, fmt.Sprintf("%s:    Verifying bundle integrity", task))
		lines = append(lines, fmt.Sprintf("%s: ✓ Compiled successfully", task))
		lines = append(lines, "")
	}

	lines = append(lines, fmt.Sprintf(" Tasks:    %d successful, 0 failed", len(packages)))
	lines = append(lines, fmt.Sprintf(" Cached:   0 cached, %d not cached", len(packages)))
	lines = append(lines, " Time:     45.678s")

	raw := strings.Join(lines, "\n")

	got, err := filterTurbo(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)\noutput:\n%s", savings, rawTokens, filteredTokens, got)
	}
}

func TestFilterTurboSanityCheck(t *testing.T) {
	raw := `• Packages in scope: api, web
• Running build in 2 packages
• Remote caching disabled

api:build: cache hit, replaying output abc123
api:build: ✓ Compiled successfully

web:build: cache hit, replaying output def456
web:build: ✓ Compiled successfully

 Tasks:    2 successful, 0 failed
 Cached:   2 cached, 0 not cached
 Time:     1.234s`

	got, err := filterTurbo(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	if filteredTokens > rawTokens {
		t.Errorf("filter expanded output: raw=%d tokens, filtered=%d tokens\noutput:\n%s", rawTokens, filteredTokens, got)
	}
}
