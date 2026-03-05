package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reNgBuildTime    = regexp.MustCompile(`(?i)(?:Build at:.*?Time:\s*(\d+)ms|Time:\s*(\d+)ms)`)
	reNgBuildHash    = regexp.MustCompile(`Hash:\s*(\w+)`)
	reNgInitialTotal = regexp.MustCompile(`Initial Total\s*\|\s*([\d.]+\s*\w+)(?:\s*\|\s*([\d.]+\s*\w+))?`)
	reNgBuildWarn    = regexp.MustCompile(`(?i)^Warning:`)
	reNgBuildErr     = regexp.MustCompile(`(?i)^Error:|^✖`)
	reNgListening    = regexp.MustCompile(`(?i)listening on\s+(.+?)(?:\s*\*\*)?$`)

	reNgTestTotal  = regexp.MustCompile(`(?i)^TOTAL:\s*(.+)`)
	reNgTestFailed = regexp.MustCompile(`(?i)✗\s+(.+)`)
)

func filterNgBuild(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeNgBuildOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	var buildTime, hash, initialSize, transferSize string
	var warnings, errors []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if m := reNgBuildTime.FindStringSubmatch(trimmed); m != nil {
			if m[1] != "" {
				buildTime = m[1] + "ms"
			} else if m[2] != "" {
				buildTime = m[2] + "ms"
			}
		}
		if m := reNgBuildHash.FindStringSubmatch(trimmed); m != nil {
			hash = m[1]
		}
		if m := reNgInitialTotal.FindStringSubmatch(trimmed); m != nil {
			initialSize = strings.TrimSpace(m[1])
			if m[2] != "" {
				transferSize = strings.TrimSpace(m[2])
			}
		}
		if reNgBuildWarn.MatchString(trimmed) {
			warnings = append(warnings, trimmed)
		}
		if reNgBuildErr.MatchString(trimmed) {
			errors = append(errors, trimmed)
		}
	}

	var out []string

	status := "build ok"
	if len(errors) > 0 {
		status = "build FAILED"
	}

	parts := []string{status}
	if buildTime != "" {
		parts = append(parts, buildTime)
	}
	if hash != "" {
		parts = append(parts, "hash "+hash)
	}
	out = append(out, strings.Join(parts, ", "))

	if initialSize != "" {
		sizeInfo := "Initial: " + initialSize
		if transferSize != "" {
			sizeInfo += " (" + transferSize + " transfer)"
		}
		out = append(out, sizeInfo)
	}

	for _, e := range errors {
		out = append(out, e)
	}
	if len(warnings) > 0 {
		out = append(out, fmt.Sprintf("%d warnings", len(warnings)))
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

func filterNgTest(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeNgTestOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	var totalLine string
	var failedTests []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if m := reNgTestTotal.FindStringSubmatch(trimmed); m != nil {
			totalLine = m[1]
		}
		if m := reNgTestFailed.FindStringSubmatch(trimmed); m != nil {
			failedTests = append(failedTests, m[1])
		}
	}

	if totalLine == "" {
		return raw, nil
	}

	var out []string
	for _, f := range failedTests {
		out = append(out, "FAILED: "+f)
	}
	if len(out) > 0 {
		out = append(out, "")
	}
	out = append(out, totalLine)

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

func filterNgServe(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeNgServeOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	var listenURL, initialSize string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if m := reNgListening.FindStringSubmatch(trimmed); m != nil {
			listenURL = strings.TrimSpace(m[1])
		}
		if m := reNgInitialTotal.FindStringSubmatch(trimmed); m != nil {
			initialSize = strings.TrimSpace(m[1])
		}
	}

	if listenURL == "" {
		return raw, nil
	}

	result := "serving on " + listenURL
	if initialSize != "" {
		result += " (" + initialSize + " initial)"
	}

	return outputSanityCheck(raw, result), nil
}
