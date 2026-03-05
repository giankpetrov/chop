package filters

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	reRubocopOffense = regexp.MustCompile(`^(.+?):(\d+):\d+:\s*[CWEF]:\s*(.+?):\s*(.+)`)
	reRubocopSummary = regexp.MustCompile(`(\d+)\s+files?\s+inspected,\s*(\d+)\s+offenses?`)
	reRubocopAuto    = regexp.MustCompile(`(\d+)\s+offenses?\s+autocorrectable`)
	reRubocopCaret   = regexp.MustCompile(`^\s*\^+\s*$`)
)

func filterRubocop(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "no offenses", nil
	}
	if !looksLikeRubocopOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	type offense struct {
		file string
		line string
		cop  string
		msg  string
	}

	var offenses []offense
	var summaryLine, autoLine string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if reRubocopCaret.MatchString(trimmed) {
			continue
		}
		if reRubocopSummary.MatchString(trimmed) {
			summaryLine = trimmed
			continue
		}
		if reRubocopAuto.MatchString(trimmed) {
			autoLine = trimmed
			continue
		}

		if m := reRubocopOffense.FindStringSubmatch(trimmed); m != nil {
			offenses = append(offenses, offense{
				file: m[1],
				line: m[2],
				cop:  m[3],
				msg:  m[4],
			})
		}
	}

	if len(offenses) == 0 {
		if summaryLine != "" {
			return summaryLine, nil
		}
		return raw, nil
	}

	// Group by cop
	copMap := make(map[string][]string)
	for _, o := range offenses {
		loc := fmt.Sprintf("%s:%s", o.file, o.line)
		copMap[o.cop] = append(copMap[o.cop], loc)
	}

	cops := make([]string, 0, len(copMap))
	for c := range copMap {
		cops = append(cops, c)
	}
	sort.Strings(cops)

	var out []string
	for _, cop := range cops {
		locs := copMap[cop]
		out = append(out, fmt.Sprintf("%s (%d): %s", cop, len(locs), strings.Join(locs, ", ")))
	}

	out = append(out, "")
	if summaryLine != "" {
		line := summaryLine
		if autoLine != "" {
			line += " (" + autoLine + ")"
		}
		out = append(out, line)
	} else {
		out = append(out, fmt.Sprintf("%d offenses", len(offenses)))
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
