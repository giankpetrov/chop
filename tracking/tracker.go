package tracking

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Stats holds aggregate token savings statistics.
type Stats struct {
	TotalCommands     int
	TotalRawTokens    int
	TotalSavedTokens  int
	OverallSavingsPct float64
	TodayCommands     int
	TodaySavedTokens  int
	WeekCommands      int
	WeekSavedTokens   int
	MonthCommands     int
	MonthSavedTokens  int
	YearCommands      int
	YearSavedTokens   int
}

// Record holds a single tracking entry.
type Record struct {
	Timestamp      string
	Command        string
	RawTokens      int
	FilteredTokens int
	SavingsPct     float64
}

var (
	db     *sql.DB
	dbOnce sync.Once
	dbErr  error
)

func dbPath() string {
	if p := os.Getenv("CHOP_DB_PATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".local", "share", "chop", "tracking.db")
}

// Init opens (or creates) the tracking database and ensures the schema exists.
func Init() error {
	dbOnce.Do(func() {
		path := dbPath()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			dbErr = err
			return
		}
		db, dbErr = sql.Open("sqlite", path)
		if dbErr != nil {
			return
		}
		_, dbErr = db.Exec(`CREATE TABLE IF NOT EXISTS tracking (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			command TEXT NOT NULL,
			raw_tokens INTEGER NOT NULL,
			filtered_tokens INTEGER NOT NULL,
			savings_pct REAL NOT NULL
		)`)
	})
	return dbErr
}

// initForTest resets the singleton so tests can re-init with a new DB path.
func initForTest() {
	dbOnce = sync.Once{}
	db = nil
	dbErr = nil
}

// Track records a command's token savings. Silent on error.
func Track(command string, rawTokens, filteredTokens int) error {
	if err := Init(); err != nil {
		return err
	}
	var savingsPct float64
	if rawTokens > 0 {
		savingsPct = 100.0 - (float64(filteredTokens) / float64(rawTokens) * 100.0)
	}
	now := time.Now().Local().Format("2006-01-02 15:04:05")
	_, err := db.Exec(
		`INSERT INTO tracking (timestamp, command, raw_tokens, filtered_tokens, savings_pct)
		 VALUES (?, ?, ?, ?, ?)`,
		now, command, rawTokens, filteredTokens, savingsPct,
	)
	return err
}

// GetStats returns aggregate statistics.
func GetStats() (Stats, error) {
	if err := Init(); err != nil {
		return Stats{}, err
	}
	var s Stats

	row := db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(raw_tokens),0), COALESCE(SUM(raw_tokens - filtered_tokens),0) FROM tracking`)
	if err := row.Scan(&s.TotalCommands, &s.TotalRawTokens, &s.TotalSavedTokens); err != nil {
		return Stats{}, err
	}
	if s.TotalRawTokens > 0 {
		s.OverallSavingsPct = float64(s.TotalSavedTokens) / float64(s.TotalRawTokens) * 100.0
	}

	today := time.Now().Local().Format("2006-01-02")
	row = db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(raw_tokens - filtered_tokens),0) FROM tracking WHERE timestamp LIKE ?`,
		today+"%",
	)
	if err := row.Scan(&s.TodayCommands, &s.TodaySavedTokens); err != nil {
		return Stats{}, err
	}

	// Calendar week: Monday 00:00 through now
	now := time.Now().Local()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-int(weekday-time.Monday), 0, 0, 0, 0, now.Location()).Format("2006-01-02 00:00:00")
	row = db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(raw_tokens - filtered_tokens),0) FROM tracking WHERE timestamp >= ?`,
		weekStart,
	)
	if err := row.Scan(&s.WeekCommands, &s.WeekSavedTokens); err != nil {
		return Stats{}, err
	}

	// Calendar month: 1st of current month through now
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02 00:00:00")
	row = db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(raw_tokens - filtered_tokens),0) FROM tracking WHERE timestamp >= ?`,
		monthStart,
	)
	if err := row.Scan(&s.MonthCommands, &s.MonthSavedTokens); err != nil {
		return Stats{}, err
	}

	// Calendar year: Jan 1 of current year through now
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02 00:00:00")
	row = db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(raw_tokens - filtered_tokens),0) FROM tracking WHERE timestamp >= ?`,
		yearStart,
	)
	if err := row.Scan(&s.YearCommands, &s.YearSavedTokens); err != nil {
		return Stats{}, err
	}

	return s, nil
}

// GetHistory returns the last N tracking records in reverse chronological order.
func GetHistory(limit int) ([]Record, error) {
	if err := Init(); err != nil {
		return nil, err
	}
	rows, err := db.Query(
		`SELECT timestamp, command, raw_tokens, filtered_tokens, savings_pct
		 FROM tracking ORDER BY id DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.Timestamp, &r.Command, &r.RawTokens, &r.FilteredTokens, &r.SavingsPct); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// CommandSummary holds per-command aggregate stats.
type CommandSummary struct {
	BaseCommand string
	Count       int
	RawTokens   int
	SavedTokens int
	SavingsPct  float64
	ZeroCount   int // times with 0% savings
}

// GetCommandSummary returns per-base-command aggregates, sorted by tokens saved descending.
func GetCommandSummary() ([]CommandSummary, error) {
	if err := Init(); err != nil {
		return nil, err
	}
	rows, err := db.Query(`
		SELECT
			CASE
				WHEN INSTR(command, ' ') > 0 THEN SUBSTR(command, 1, INSTR(command, ' ') - 1)
				ELSE command
			END AS base_cmd,
			COUNT(*) AS cnt,
			COALESCE(SUM(raw_tokens), 0) AS raw,
			COALESCE(SUM(raw_tokens - filtered_tokens), 0) AS saved,
			SUM(CASE WHEN savings_pct = 0 AND raw_tokens > 0 THEN 1 ELSE 0 END) AS zero_cnt
		FROM tracking
		GROUP BY base_cmd
		ORDER BY saved DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []CommandSummary
	for rows.Next() {
		var s CommandSummary
		if err := rows.Scan(&s.BaseCommand, &s.Count, &s.RawTokens, &s.SavedTokens, &s.ZeroCount); err != nil {
			return nil, err
		}
		if s.RawTokens > 0 {
			s.SavingsPct = float64(s.SavedTokens) / float64(s.RawTokens) * 100.0
		}
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}

// Cleanup removes records older than the given number of days.
func Cleanup(days int) error {
	if err := Init(); err != nil {
		return err
	}
	cutoff := time.Now().Local().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	_, err := db.Exec(`DELETE FROM tracking WHERE timestamp < ?`, cutoff)
	return err
}

// CountTokens returns the word count of a string (whitespace-split).
func CountTokens(s string) int {
	return len(strings.Fields(s))
}

// FormatGain prints the gain summary report.
func FormatGain(s Stats) string {
	return fmt.Sprintf(`chop - token savings report

  today: %d commands, %s tokens saved
  week:  %d commands, %s tokens saved
  month: %d commands, %s tokens saved
  year:  %d commands, %s tokens saved
  total: %d commands, %s tokens saved (%.1f%% avg)

run 'chop gain --history' for command history`,
		s.TodayCommands, formatNum(s.TodaySavedTokens),
		s.WeekCommands, formatNum(s.WeekSavedTokens),
		s.MonthCommands, formatNum(s.MonthSavedTokens),
		s.YearCommands, formatNum(s.YearSavedTokens),
		s.TotalCommands, formatNum(s.TotalSavedTokens), s.OverallSavingsPct,
	)
}

// FormatHistory formats history records for display.
func FormatHistory(records []Record) string {
	if len(records) == 0 {
		return "no commands tracked yet"
	}
	var b strings.Builder
	b.WriteString("recent commands:\n")
	for _, r := range records {
		marker := " "
		if r.SavingsPct == 0 && r.RawTokens > 0 {
			marker = "!"
		}
		b.WriteString(fmt.Sprintf(" %s %s  %-25s %5.1f%%  (%d -> %d tokens)\n",
			marker, r.Timestamp, r.Command, r.SavingsPct, r.RawTokens, r.FilteredTokens))
	}
	b.WriteString("\n ! = 0% savings (filter may need improvement)\n")
	return b.String()
}

// FormatSummary formats per-command aggregates.
func FormatSummary(summaries []CommandSummary) string {
	if len(summaries) == 0 {
		return "no commands tracked yet"
	}
	var b strings.Builder
	b.WriteString("per-command savings:\n")
	b.WriteString(fmt.Sprintf("  %-12s %5s %8s %7s %s\n", "COMMAND", "CALLS", "SAVED", "AVG", ""))
	for _, s := range summaries {
		warn := ""
		if s.ZeroCount > 0 {
			warn = fmt.Sprintf("(%d calls at 0%%)", s.ZeroCount)
		}
		b.WriteString(fmt.Sprintf("  %-12s %5d %8s %6.0f%%  %s\n",
			s.BaseCommand, s.Count, formatNum(s.SavedTokens), s.SavingsPct, warn))
	}
	return b.String()
}

func formatNum(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d,%03d", n/1000, n%1000)
}
