package tracking

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// DBPath
// ---------------------------------------------------------------------------

func TestDBPath_ReturnsConfiguredPath(t *testing.T) {
	setupTestDB(t)
	got := DBPath()
	want := os.Getenv("CHOP_DB_PATH")
	if got != want {
		t.Errorf("DBPath() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// escapeLike
// ---------------------------------------------------------------------------

func TestEscapeLike(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal", "normal"},
		{"with%percent", `with\%percent`},
		{"with_underscore", `with\_underscore`},
		{`with\backslash`, `with\\backslash`},
		{`%_\combo`, `\%\_\\combo`},
	}
	for _, tc := range tests {
		got := escapeLike(tc.input)
		if got != tc.want {
			t.Errorf("escapeLike(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// GetStatsSince
// ---------------------------------------------------------------------------

func TestGetStatsSince_ReturnsOnlyRecentRecords(t *testing.T) {
	setupTestDB(t)

	if err := Init(); err != nil {
		t.Fatal(err)
	}

	// Insert an old record (200 days ago) directly.
	old := time.Now().Local().AddDate(0, 0, -200).Format("2006-01-02 15:04:05")
	_, err := db.Exec(
		`INSERT INTO tracking (timestamp, command, raw_tokens, filtered_tokens, savings_pct) VALUES (?, ?, ?, ?, ?)`,
		old, "old cmd", 500, 100, 80.0,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Insert a recent record.
	if err := Track("recent cmd", 200, 40); err != nil {
		t.Fatal(err)
	}

	// Query for the last 30 days — should only see the recent record.
	s, err := GetStatsSince(30 * 24 * time.Hour)
	if err != nil {
		t.Fatalf("GetStatsSince failed: %v", err)
	}

	if s.TotalCommands != 1 {
		t.Errorf("expected 1 command in last 30d, got %d", s.TotalCommands)
	}
	if s.TotalRawTokens != 200 {
		t.Errorf("expected 200 raw tokens, got %d", s.TotalRawTokens)
	}
	if s.TotalSavedTokens != 160 {
		t.Errorf("expected 160 saved tokens, got %d", s.TotalSavedTokens)
	}
	if s.OverallSavingsPct < 79 || s.OverallSavingsPct > 81 {
		t.Errorf("expected ~80%% savings, got %.1f%%", s.OverallSavingsPct)
	}
}

func TestGetStatsSince_EmptyDB(t *testing.T) {
	setupTestDB(t)

	s, err := GetStatsSince(7 * 24 * time.Hour)
	if err != nil {
		t.Fatalf("GetStatsSince on empty DB failed: %v", err)
	}
	if s.TotalCommands != 0 {
		t.Errorf("expected 0 commands, got %d", s.TotalCommands)
	}
	if s.OverallSavingsPct != 0 {
		t.Errorf("expected 0%% savings, got %.1f%%", s.OverallSavingsPct)
	}
}

// ---------------------------------------------------------------------------
// GetHistorySince
// ---------------------------------------------------------------------------

func TestGetHistorySince_FiltersOldRecords(t *testing.T) {
	setupTestDB(t)

	if err := Init(); err != nil {
		t.Fatal(err)
	}

	// Old record.
	old := time.Now().Local().AddDate(0, 0, -60).Format("2006-01-02 15:04:05")
	_, err := db.Exec(
		`INSERT INTO tracking (timestamp, command, raw_tokens, filtered_tokens, savings_pct) VALUES (?, ?, ?, ?, ?)`,
		old, "ancient cmd", 100, 50, 50.0,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Recent record.
	if err := Track("fresh cmd", 100, 20); err != nil {
		t.Fatal(err)
	}

	records, err := GetHistorySince(100, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("GetHistorySince failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Command != "fresh cmd" {
		t.Errorf("expected 'fresh cmd', got %q", records[0].Command)
	}
}

func TestGetHistorySince_HonorsLimit(t *testing.T) {
	setupTestDB(t)

	for i := 0; i < 5; i++ {
		if err := Track("cmd", 100, 50); err != nil {
			t.Fatal(err)
		}
	}

	records, err := GetHistorySince(3, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records (limit), got %d", len(records))
	}
}

// ---------------------------------------------------------------------------
// FormatGainSince
// ---------------------------------------------------------------------------

func TestFormatGainSince_ContainsKeyFields(t *testing.T) {
	s := Stats{
		TotalCommands:     42,
		TotalSavedTokens:  1234,
		OverallSavingsPct: 61.7,
	}
	out := FormatGainSince(s, "7d")

	if !strings.Contains(out, "7d") {
		t.Errorf("missing duration in output: %s", out)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("missing command count in output: %s", out)
	}
	if !strings.Contains(out, "1,234") {
		t.Errorf("missing formatted token count in output: %s", out)
	}
	if !strings.Contains(out, "61.7") {
		t.Errorf("missing savings pct in output: %s", out)
	}
}

// ---------------------------------------------------------------------------
// GetProjectSummary / FormatProjectSummary
// ---------------------------------------------------------------------------

func TestGetProjectSummary_GroupsByProject(t *testing.T) {
	setupTestDB(t)

	if err := Init(); err != nil {
		t.Fatal(err)
	}

	// Insert records with known projects.
	insert := func(ts, cmd string, raw, filtered int, pct float64, proj string) {
		t.Helper()
		_, err := db.Exec(
			`INSERT INTO tracking (timestamp, command, raw_tokens, filtered_tokens, savings_pct, project) VALUES (?, ?, ?, ?, ?, ?)`,
			ts, cmd, raw, filtered, pct, proj,
		)
		if err != nil {
			t.Fatalf("insert failed: %v", err)
		}
	}

	insert("2026-03-01 10:00:00", "git status", 200, 40, 80.0, "/proj/alpha")
	insert("2026-03-01 10:01:00", "git log", 300, 60, 80.0, "/proj/alpha")
	insert("2026-03-01 10:02:00", "npm test", 100, 20, 80.0, "/proj/beta")

	summaries, err := GetProjectSummary()
	if err != nil {
		t.Fatalf("GetProjectSummary failed: %v", err)
	}

	if len(summaries) != 2 {
		t.Fatalf("expected 2 project summaries, got %d", len(summaries))
	}

	// /proj/alpha has 400 raw, 340 saved — should be first.
	if summaries[0].Project != "/proj/alpha" {
		t.Errorf("expected /proj/alpha first, got %q", summaries[0].Project)
	}
	if summaries[0].Count != 2 {
		t.Errorf("expected count 2 for alpha, got %d", summaries[0].Count)
	}
	if summaries[0].SavedTokens != 400 {
		t.Errorf("expected 400 saved for alpha, got %d", summaries[0].SavedTokens)
	}
}

func TestGetProjectSummary_Empty(t *testing.T) {
	setupTestDB(t)

	summaries, err := GetProjectSummary()
	if err != nil {
		t.Fatalf("GetProjectSummary failed: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries on empty DB, got %d", len(summaries))
	}
}

func TestFormatProjectSummary_ContainsProjects(t *testing.T) {
	summaries := []ProjectSummary{
		{Project: "/home/user/dev/alpha", Count: 10, RawTokens: 1000, SavedTokens: 800, SavingsPct: 80.0},
		{Project: "/home/user/dev/beta", Count: 3, RawTokens: 300, SavedTokens: 150, SavingsPct: 50.0},
	}
	out := FormatProjectSummary(summaries)

	if !strings.Contains(out, "alpha") {
		t.Errorf("missing alpha project in output: %s", out)
	}
	if !strings.Contains(out, "beta") {
		t.Errorf("missing beta project in output: %s", out)
	}
	if !strings.Contains(out, "by Project") {
		t.Errorf("missing header in output: %s", out)
	}
}

func TestFormatProjectSummary_Empty(t *testing.T) {
	out := FormatProjectSummary(nil)
	if out != "no projects tracked yet" {
		t.Errorf("unexpected empty message: %q", out)
	}
}

func TestFormatProjectSummary_LongProjectNameTruncated(t *testing.T) {
	longName := "/home/user/very/deeply/nested/project/path/that/exceeds/forty/chars"
	summaries := []ProjectSummary{
		{Project: longName, Count: 1, RawTokens: 100, SavedTokens: 50, SavingsPct: 50.0},
	}
	out := FormatProjectSummary(summaries)
	// The output should contain "..." as a truncation marker.
	if !strings.Contains(out, "...") {
		t.Errorf("expected long project name to be truncated with '...': %s", out)
	}
}

func TestFormatProjectSummary_UnknownProject(t *testing.T) {
	summaries := []ProjectSummary{
		{Project: "", Count: 2, RawTokens: 200, SavedTokens: 100, SavingsPct: 50.0},
	}
	out := FormatProjectSummary(summaries)
	if !strings.Contains(out, "(unknown)") {
		t.Errorf("expected '(unknown)' for empty project: %s", out)
	}
}

// ---------------------------------------------------------------------------
// FormatHistory color mode
// ---------------------------------------------------------------------------

func TestFormatHistory_ColorMode(t *testing.T) {
	records := []Record{
		{Timestamp: "2026-03-05 14:23:00", Command: "git status", RawTokens: 100, FilteredTokens: 10, SavingsPct: 90.0, Project: "/proj/a"},
		{Timestamp: "2026-03-05 14:22:00", Command: "git diff", RawTokens: 50, FilteredTokens: 40, SavingsPct: 20.0, Project: "/proj/a"},
		{Timestamp: "2026-03-05 14:21:00", Command: "ls", RawTokens: 10, FilteredTokens: 10, SavingsPct: 0.0, Project: "/proj/a"},
	}
	out := FormatHistory(records, false, true)
	// ANSI color codes should be present.
	if !strings.Contains(out, "\033[") {
		t.Errorf("expected ANSI color codes in color=true output: %s", out)
	}
	// All commands should appear.
	if !strings.Contains(out, "git status") {
		t.Errorf("missing 'git status' in color output: %s", out)
	}
	if !strings.Contains(out, "git diff") {
		t.Errorf("missing 'git diff' in color output: %s", out)
	}
}

func TestFormatHistory_ColorMode_ZeroSavingsMarker(t *testing.T) {
	records := []Record{
		{Timestamp: "2026-03-05 14:21:00", Command: "ls", RawTokens: 10, FilteredTokens: 10, SavingsPct: 0.0, Project: "/proj/a"},
	}
	out := FormatHistory(records, false, true)
	// The zero-savings row should have the "!" marker.
	if !strings.Contains(out, "!") {
		t.Errorf("expected '!' marker for 0%% savings in color output: %s", out)
	}
}

func TestFormatHistory_VerboseColorProjectHeaders(t *testing.T) {
	records := []Record{
		{Timestamp: "2026-03-05 14:23:00", Command: "git push", RawTokens: 100, FilteredTokens: 80, SavingsPct: 20.0, Project: "/proj/a"},
		{Timestamp: "2026-03-05 14:22:00", Command: "npm test", RawTokens: 200, FilteredTokens: 100, SavingsPct: 50.0, Project: "/proj/b"},
	}
	out := FormatHistory(records, true, true)
	if !strings.Contains(out, "[project: /proj/a]") {
		t.Errorf("missing project header in verbose+color output: %s", out)
	}
	if !strings.Contains(out, "[project: /proj/b]") {
		t.Errorf("missing second project header in verbose+color output: %s", out)
	}
}

func TestFormatHistory_CommandTruncation(t *testing.T) {
	longCmd := strings.Repeat("x", 80)
	records := []Record{
		{Timestamp: "2026-03-05 14:23:00", Command: longCmd, RawTokens: 100, FilteredTokens: 20, SavingsPct: 80.0},
	}
	out := FormatHistory(records, false, false)
	// Non-verbose: long commands should be truncated with "..."
	if !strings.Contains(out, "...") {
		t.Errorf("expected long command to be truncated with '...': %s", out)
	}
}

func TestFormatHistory_VerboseNoTruncation(t *testing.T) {
	longCmd := strings.Repeat("x", 80)
	records := []Record{
		{Timestamp: "2026-03-05 14:23:00", Command: longCmd, RawTokens: 100, FilteredTokens: 20, SavingsPct: 80.0},
	}
	out := FormatHistory(records, true, false)
	// Verbose: full command should appear.
	if !strings.Contains(out, longCmd) {
		t.Errorf("expected full long command in verbose output")
	}
}

// ---------------------------------------------------------------------------
// FormatUnchopped with filtered section
// ---------------------------------------------------------------------------

func TestFormatUnchopped_WithFilteredSection(t *testing.T) {
	summaries := []UnchoppedSummary{
		{Command: "make test", Count: 5, TotalTokens: 500},
	}
	filtered := []UnchoppedSummary{
		{Command: "git status", Count: 10, TotalTokens: 200},
	}
	out := FormatUnchopped(summaries, nil, filtered, false)

	if !strings.Contains(out, "filter registered") {
		t.Errorf("missing filtered section header: %s", out)
	}
	if !strings.Contains(out, "git status") {
		t.Errorf("missing filtered command: %s", out)
	}
	if !strings.Contains(out, "make test") {
		t.Errorf("missing active candidate: %s", out)
	}
}

func TestFormatUnchopped_VerboseLongCommand(t *testing.T) {
	longCmd := strings.Repeat("z", 30)
	summaries := []UnchoppedSummary{
		{Command: longCmd, Count: 1, TotalTokens: 100},
	}
	// verbose=false: should truncate
	outTruncated := FormatUnchopped(summaries, nil, nil, false)
	if !strings.Contains(outTruncated, "...") {
		t.Errorf("expected truncation in non-verbose mode: %s", outTruncated)
	}
	// verbose=true: full name should appear
	outFull := FormatUnchopped(summaries, nil, nil, true)
	if !strings.Contains(outFull, longCmd) {
		t.Errorf("expected full command name in verbose mode: %s", outFull)
	}
}

// ---------------------------------------------------------------------------
// Cleanup edge cases
// ---------------------------------------------------------------------------

func TestCleanup_ZeroDays_RemovesOldRecords(t *testing.T) {
	setupTestDB(t)
	if err := Init(); err != nil {
		t.Fatal(err)
	}

	// Insert records with a timestamp clearly in the past (yesterday).
	yesterday := time.Now().Local().AddDate(0, 0, -1).Format("2006-01-02 15:04:05")
	for _, cmd := range []string{"cmd a", "cmd b"} {
		_, err := db.Exec(
			`INSERT INTO tracking (timestamp, command, raw_tokens, filtered_tokens, savings_pct) VALUES (?, ?, ?, ?, ?)`,
			yesterday, cmd, 100, 50, 50.0,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Cleanup(0): cutoff = now — records from yesterday are older than cutoff.
	if err := Cleanup(0); err != nil {
		t.Fatalf("Cleanup(0) failed: %v", err)
	}

	records, err := GetHistory(100)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records after Cleanup(0), got %d", len(records))
	}
}

func TestCleanup_LargeDays_KeepsAll(t *testing.T) {
	setupTestDB(t)

	if err := Track("cmd a", 100, 50); err != nil {
		t.Fatal(err)
	}
	if err := Track("cmd b", 200, 100); err != nil {
		t.Fatal(err)
	}

	// Cleanup with a very large window should not remove anything recent.
	if err := Cleanup(36500); err != nil { // 100 years
		t.Fatalf("Cleanup(36500) failed: %v", err)
	}

	records, err := GetHistory(100)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
}

// ---------------------------------------------------------------------------
// ExportJSON / ExportCSV edge cases
// ---------------------------------------------------------------------------

func TestExportJSON_EmptyRecords(t *testing.T) {
	var buf bytes.Buffer
	s := Stats{TotalCommands: 0}
	if err := ExportJSON(&buf, nil, s); err != nil {
		t.Fatalf("ExportJSON with nil records failed: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	history, ok := out["history"].([]interface{})
	if !ok || len(history) != 0 {
		// json.Encoder may encode nil slice as null; both are acceptable.
		if out["history"] != nil {
			t.Errorf("expected empty/null history, got %v", out["history"])
		}
	}
}

func TestExportCSV_EmptyRecords(t *testing.T) {
	var buf bytes.Buffer
	if err := ExportCSV(&buf, nil); err != nil {
		t.Fatalf("ExportCSV with nil records failed: %v", err)
	}
	r := csv.NewReader(&buf)
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV: %v", err)
	}
	// Only the header row should be present.
	if len(rows) != 1 {
		t.Errorf("expected 1 row (header only), got %d", len(rows))
	}
}

func TestExportCSV_SavedTokensCalculation(t *testing.T) {
	records := []Record{
		{Timestamp: "2026-03-01 10:00:00", Command: "git log", RawTokens: 300, FilteredTokens: 100, SavingsPct: 66.7},
	}
	var buf bytes.Buffer
	if err := ExportCSV(&buf, records); err != nil {
		t.Fatal(err)
	}
	r := csv.NewReader(&buf)
	rows, _ := r.ReadAll()
	// rows[1] = data row; col index 4 = saved_tokens
	if rows[1][4] != "200" {
		t.Errorf("expected saved_tokens=200, got %q", rows[1][4])
	}
}

// ---------------------------------------------------------------------------
// ParseSinceDuration additional cases
// ---------------------------------------------------------------------------

func TestParseSinceDuration_StandardGoFormat(t *testing.T) {
	// "30s" is a valid Go duration string (not handled by the custom switch).
	d, err := ParseSinceDuration("30s")
	if err != nil {
		t.Fatalf("expected no error for '30s', got %v", err)
	}
	if d != 30*time.Second {
		t.Errorf("expected 30s, got %v", d)
	}
}

func TestParseSinceDuration_InvalidUnit(t *testing.T) {
	// "5z" is not a valid Go duration or custom unit.
	_, err := ParseSinceDuration("5z")
	if err == nil {
		t.Error("expected error for '5z', got nil")
	}
}

// ---------------------------------------------------------------------------
// IsColorEnabled
// ---------------------------------------------------------------------------

func TestIsColorEnabled_NoColorEnvDisablesColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if IsColorEnabled() {
		t.Error("IsColorEnabled() should return false when NO_COLOR is set")
	}
}

func TestIsColorEnabled_EmptyNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	// Can't guarantee a terminal in CI, but the function should not panic.
	_ = IsColorEnabled()
}

// ---------------------------------------------------------------------------
// MigrateWindowsDataDir (Windows-only build; runs on Windows inside Docker
// only if GOOS=windows — tested here to exercise the no-op path on Linux)
// ---------------------------------------------------------------------------

func TestMigrateWindowsDataDir_NoOp(t *testing.T) {
	// On non-Windows (Linux Docker) this is a no-op. On Windows it may do real
	// work; to keep things safe we call it with a CHOP_DB_PATH pointing at a
	// fresh temp dir so it cannot touch real user data.
	setupTestDB(t)
	// Should not panic regardless of platform.
	MigrateWindowsDataDir()
}

// ---------------------------------------------------------------------------
// GetHistoryByProject limit
// ---------------------------------------------------------------------------

func TestGetHistoryByProject_HonorsLimit(t *testing.T) {
	setupTestDB(t)
	if err := Init(); err != nil {
		t.Fatal(err)
	}

	proj := "/test/project"
	for i := 0; i < 5; i++ {
		_, err := db.Exec(
			`INSERT INTO tracking (timestamp, command, raw_tokens, filtered_tokens, savings_pct, project) VALUES (?, ?, ?, ?, ?, ?)`,
			"2026-03-01 10:00:00", "git status", 100, 20, 80.0, proj,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	records, err := GetHistoryByProject(proj, 3)
	if err != nil {
		t.Fatalf("GetHistoryByProject failed: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records (limit), got %d", len(records))
	}
}

func TestGetHistoryByProject_EmptyForUnknownProject(t *testing.T) {
	setupTestDB(t)

	if err := Track("git status", 100, 20); err != nil {
		t.Fatal(err)
	}

	records, err := GetHistoryByProject("/does/not/exist", 10)
	if err != nil {
		t.Fatalf("GetHistoryByProject failed: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records for unknown project, got %d", len(records))
	}
}

// ---------------------------------------------------------------------------
// GetSkippedCommands
// ---------------------------------------------------------------------------

func TestGetSkippedCommands_AlphabeticOrder(t *testing.T) {
	setupTestDB(t)

	cmds := []string{"zzz", "aaa", "mmm"}
	for _, c := range cmds {
		if err := SkipUnchopped(c); err != nil {
			t.Fatalf("SkipUnchopped(%q) failed: %v", c, err)
		}
	}

	skipped, err := GetSkippedCommands()
	if err != nil {
		t.Fatalf("GetSkippedCommands failed: %v", err)
	}
	if len(skipped) != 3 {
		t.Fatalf("expected 3 skipped, got %d", len(skipped))
	}
	if skipped[0] != "aaa" || skipped[1] != "mmm" || skipped[2] != "zzz" {
		t.Errorf("expected alphabetic order, got %v", skipped)
	}
}

func TestGetSkippedCommands_EmptyWhenNoneSkipped(t *testing.T) {
	setupTestDB(t)

	skipped, err := GetSkippedCommands()
	if err != nil {
		t.Fatalf("GetSkippedCommands failed: %v", err)
	}
	if len(skipped) != 0 {
		t.Errorf("expected empty skip list, got %v", skipped)
	}
}

// ---------------------------------------------------------------------------
// Track — skip list behaviour
// ---------------------------------------------------------------------------

func TestTrack_SkippedCommandNotRecorded(t *testing.T) {
	setupTestDB(t)

	if err := AddTrackingSkip("secret cmd"); err != nil {
		t.Fatal(err)
	}

	if err := Track("secret cmd", 100, 50); err != nil {
		t.Fatal(err)
	}

	records, err := GetHistory(100)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range records {
		if r.Command == "secret cmd" {
			t.Errorf("skipped command 'secret cmd' should not appear in history")
		}
	}
}

func TestTrack_SkipPrefixMatch(t *testing.T) {
	setupTestDB(t)

	// Skip "docker" — any command starting with "docker " should also be skipped.
	if err := AddTrackingSkip("docker"); err != nil {
		t.Fatal(err)
	}

	if err := Track("docker ps", 100, 50); err != nil {
		t.Fatal(err)
	}

	records, err := GetHistory(100)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range records {
		if strings.HasPrefix(r.Command, "docker") {
			t.Errorf("docker command should have been skipped, got: %q", r.Command)
		}
	}
}

// ---------------------------------------------------------------------------
// formatNum (internal) — already tested, but verify negative large values.
// ---------------------------------------------------------------------------

func TestFormatNum_NegativeLarge(t *testing.T) {
	got := formatNum(-1234567)
	want := "-1,234,567"
	if got != want {
		t.Errorf("formatNum(-1234567) = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// MigrateWindowsDataDir (Windows-specific logic tested on Windows only)
// ---------------------------------------------------------------------------

func TestMigrateWindowsDataDir_AlreadyMigrated(t *testing.T) {
	setupTestDB(t)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	legacyDir := filepath.Join(tmp, ".local", "share", "chop")
	if err := os.MkdirAll(legacyDir, 0o700); err != nil {
		t.Fatal(err)
	}

	// Create a legacy DB so it looks like there's something to migrate.
	if err := os.WriteFile(filepath.Join(legacyDir, "tracking.db"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Create the sentinel to simulate already-migrated state.
	sentinel := filepath.Join(legacyDir, ".migrated")
	if err := os.WriteFile(sentinel, []byte{}, 0o600); err != nil {
		t.Fatal(err)
	}

	// Should return early without doing anything.
	MigrateWindowsDataDir()

	// Sentinel should still exist.
	if _, err := os.Stat(sentinel); os.IsNotExist(err) {
		t.Error("sentinel should still exist after already-migrated early return")
	}
}
