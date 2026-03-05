package filters

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	reMypyError   = regexp.MustCompile(`^(.+?):(\d+):\s*(error|warning|note):\s*(.+?)(?:\s+\[(.+?)\])?\s*$`)
	reMypySummary = regexp.MustCompile(`(?i)^Found\s+\d+\s+errors?\s+in`)
	reMypySuccess = regexp.MustCompile(`(?i)^Success:`)
)

func filterMypy(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeMypyOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	type mypyError struct {
		file string
		line string
		code string
		msg  string
	}

	var errors []mypyError
	var summaryLine string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if reMypySummary.MatchString(trimmed) || reMypySuccess.MatchString(trimmed) {
			summaryLine = trimmed
			continue
		}

		if m := reMypyError.FindStringSubmatch(trimmed); m != nil {
			level := m[3]
			if level == "note" {
				continue
			}
			code := m[5]
			if code == "" {
				code = "other"
			}
			errors = append(errors, mypyError{
				file: m[1],
				line: m[2],
				code: code,
				msg:  m[4],
			})
		}
	}

	if len(errors) == 0 && summaryLine != "" {
		return summaryLine, nil
	}
	if len(errors) == 0 {
		return raw, nil
	}

	// Group by code
	codeMap := make(map[string][]string)
	codeMsg := make(map[string]string)
	for _, e := range errors {
		loc := fmt.Sprintf("%s:%s", e.file, e.line)
		codeMap[e.code] = append(codeMap[e.code], loc)
		if _, ok := codeMsg[e.code]; !ok {
			codeMsg[e.code] = e.msg
		}
	}

	codes := make([]string, 0, len(codeMap))
	for c := range codeMap {
		codes = append(codes, c)
	}
	sort.Strings(codes)

	var out []string
	for _, code := range codes {
		locs := codeMap[code]
		out = append(out, fmt.Sprintf("[%s] (%d): %s", code, len(locs), codeMsg[code]))
		out = append(out, fmt.Sprintf("  %s", strings.Join(locs, ", ")))
	}

	if summaryLine != "" {
		out = append(out, "")
		out = append(out, summaryLine)
	} else {
		out = append(out, "")
		out = append(out, fmt.Sprintf("%d errors", len(errors)))
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
