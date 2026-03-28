package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var reJournalLine = regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+\S+\s+`)
var reJournalRepeat = regexp.MustCompile(`^-- last message repeated \d+ times? --`)

// isJournalImportant returns true if the line contains a high-severity log level.
func isJournalImportant(line string) bool {
	upper := strings.ToUpper(line)
	return strings.Contains(upper, "ERROR") ||
		strings.Contains(upper, "WARN") ||
		strings.Contains(upper, "CRIT") ||
		strings.Contains(upper, "ALERT") ||
		strings.Contains(upper, "EMERG")
}

func filterJournalctl(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeJournalctlOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")

	// Collapse consecutive repeated messages
	var deduped []string
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		// Collapse repeated messages
		if reJournalRepeat.MatchString(t) {
			deduped = append(deduped, t)
			continue
		}
		// Simple dedup: skip if same as last line (ignoring timestamp)
		if i > 0 {
			prev := strings.TrimSpace(lines[i-1])
			// Strip timestamp from both for comparison
			normCurr := reJournalLine.ReplaceAllString(t, "")
			normPrev := reJournalLine.ReplaceAllString(prev, "")
			if normCurr == normPrev && normCurr != "" {
				// Replace last with repeated indicator
				if len(deduped) > 0 && !reJournalRepeat.MatchString(deduped[len(deduped)-1]) {
					deduped = append(deduped, "-- last message repeated --")
				}
				continue
			}
		}
		deduped = append(deduped, t)
	}

	// If output is short enough, return as-is
	if len(deduped) <= 50 {
		return strings.Join(deduped, "\n"), nil
	}

	// Output > 50 lines: keep first 10 + last 20, always keeping ERROR/WARN/CRIT lines
	var important []string
	for _, line := range deduped {
		if isJournalImportant(line) {
			important = append(important, line)
		}
	}

	first10 := deduped[:10]
	last20 := deduped[len(deduped)-20:]
	hidden := len(deduped) - 30

	var out []string
	out = append(out, first10...)
	if hidden > 0 {
		out = append(out, fmt.Sprintf("... (%d lines hidden)", hidden))
	}
	out = append(out, last20...)

	// Prepend any important lines that aren't already in the window
	var extraImportant []string
	for _, imp := range important {
		inFirst := false
		for _, f := range first10 {
			if f == imp {
				inFirst = true
				break
			}
		}
		inLast := false
		for _, l := range last20 {
			if l == imp {
				inLast = true
				break
			}
		}
		if !inFirst && !inLast {
			extraImportant = append(extraImportant, imp)
		}
	}

	if len(extraImportant) > 0 {
		result := append(extraImportant, out...)
		return strings.Join(result, "\n"), nil
	}

	return strings.Join(out, "\n"), nil
}
