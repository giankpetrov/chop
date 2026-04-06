package filters

import (
	"fmt"
	"strings"
	"testing"
)

func TestFilterDbQueryEmpty(t *testing.T) {
	got, err := filterDbQuery("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty passthrough, got %q", got)
	}
}

func TestFilterDbQueryPsqlFewRows(t *testing.T) {
	raw := ` id | name          | email
----+---------------+------------------
  1 | Alice Smith   | alice@example.com
  2 | Bob Johnson   | bob@example.com
(2 rows)

Time: 5.123 ms`

	got, err := filterDbQuery(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// <=20 rows — passthrough (filter trims whitespace, so compare trimmed)
	if strings.TrimSpace(got) != strings.TrimSpace(raw) {
		t.Errorf("expected passthrough for few rows, got:\n%s", got)
	}
}

func TestFilterDbQueryPsqlManyRows(t *testing.T) {
	// Build psql output with 50 rows
	var lines []string
	lines = append(lines, " id | name          | email")
	lines = append(lines, "----+---------------+------------------")
	for i := 1; i <= 50; i++ {
		lines = append(lines, fmt.Sprintf("  %d | User %d   | user%d@example.com", i, i, i))
	}
	lines = append(lines, "(50 rows)")
	lines = append(lines, "")
	lines = append(lines, "Time: 12.456 ms")
	raw := strings.Join(lines, "\n")

	got, err := filterDbQuery(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should keep header
	if !strings.Contains(got, "id") || !strings.Contains(got, "name") || !strings.Contains(got, "email") {
		t.Errorf("expected header columns in output, got:\n%s", got)
	}

	// Should keep separator
	if !strings.Contains(got, "----") {
		t.Errorf("expected separator line in output, got:\n%s", got)
	}

	// Should show first 5 rows
	if !strings.Contains(got, "User 1") {
		t.Errorf("expected first row in output, got:\n%s", got)
	}

	// Should truncate: 50 rows - 5 shown = 45 more
	if !strings.Contains(got, "45 more rows") {
		t.Errorf("expected '45 more rows' in output, got:\n%s", got)
	}

	// Should show summary "(50 rows)"
	if !strings.Contains(got, "(50 rows)") {
		t.Errorf("expected '(50 rows)' summary in output, got:\n%s", got)
	}

	// Token savings >= 50%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterDbQueryMysqlManyRows(t *testing.T) {
	// Build MySQL +---+ format with 30 rows
	header := "+----+---------------+------------------+"
	colLine := "| id | name          | email            |"
	sep := "+----+---------------+------------------+"
	var lines []string
	lines = append(lines, header)
	lines = append(lines, colLine)
	lines = append(lines, sep)
	for i := 1; i <= 30; i++ {
		lines = append(lines, fmt.Sprintf("|  %d | User %-8d | user%d@example.com |", i, i, i))
	}
	lines = append(lines, header)
	lines = append(lines, "30 rows in set (0.01 sec)")
	raw := strings.Join(lines, "\n")

	got, err := filterDbQuery(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should truncate
	if !strings.Contains(got, "more") {
		t.Errorf("expected truncation indicator in output, got:\n%s", got)
	}

	// Should show summary with row count
	if !strings.Contains(got, "30") {
		t.Errorf("expected row count in output, got:\n%s", got)
	}

	// Token savings >= 50%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterDbQuerySqliteMany(t *testing.T) {
	// SQLite pipe-separated rows, 25 rows
	var lines []string
	for i := 1; i <= 25; i++ {
		lines = append(lines, fmt.Sprintf("%d|User %d|user%d@example.com", i, i, i))
	}
	raw := strings.Join(lines, "\n")

	got, err := filterDbQuery(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should truncate: output should be shorter or equal
	if len(got) > len(raw) {
		t.Errorf("output longer than input: raw=%d bytes, filtered=%d bytes", len(raw), len(got))
	}

	// Should show truncation (sqlite has no row-count summary line)
	if !strings.Contains(got, "more") {
		t.Errorf("expected truncation indicator in output, got:\n%s", got)
	}
}

func TestFilterDbQueryExplain(t *testing.T) {
	raw := `QUERY PLAN
----------------------------------------------
 Seq Scan on users  (cost=0.00..35.50 rows=2550 width=36)
(1 row)`

	got, err := filterDbQuery(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// EXPLAIN output should always pass through
	if got != raw {
		t.Errorf("expected passthrough for EXPLAIN output, got:\n%s", got)
	}
}

func TestFilterDbQuerySanityCheck(t *testing.T) {
	// Large psql output
	var lines []string
	lines = append(lines, " id | value")
	lines = append(lines, "----+-------")
	for i := 1; i <= 100; i++ {
		lines = append(lines, fmt.Sprintf("  %d | val_%d", i, i))
	}
	lines = append(lines, "(100 rows)")
	raw := strings.Join(lines, "\n")

	got, err := filterDbQuery(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) > len(raw) {
		t.Errorf("output longer than input: raw=%d bytes, filtered=%d bytes", len(raw), len(got))
	}
}
