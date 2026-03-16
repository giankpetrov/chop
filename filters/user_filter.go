package filters

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/giankpetrov/openchop/config"
)

// warnf writes a warning to stderr. Used for non-fatal config issues.
var warnf = func(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "openchop: warning: "+format+"\n", args...)
}

// BuildUserFilter creates a FilterFunc from a user-defined CustomFilter.
// Returns nil if the filter definition is empty/invalid.
func BuildUserFilter(cf *config.CustomFilter) FilterFunc {
	if cf == nil {
		return nil
	}

	// Exec-based filter takes priority - it's a full pipeline replacement
	if cf.Exec != "" {
		if !cf.Trusted {
			warnf("skipping untrusted exec filter for security. Move it to global filters.yml to enable.")
			return nil
		}
		return buildExecFilter(cf.Exec)
	}

	// Declarative rules: keep/drop regex + head/tail truncation
	if len(cf.Keep) == 0 && len(cf.Drop) == 0 && cf.Head == 0 && cf.Tail == 0 {
		return nil
	}

	return buildRuleFilter(cf.Keep, cf.Drop, cf.Head, cf.Tail)
}

// buildRuleFilter creates a FilterFunc from declarative keep/drop/head/tail rules.
func buildRuleFilter(keep, drop []string, head, tail int) FilterFunc {
	// Pre-compile regexes
	keepRe := compilePatterns(keep)
	dropRe := compilePatterns(drop)

	return func(raw string) (string, error) {
		if raw == "" {
			return "", nil
		}

		lines := strings.Split(raw, "\n")

		// Phase 1: Drop matching lines
		if len(dropRe) > 0 {
			var filtered []string
			for _, line := range lines {
				if matchesAny(line, dropRe) {
					continue
				}
				filtered = append(filtered, line)
			}
			lines = filtered
		}

		// Phase 2: Keep only matching lines
		if len(keepRe) > 0 {
			var filtered []string
			for _, line := range lines {
				// Skip empty lines - they add noise when filtering by pattern
				if strings.TrimSpace(line) == "" {
					continue
				}
				if matchesAny(line, keepRe) {
					filtered = append(filtered, line)
				}
			}
			lines = filtered
		}

		total := len(lines)

		// Phase 3: Head/tail truncation
		if head > 0 && tail > 0 && head+tail < total {
			headPart := strings.Join(lines[:head], "\n")
			tailPart := strings.Join(lines[total-tail:], "\n")
			hiddenStr := strconv.Itoa(total - head - tail)

			var b strings.Builder
			b.Grow(len(headPart) + len(tailPart) + len(hiddenStr) + 30)
			b.WriteString(headPart)
			b.WriteString("\n... (")
			b.WriteString(hiddenStr)
			b.WriteString(" lines hidden)\n")
			b.WriteString(tailPart)
			return b.String(), nil
		}

		if head > 0 && head < total {
			headPart := strings.Join(lines[:head], "\n")
			remainingStr := strconv.Itoa(total - head)

			var b strings.Builder
			b.Grow(len(headPart) + len(remainingStr) + 30)
			b.WriteString(headPart)
			b.WriteString("\n... (")
			b.WriteString(remainingStr)
			b.WriteString(" more lines)")
			return b.String(), nil
		}

		if tail > 0 && tail < total {
			tailPart := strings.Join(lines[total-tail:], "\n")
			skippedStr := strconv.Itoa(total - tail)

			var b strings.Builder
			b.Grow(len(tailPart) + len(skippedStr) + 30)
			b.WriteString("... (")
			b.WriteString(skippedStr)
			b.WriteString(" lines skipped)\n")
			b.WriteString(tailPart)
			return b.String(), nil
		}

		return strings.Join(lines, "\n"), nil
	}
}

// buildExecFilter creates a FilterFunc that pipes output through an external command.
func buildExecFilter(execCmd string) FilterFunc {
	return func(raw string) (string, error) {
		parts := splitCommand(execCmd)
		if len(parts) == 0 {
			return raw, fmt.Errorf("empty exec command")
		}

		for i, p := range parts {
			parts[i] = expandHome(p)
		}

		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Stdin = strings.NewReader(raw)

		out, err := cmd.Output()
		if err != nil {
			// On script failure, return raw output rather than losing data
			return raw, fmt.Errorf("exec filter failed (%s): %w", strings.Join(parts, " "), err)
		}

		return string(out), nil
	}
}

// splitCommand splits a command string into arguments, respecting quotes.
func splitCommand(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	var quoteChar rune

	for _, r := range s {
		switch {
		case inQuotes:
			if r == quoteChar {
				inQuotes = false
			} else {
				current.WriteRune(r)
			}
		case r == '"' || r == '\'':
			inQuotes = true
			quoteChar = r
		case r == ' ' || r == '\t':
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// compilePatterns compiles a list of regex pattern strings.
// Invalid patterns are skipped with a warning to stderr.
func compilePatterns(patterns []string) []*regexp.Regexp {
	var compiled []*regexp.Regexp
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			warnf("invalid regex pattern %q: %v", p, err)
			continue
		}
		compiled = append(compiled, re)
	}
	return compiled
}

// matchesAny returns true if the line matches any of the compiled patterns.
func matchesAny(line string, patterns []*regexp.Regexp) bool {
	for _, re := range patterns {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return strings.Replace(path, "~", home, 1)
}
