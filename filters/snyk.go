package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Snyk dep vuln line: "  ✗ Cross-site Scripting (XSS) [High Severity][https://...] in pkg@version"
	reSnykDepVuln = regexp.MustCompile(`(?i)✗\s+(.+?)\s+\[(critical|high|medium|low)\s+severity\](?:\[https?://[^\]]*\])?\s+in\s+(\S+)`)
	// Snyk code vuln line: "  ✗ [High] SQL Injection"
	reSnykCodeVuln = regexp.MustCompile(`(?i)✗\s+\[(critical|high|medium|low)\]\s+(.+)`)
	// Snyk code path line: "     Path: src/db.js, line 45"
	reSnykCodePath = regexp.MustCompile(`(?i)^\s*path:\s*(\S+),\s*line\s*(\d+)`)
)

type snykVuln struct {
	severity string // "critical", "high", "medium", "low"
	name     string // vuln title
	location string // package@version or file:line
}

var snykSevOrder = []string{"critical", "high", "medium", "low"}

var snykSevLabel = map[string]string{
	"critical": "Critical",
	"high":     "High",
	"medium":   "Medium",
	"low":      "Low",
}

func snykFormatVulns(vulns []snykVuln, counts map[string]int, raw string) (string, error) {
	var out []string
	for _, sev := range snykSevOrder {
		var group []snykVuln
		for _, v := range vulns {
			if v.severity == sev {
				group = append(group, v)
			}
		}
		if len(group) == 0 {
			continue
		}
		out = append(out, fmt.Sprintf("%s (%d):", snykSevLabel[sev], len(group)))
		for _, v := range group {
			loc := v.location
			if loc == "" {
				loc = "(unknown)"
			}
			out = append(out, fmt.Sprintf("  %s — %s", loc, v.name))
		}
	}

	out = append(out, "")
	total := len(vulns)
	var parts []string
	for _, sev := range snykSevOrder {
		if counts[sev] > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", counts[sev], sev))
		}
	}
	out = append(out, fmt.Sprintf("%d issues found (%s)", total, strings.Join(parts, ", ")))

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

func filterSnykTest(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "no issues found", nil
	}
	if !looksLikeSnykOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")
	var vulns []snykVuln
	counts := map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0}

	for _, line := range lines {
		stripped := stripAnsi(line)
		if m := reSnykDepVuln.FindStringSubmatch(stripped); m != nil {
			sev := strings.ToLower(m[2])
			vulns = append(vulns, snykVuln{
				severity: sev,
				name:     strings.TrimSpace(m[1]),
				location: m[3],
			})
			counts[sev]++
		}
	}

	if len(vulns) == 0 {
		return "no issues found", nil
	}

	return snykFormatVulns(vulns, counts, trimmed)
}

func filterSnykCodeTest(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "no issues found", nil
	}
	if !looksLikeSnykOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")
	var vulns []snykVuln
	counts := map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0}

	var pendingVuln *snykVuln
	for _, line := range lines {
		stripped := stripAnsi(line)

		if m := reSnykCodeVuln.FindStringSubmatch(stripped); m != nil {
			if pendingVuln != nil {
				vulns = append(vulns, *pendingVuln)
				counts[pendingVuln.severity]++
			}
			sev := strings.ToLower(m[1])
			pendingVuln = &snykVuln{
				severity: sev,
				name:     strings.TrimSpace(m[2]),
			}
			continue
		}

		if pendingVuln != nil {
			if m := reSnykCodePath.FindStringSubmatch(stripped); m != nil {
				pendingVuln.location = fmt.Sprintf("%s:%s", m[1], m[2])
				vulns = append(vulns, *pendingVuln)
				counts[pendingVuln.severity]++
				pendingVuln = nil
			}
		}
	}
	if pendingVuln != nil {
		vulns = append(vulns, *pendingVuln)
		counts[pendingVuln.severity]++
	}

	if len(vulns) == 0 {
		if strings.Contains(trimmed, "no issues") || strings.Contains(trimmed, "✔") {
			return "no issues found", nil
		}
		return raw, nil
	}

	return snykFormatVulns(vulns, counts, trimmed)
}

func getSnykFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return func(raw string) (string, error) { return filterSnykTest(raw) }
	}
	switch args[0] {
	case "code":
		return func(raw string) (string, error) { return filterSnykCodeTest(raw) }
	case "test":
		return func(raw string) (string, error) { return filterSnykTest(raw) }
	default:
		return nil
	}
}
