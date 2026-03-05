package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reYarnDone = regexp.MustCompile(`(?i)^(?:Done in|✨\s*Done in)\s+[\d.]+s`)
	reYarnWarn = regexp.MustCompile(`(?i)^warning|WARN`)
	reYarnErr  = regexp.MustCompile(`(?i)^error|ERR`)
)

func filterYarnInstall(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikeYarnInstallOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")

	var doneLine string
	var warnings int
	var errors []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if reYarnDone.MatchString(trimmed) {
			doneLine = trimmed
			continue
		}
		if reYarnErr.MatchString(trimmed) {
			errors = append(errors, trimmed)
			continue
		}
		if reYarnWarn.MatchString(trimmed) {
			warnings++
			continue
		}
	}

	var out strings.Builder
	if doneLine != "" {
		out.WriteString(doneLine)
	}

	for _, e := range errors {
		fmt.Fprintf(&out, "\n%s", e)
	}
	if warnings > 0 {
		fmt.Fprintf(&out, "\n%d warnings", warnings)
	}

	result := strings.TrimSpace(out.String())
	if result == "" {
		return raw, nil
	}
	return outputSanityCheck(raw, result), nil
}
