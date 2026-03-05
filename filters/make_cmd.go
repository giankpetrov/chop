package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reMakeEnter   = regexp.MustCompile(`(?i)^make\[\d+\]:\s*(Entering|Leaving)\s+directory`)
	reMakeCompile = regexp.MustCompile(`(?i)^\s*(gcc|g\+\+|cc|c\+\+|clang|clang\+\+)\s+-c\s+`)
	reMakeWarn    = regexp.MustCompile(`:(\d+):\d+:\s*warning:`)
	reMakeErr     = regexp.MustCompile(`:(\d+):\d+:\s*error:`)
	reMakeLink    = regexp.MustCompile(`(?i)^\s*(gcc|g\+\+|cc|clang)\s+-o\s+`)
)

func filterMake(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeMakeOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	var warnings []string
	var errors []string
	compiled := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if reMakeEnter.MatchString(trimmed) {
			continue
		}
		if reMakeCompile.MatchString(trimmed) {
			compiled++
			continue
		}
		if reMakeErr.MatchString(trimmed) {
			errors = append(errors, trimmed)
			continue
		}
		if reMakeWarn.MatchString(trimmed) {
			warnings = append(warnings, trimmed)
			continue
		}
	}

	var out []string

	for _, e := range errors {
		out = append(out, e)
	}
	for _, w := range warnings {
		out = append(out, w)
	}

	if len(errors) > 0 {
		out = append(out, fmt.Sprintf("build FAILED (%d errors, %d warnings)", len(errors), len(warnings)))
	} else if compiled > 0 {
		msg := fmt.Sprintf("build ok (%d files compiled", compiled)
		if len(warnings) > 0 {
			msg += fmt.Sprintf(", %d warnings", len(warnings))
		}
		msg += ")"
		out = append(out, msg)
	} else if len(warnings) > 0 {
		out = append(out, fmt.Sprintf("build ok (%d warnings)", len(warnings)))
	} else {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
