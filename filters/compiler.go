package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reCompilerDiag   = regexp.MustCompile(`^(.+?):(\d+):\d+:\s*(error|warning|note):\s*(.+)`)
	reCompilerSource = regexp.MustCompile(`^\s*\d+\s*\|`)
	reCompilerCaret  = regexp.MustCompile(`^\s*\|?\s*[\^~]+`)
	reCompilerIncl   = regexp.MustCompile(`(?i)^In file included from`)
	reCompilerInFunc = regexp.MustCompile(`(?i)^.+?: In function`)
)

func filterCompiler(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeCompilerOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	type diag struct {
		file    string
		line    string
		level   string
		message string
	}

	var errorDiags []diag
	var warnDiags []diag

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip source lines, carets, include chains
		if reCompilerSource.MatchString(trimmed) || reCompilerCaret.MatchString(trimmed) ||
			reCompilerIncl.MatchString(trimmed) || reCompilerInFunc.MatchString(trimmed) {
			continue
		}

		if m := reCompilerDiag.FindStringSubmatch(trimmed); m != nil {
			d := diag{file: m[1], line: m[2], level: m[3], message: m[4]}
			switch d.level {
			case "error":
				errorDiags = append(errorDiags, d)
			case "warning":
				warnDiags = append(warnDiags, d)
			}
		}
	}

	if len(errorDiags) == 0 && len(warnDiags) == 0 {
		return raw, nil
	}

	var out []string

	if len(errorDiags) > 0 {
		out = append(out, fmt.Sprintf("errors(%d):", len(errorDiags)))
		for _, d := range errorDiags {
			out = append(out, fmt.Sprintf("  %s:%s: %s", d.file, d.line, d.message))
		}
	}

	if len(warnDiags) > 0 {
		out = append(out, fmt.Sprintf("warnings(%d):", len(warnDiags)))
		for _, d := range warnDiags {
			out = append(out, fmt.Sprintf("  %s:%s: %s", d.file, d.line, d.message))
		}
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
