package filters

import (
	"regexp"
	"strings"
)

var (
	rePytestSummary  = regexp.MustCompile(`(\d+\s+(?:failed|passed|error|skipped|warnings?|deselected)[\s,]*)+in\s+[\d.]+s`)
	rePytestFailed   = regexp.MustCompile(`^FAILED\s+(.+)`)
	rePytestShortSep = regexp.MustCompile(`^=+\s*short test summary`)
	rePytestSep      = regexp.MustCompile(`^=+\s*`)
)

func filterPytest(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikePytestOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	var failedTests []string
	var summaryLine string
	inShortSummary := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if rePytestShortSep.MatchString(trimmed) {
			inShortSummary = true
			continue
		}

		if inShortSummary {
			if rePytestSep.MatchString(trimmed) && !rePytestShortSep.MatchString(trimmed) {
				inShortSummary = false
			} else if m := rePytestFailed.FindStringSubmatch(trimmed); m != nil {
				failedTests = append(failedTests, trimmed)
			}
		}

		if rePytestSummary.MatchString(trimmed) {
			summaryLine = trimmed
		}
	}

	if summaryLine == "" {
		return raw, nil
	}

	var out []string
	for _, f := range failedTests {
		out = append(out, f)
	}
	if len(out) > 0 {
		out = append(out, "")
	}
	// Clean the summary line of = decorations
	summary := strings.Trim(summaryLine, "= ")
	out = append(out, summary)

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
