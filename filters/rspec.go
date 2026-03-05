package filters

import (
	"regexp"
	"strings"
)

var (
	reRspecSummary    = regexp.MustCompile(`(\d+)\s+examples?,\s*(\d+)\s+failures?`)
	reRspecFinished   = regexp.MustCompile(`(?i)^Finished in\s+([\d.]+)\s+seconds?`)
	reRspecFailedLine = regexp.MustCompile(`^rspec\s+(.+)`)
)

func filterRspec(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeRspecOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	var summaryLine, timeLine string
	var failedExamples []string
	inFailedExamples := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "Failed examples:" {
			inFailedExamples = true
			continue
		}

		if inFailedExamples {
			if m := reRspecFailedLine.FindStringSubmatch(trimmed); m != nil {
				failedExamples = append(failedExamples, trimmed)
			}
		}

		if reRspecSummary.MatchString(trimmed) {
			summaryLine = trimmed
		}
		if m := reRspecFinished.FindStringSubmatch(trimmed); m != nil {
			timeLine = m[1] + "s"
		}
	}

	if summaryLine == "" {
		return raw, nil
	}

	var out []string
	for _, f := range failedExamples {
		out = append(out, f)
	}
	if len(out) > 0 {
		out = append(out, "")
	}

	summary := summaryLine
	if timeLine != "" {
		summary += " (" + timeLine + ")"
	}
	out = append(out, summary)

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
