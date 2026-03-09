package tracking

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	initForTest()
	dir := t.TempDir()
	dbFile := filepath.Join(dir, "test_tracking.db")
	t.Setenv("CHOP_DB_PATH", dbFile)
}

func TestInitCreatesDBAndTable(t *testing.T) {
	setupTestDB(t)

	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	path := os.Getenv("CHOP_DB_PATH")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}
}

func TestTrackAndGetStats(t *testing.T) {
	setupTestDB(t)

	// Track 5 commands with varying savings
	commands := []struct {
		cmd      string
		raw, fil int
	}{
		{"git status", 100, 20},
		{"git log", 200, 40},
		{"docker ps", 80, 30},
		{"npm list", 150, 45},
		{"git diff", 300, 60},
	}

	for _, c := range commands {
		if err := Track(c.cmd, c.raw, c.fil); err != nil {
			t.Fatalf("Track(%s) failed: %v", c.cmd, err)
		}
	}

	stats, err := GetStats()
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.TotalCommands != 5 {
		t.Errorf("expected 5 commands, got %d", stats.TotalCommands)
	}

	expectedRaw := 100 + 200 + 80 + 150 + 300
	if stats.TotalRawTokens != expectedRaw {
		t.Errorf("expected %d raw tokens, got %d", expectedRaw, stats.TotalRawTokens)
	}

	expectedSaved := (100 - 20) + (200 - 40) + (80 - 30) + (150 - 45) + (300 - 60)
	if stats.TotalSavedTokens != expectedSaved {
		t.Errorf("expected %d saved tokens, got %d", expectedSaved, stats.TotalSavedTokens)
	}

	expectedPct := float64(expectedSaved) / float64(expectedRaw) * 100.0
	if diff := stats.OverallSavingsPct - expectedPct; diff > 0.1 || diff < -0.1 {
		t.Errorf("expected %.1f%% savings, got %.1f%%", expectedPct, stats.OverallSavingsPct)
	}

	// Today's stats should match total (all tracked today)
	if stats.TodayCommands != 5 {
		t.Errorf("expected 5 today commands, got %d", stats.TodayCommands)
	}
	if stats.TodaySavedTokens != expectedSaved {
		t.Errorf("expected %d today saved, got %d", expectedSaved, stats.TodaySavedTokens)
	}
}

func TestGetHistoryReverseChronological(t *testing.T) {
	setupTestDB(t)

	commands := []string{"git status", "git log", "docker ps"}
	for _, cmd := range commands {
		if err := Track(cmd, 100, 50); err != nil {
			t.Fatalf("Track(%s) failed: %v", cmd, err)
		}
	}

	records, err := GetHistory(10)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}

	// Most recent first
	if records[0].Command != "docker ps" {
		t.Errorf("expected first record 'docker ps', got %q", records[0].Command)
	}
	if records[2].Command != "git status" {
		t.Errorf("expected last record 'git status', got %q", records[2].Command)
	}
}

func TestGetHistoryLimit(t *testing.T) {
	setupTestDB(t)

	for i := 0; i < 10; i++ {
		if err := Track("cmd", 100, 50); err != nil {
			t.Fatal(err)
		}
	}

	records, err := GetHistory(3)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
}

func TestCleanupRemovesOldRecords(t *testing.T) {
	setupTestDB(t)

	if err := Init(); err != nil {
		t.Fatal(err)
	}

	// Insert an old record directly
	old := time.Now().UTC().AddDate(0, 0, -100).Format("2006-01-02 15:04:05")
	_, err := db.Exec(
		`INSERT INTO tracking (timestamp, command, raw_tokens, filtered_tokens, savings_pct) VALUES (?, ?, ?, ?, ?)`,
		old, "old cmd", 100, 50, 50.0,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Insert a recent record
	if err := Track("new cmd", 100, 50); err != nil {
		t.Fatal(err)
	}

	if err := Cleanup(90); err != nil {
		t.Fatal(err)
	}

	records, err := GetHistory(100)
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Errorf("expected 1 record after cleanup, got %d", len(records))
	}
	if records[0].Command != "new cmd" {
		t.Errorf("expected 'new cmd', got %q", records[0].Command)
	}
}

func TestTrackSilentOnZeroRaw(t *testing.T) {
	setupTestDB(t)

	// Should not panic or error on zero raw tokens
	if err := Track("empty", 0, 0); err != nil {
		t.Fatalf("Track with zero tokens failed: %v", err)
	}

	stats, err := GetStats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalCommands != 1 {
		t.Errorf("expected 1 command, got %d", stats.TotalCommands)
	}
}

func TestCountTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"hello world", 2},
		{"  spaces  everywhere  ", 2},
		{"one", 1},
		{"", 0},
		{"multi\nline\ntext", 3},
	}
	for _, tc := range tests {
		got := CountTokens(tc.input)
		if got != tc.expected {
			t.Errorf("CountTokens(%q) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}

func TestFormatGain(t *testing.T) {
	s := Stats{
		TotalCommands:     1203,
		TotalRawTokens:    500000,
		TotalSavedTokens:  456789,
		OverallSavingsPct: 78.3,
		TodayCommands:     45,
		TodaySavedTokens:  12340,
		WeekCommands:      210,
		WeekSavedTokens:   98765,
		MonthCommands:     890,
		MonthSavedTokens:  345678,
		YearCommands:      1150,
		YearSavedTokens:   430000,
	}
	out := FormatGain(s)
	if !strings.Contains(out, "45 commands") {
		t.Errorf("missing today commands in output: %s", out)
	}
	if !strings.Contains(out, "12,340") {
		t.Errorf("missing formatted today saved in output: %s", out)
	}
	if !strings.Contains(out, "210 commands") {
		t.Errorf("missing week commands in output: %s", out)
	}
	if !strings.Contains(out, "890 commands") {
		t.Errorf("missing month commands in output: %s", out)
	}
	if !strings.Contains(out, "1150 commands") {
		t.Errorf("missing year commands in output: %s", out)
	}
	if !strings.Contains(out, "78.3%") {
		t.Errorf("missing overall pct in output: %s", out)
	}
}

func TestFormatHistory(t *testing.T) {
	records := []Record{
		{Timestamp: "2026-03-05 14:23:00", Command: "git status", RawTokens: 67, FilteredTokens: 8, SavingsPct: 88.1},
	}
	out := FormatHistory(records)
	if !strings.Contains(out, "git status") {
		t.Errorf("missing command in history: %s", out)
	}
	if !strings.Contains(out, "88.1%") {
		t.Errorf("missing savings in history: %s", out)
	}
}

func TestFormatHistoryEmpty(t *testing.T) {
	out := FormatHistory(nil)
	if out != "no commands tracked yet" {
		t.Errorf("unexpected empty history: %s", out)
	}
}
