package filters

import (
	"fmt"
	"strings"
)

const dbPassthroughRows = 20
const dbPreviewRows = 5

func filterDbQuery(raw string) (string, error) {
	raw = stripAnsi(strings.TrimSpace(raw))
	if raw == "" {
		return raw, nil
	}
	if !looksLikeDbOutput(raw) {
		return raw, nil
	}

	// Passthrough EXPLAIN and describe (\d) — structural, always short enough
	firstLine := strings.ToUpper(strings.TrimSpace(strings.SplitN(raw, "\n", 2)[0]))
	if strings.HasPrefix(firstLine, "EXPLAIN") || strings.HasPrefix(firstLine, "\\D") {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")

	rows, header, separator, summary, format := parseDbOutput(lines)

	if format == "unknown" {
		return raw, nil
	}

	if len(rows) <= dbPassthroughRows {
		return raw, nil
	}

	var out []string
	if header != "" {
		out = append(out, header)
	}
	if separator != "" {
		out = append(out, separator)
	}
	for i := 0; i < dbPreviewRows && i < len(rows); i++ {
		out = append(out, rows[i])
	}
	remaining := len(rows) - dbPreviewRows
	out = append(out, fmt.Sprintf("... and %d more rows", remaining))
	if summary != "" {
		out = append(out, summary)
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

// parseDbOutput identifies psql / mysql / sqlite3 format and extracts components.
func parseDbOutput(lines []string) (rows []string, header, separator, summary, format string) {
	format = "unknown"
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		// Detect MySQL: separator lines start with +
		if strings.HasPrefix(t, "+") && strings.Contains(t, "-") {
			if format == "unknown" {
				format = "mysql"
			}
			if separator == "" {
				separator = line
			}
			continue
		}
		if format == "mysql" {
			if strings.HasPrefix(t, "|") {
				if header == "" {
					header = line
					continue
				}
				rows = append(rows, line)
				continue
			}
			// Summary lines like "30 rows in set (0.01 sec)" don't start with |
			if isMySQLSummary(t) {
				summary = t
				continue
			}
			continue
		}

		// Detect psql: separator line contains only dashes, plusses, spaces.
		// This overrides sqlite if we previously misclassified a header line.
		if isPsqlSeparator(t) {
			format = "psql"
			if separator == "" {
				separator = line
			}
			continue
		}
		if format == "psql" {
			if isPsqlSummary(t) {
				summary = t
				continue
			}
			if strings.HasPrefix(t, "Time:") {
				continue
			}
			rows = append(rows, line)
			continue
		}

		// Detect SQLite: pipe-separated values, no decorators
		if strings.Contains(t, "|") && !strings.HasPrefix(t, "|") {
			if format == "unknown" {
				format = "sqlite"
			}
			if isSqliteSummary(t) {
				summary = t
				continue
			}
			rows = append(rows, line)
			continue
		}
	}

	// For psql, the header is the line before the separator.
	if format == "psql" {
		rows = nil
		sepSeen := false
		for _, line := range lines {
			t := strings.TrimSpace(line)
			if t == "" {
				continue
			}
			if isPsqlSeparator(t) {
				sepSeen = true
				separator = line
				continue
			}
			if !sepSeen {
				header = line
				continue
			}
			if isPsqlSummary(t) {
				summary = t
				continue
			}
			if strings.HasPrefix(t, "Time:") {
				continue
			}
			rows = append(rows, line)
		}
	}

	return rows, header, separator, summary, format
}

func isPsqlSeparator(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c != '-' && c != '+' && c != ' ' {
			return false
		}
	}
	return strings.Contains(s, "-")
}

func isPsqlSummary(s string) bool {
	return strings.HasPrefix(s, "(") && (strings.Contains(s, " rows)") || strings.Contains(s, " row)"))
}

func isMySQLSummary(s string) bool {
	return strings.Contains(s, "rows in set") || strings.Contains(s, "row in set")
}

func isSqliteSummary(s string) bool {
	return strings.HasPrefix(s, "Run Time:") || strings.HasPrefix(s, "CPU:")
}

func looksLikeDbOutput(s string) bool {
	return (strings.Contains(s, "rows in set") || strings.Contains(s, " rows)") || strings.Contains(s, " row)")) ||
		(strings.Contains(s, "+---") && strings.Contains(s, "|")) ||
		isPsqlSeparator(strings.TrimSpace(strings.SplitN(s, "\n", 3)[0])) ||
		(strings.Contains(s, "|") && !strings.Contains(s, "CONTAINER") && !strings.Contains(s, "IMAGE"))
}
