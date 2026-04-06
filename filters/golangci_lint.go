package filters

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	// golangci-lint issue line: "path/file.go:42:9: message (lintername)"
	reGolangciIssue = regexp.MustCompile(`^(.+\.go:\d+(?::\d+)?): .+ \((\w+)\)\s*$`)
	// golangci-lint summary: "N issues."
	reGolangciSummary = regexp.MustCompile(`^(\d+) issues?\.$`)
)

func filterGolangciLint(raw string) (string, error) {
	trimmed := strings.TrimSpace(stripAnsi(raw))
	if trimmed == "" {
		return "no issues", nil
	}
	if !looksLikeGolangciLintOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")

	// linterLocs maps linter name -> []location strings
	linterLocs := make(map[string][]string)
	// linterOrder preserves first-seen order for deterministic output
	var linterOrder []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := reGolangciIssue.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		loc, linter := m[1], m[2]
		if _, exists := linterLocs[linter]; !exists {
			linterOrder = append(linterOrder, linter)
		}
		linterLocs[linter] = append(linterLocs[linter], loc)
	}

	if len(linterLocs) == 0 {
		return "no issues", nil
	}

	// Sort linters alphabetically for deterministic output
	sort.Strings(linterOrder)

	var out []string
	totalIssues := 0

	for _, linter := range linterOrder {
		locs := linterLocs[linter]
		totalIssues += len(locs)

		var samples []string
		if len(locs) <= 2 {
			samples = locs
		} else {
			samples = locs[:2]
		}

		line := fmt.Sprintf("%s (%d): %s", linter, len(locs), strings.Join(samples, ", "))
		if len(locs) > 2 {
			line += fmt.Sprintf(", +%d more", len(locs)-2)
		}
		out = append(out, line)
	}

	out = append(out, "")
	out = append(out, fmt.Sprintf("%d issue(s)", totalIssues))

	result := strings.Join(out, "\n")
	return outputSanityCheck(trimmed, result), nil
}
