package filters

import (
	"fmt"
	"strings"
	"testing"
)

func TestFilterGolangciLintEmpty(t *testing.T) {
	got, err := filterGolangciLint("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no issues" {
		t.Errorf("expected 'no issues', got %q", got)
	}
}

func TestFilterGolangciLintNoIssues(t *testing.T) {
	// Input must contain ".go:" to trigger the filter; use a progress line that
	// looks like a file path but carries no issue annotation.
	raw := `pkg/server/server.go:checking...
Run MegaCheck...
level=warning msg="[runner] skipping..."
level=info msg="File cache is not empty, 10 entries"
golangci-lint took 3.2s

0 issues.`

	got, err := filterGolangciLint(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no issues" {
		t.Errorf("expected 'no issues', got %q", got)
	}
}

func TestFilterGolangciLintCompressed(t *testing.T) {
	// 20+ issues across 5+ linters
	raw := `pkg/server/server.go:42:9: Error return value of ` + "`" + `db.Close` + "`" + ` is not checked (errcheck)
pkg/server/server.go:55:3: Error return value of ` + "`" + `tx.Rollback` + "`" + ` is not checked (errcheck)
pkg/server/server.go:71:5: Error return value of ` + "`" + `rows.Close` + "`" + ` is not checked (errcheck)
pkg/server/server.go:88:11: Error return value of ` + "`" + `stmt.Close` + "`" + ` is not checked (errcheck)
pkg/server/server.go:102:7: Error return value of ` + "`" + `f.Close` + "`" + ` is not checked (errcheck)
cmd/root.go:15:2: should not use dot imports (stylecheck)
cmd/root.go:28:2: should not use dot imports (stylecheck)
cmd/root.go:34:2: comment on exported function should be of the form (stylecheck)
cmd/root.go:41:2: comment on exported type should be of the form (stylecheck)
cmd/server.go:9:2: should not use dot imports (stylecheck)
internal/util/util.go:88:12: ineffectual assignment to ` + "`" + `err` + "`" + ` (ineffassign)
internal/util/util.go:92:5: ineffectual assignment to ` + "`" + `err` + "`" + ` (ineffassign)
internal/util/util.go:107:3: ineffectual assignment to ` + "`" + `result` + "`" + ` (ineffassign)
internal/cache/cache.go:23:9: ineffectual assignment to ` + "`" + `val` + "`" + ` (ineffassign)
internal/cache/cache.go:45:6: ineffectual assignment to ` + "`" + `ok` + "`" + ` (ineffassign)
pkg/client/client.go:12:1: exported function NewClient should have comment (golint)
pkg/client/client.go:34:1: exported type Client should have comment (golint)
pkg/client/client.go:56:1: exported method Client.Do should have comment (golint)
pkg/client/client.go:78:1: exported function DefaultConfig should have comment (golint)
internal/util/util.go:12:1: exported function ParseFlags should have comment (golint)
cmd/root.go:5:2: SA1006: printf with dynamic first argument (govet)
cmd/root.go:67:4: SA1006: printf with dynamic first argument (govet)

Run MegaCheck...
level=warning msg="[runner] skipping..."

22 issues.`

	got, err := filterGolangciLint(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Groups by linter
	if !strings.Contains(got, "errcheck") {
		t.Errorf("expected errcheck linter in output, got:\n%s", got)
	}
	if !strings.Contains(got, "stylecheck") {
		t.Errorf("expected stylecheck linter in output, got:\n%s", got)
	}
	if !strings.Contains(got, "ineffassign") {
		t.Errorf("expected ineffassign linter in output, got:\n%s", got)
	}
	if !strings.Contains(got, "golint") {
		t.Errorf("expected golint linter in output, got:\n%s", got)
	}
	if !strings.Contains(got, "govet") {
		t.Errorf("expected govet linter in output, got:\n%s", got)
	}

	// Summary line
	if !strings.Contains(got, "issue(s)") {
		t.Errorf("expected 'issue(s)' summary in output, got:\n%s", got)
	}

	// Shows counts
	if !strings.Contains(got, "(5)") && !strings.Contains(got, "(22)") {
		// errcheck has 5 issues; summary has 22
		if !strings.Contains(got, "22 issue(s)") {
			t.Errorf("expected total issue count in output, got:\n%s", got)
		}
	}

	// Sample locations present
	if !strings.Contains(got, "pkg/server/server.go") && !strings.Contains(got, "cmd/root.go") {
		t.Errorf("expected sample file locations in output, got:\n%s", got)
	}

	// Token savings >= 60%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 60.0 {
		t.Errorf("expected >=60%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterGolangciLintFewIssues(t *testing.T) {
	raw := `pkg/server/server.go:42:9: Error return value of ` + "`" + `db.Close` + "`" + ` is not checked (errcheck)
pkg/server/server.go:55:3: Error return value of ` + "`" + `tx.Rollback` + "`" + ` is not checked (errcheck)
cmd/root.go:15:2: should not use dot imports (stylecheck)

3 issues.`

	got, err := filterGolangciLint(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still compress: grouped by linter
	if !strings.Contains(got, "errcheck") {
		t.Errorf("expected errcheck group in output, got:\n%s", got)
	}
	if !strings.Contains(got, "stylecheck") {
		t.Errorf("expected stylecheck group in output, got:\n%s", got)
	}

	// Shows counts per linter
	if !strings.Contains(got, "(2)") {
		t.Errorf("expected errcheck count (2) in output, got:\n%s", got)
	}
	if !strings.Contains(got, "(1)") {
		t.Errorf("expected stylecheck count (1) in output, got:\n%s", got)
	}

	// Summary
	if !strings.Contains(got, "3 issue(s)") {
		t.Errorf("expected '3 issue(s)' in output, got:\n%s", got)
	}
}

func TestFilterGolangciLintSanityCheck(t *testing.T) {
	// Build a moderately sized input and verify output never exceeds it in tokens
	var lines []string
	linters := []string{"errcheck", "stylecheck", "ineffassign"}
	files := []string{"pkg/a/a.go", "pkg/b/b.go", "internal/c/c.go"}
	for i, linter := range linters {
		for j := 1; j <= 4; j++ {
			lines = append(lines, fmt.Sprintf("%s:%d:%d: some lint message here (%s)", files[i%len(files)], j*10, j*2, linter))
		}
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%d issues.", len(linters)*4))
	raw := strings.Join(lines, "\n")

	got, err := filterGolangciLint(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	if filteredTokens > rawTokens {
		t.Errorf("filter expanded output: raw=%d tokens, filtered=%d tokens\noutput:\n%s", rawTokens, filteredTokens, got)
	}
}
