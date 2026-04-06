package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Timestamped INFO log line: "2024-01-15T10:23:45.678Z	INFO	..."
	reTrivyLog = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)
	// Target header + total: "myapp:latest (debian 11.6)" followed by "Total: N (UNKNOWN: 0, LOW: 18, ...)"
	reTrivyTotal = regexp.MustCompile(`(?i)^Total:\s*(\d+)\s*\(([^)]+)\)`)
	// Individual vuln table row — detect by │ box-drawing char
	// Format: │ pkg │ CVE-ID │ SEVERITY │ installed │ fixed │ title │
	reTrivyTableRow = regexp.MustCompile(`^│\s*(\S+)\s*│\s*(CVE-\S+|GHSA-\S+)\s*│\s*(CRITICAL|HIGH|MEDIUM|LOW|UNKNOWN)\s*│[^│]*│\s*([^│]*?)\s*│\s*(.+?)\s*│`)
	// Severity count inside total line: "HIGH: 5"
	reTrivySevCount = regexp.MustCompile(`(?i)(CRITICAL|HIGH|MEDIUM|LOW|UNKNOWN):\s*(\d+)`)
	// Target section header — non-empty line NOT starting with table chars, spaces, or digits,
	// followed (somewhere) by a line of ===
	// We detect the === separator line instead.
	reTrivySectionSep = regexp.MustCompile(`^[=─]+$`)
)

type trivyVuln struct {
	pkg      string
	cveID    string
	severity string
	fixed    string
	title    string
}

type trivyTarget struct {
	name   string
	total  int
	counts map[string]int
	vulns  []trivyVuln
}

func filterTrivy(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "No vulnerabilities found", nil
	}
	if !looksLikeTrivyOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")

	var targets []trivyTarget
	var current *trivyTarget
	var pendingHeader string

	for i, line := range lines {
		stripped := stripAnsi(line)
		t := strings.TrimSpace(stripped)

		if t == "" {
			continue
		}

		// Skip INFO log lines
		if reTrivyLog.MatchString(t) {
			continue
		}

		// Skip box-drawing border lines (top/bottom/separator rows with no data)
		if strings.ContainsAny(t, "┌┬┐├┼┤└┴┘") {
			continue
		}

		// Skip header row of table (Library, Vulnerability, Severity, ...)
		if strings.Contains(t, "│") && strings.Contains(t, "Library") && strings.Contains(t, "Vulnerability") {
			continue
		}

		// Detect section separator (===...) — the line before it is the target name
		if reTrivySectionSep.MatchString(t) {
			if i > 0 {
				// Walk back to find the last non-empty, non-log line
				for j := i - 1; j >= 0; j-- {
					prev := strings.TrimSpace(stripAnsi(lines[j]))
					if prev == "" || reTrivyLog.MatchString(prev) {
						continue
					}
					pendingHeader = prev
					break
				}
			}
			if current != nil {
				targets = append(targets, *current)
			}
			current = &trivyTarget{
				name:   pendingHeader,
				counts: make(map[string]int),
			}
			pendingHeader = ""
			continue
		}

		if current == nil {
			continue
		}

		// Total line
		if m := reTrivyTotal.FindStringSubmatch(t); m != nil {
			fmt.Sscanf(m[1], "%d", &current.total)
			// Parse individual severity counts
			for _, sm := range reTrivySevCount.FindAllStringSubmatch(m[2], -1) {
				sev := strings.ToUpper(sm[1])
				var n int
				fmt.Sscanf(sm[2], "%d", &n)
				current.counts[sev] = n
			}
			continue
		}

		// Vuln table row
		if m := reTrivyTableRow.FindStringSubmatch(stripped); m != nil {
			current.vulns = append(current.vulns, trivyVuln{
				pkg:      strings.TrimSpace(m[1]),
				cveID:    strings.TrimSpace(m[2]),
				severity: strings.ToUpper(strings.TrimSpace(m[3])),
				fixed:    strings.TrimSpace(m[4]),
				title:    strings.TrimSpace(m[5]),
			})
			continue
		}
	}

	if current != nil {
		targets = append(targets, *current)
	}

	if len(targets) == 0 {
		return "No vulnerabilities found", nil
	}

	var out []string
	allZero := true
	for _, tgt := range targets {
		if tgt.total > 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return "No vulnerabilities found", nil
	}

	sevOrder := []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"}

	for _, tgt := range targets {
		if tgt.total == 0 {
			out = append(out, fmt.Sprintf("%s: No vulnerabilities found", tgt.name))
			continue
		}

		// Build summary counts string
		var parts []string
		for _, sev := range sevOrder {
			if n, ok := tgt.counts[sev]; ok && n > 0 {
				parts = append(parts, fmt.Sprintf("%s: %d", sev, n))
			}
		}
		summary := fmt.Sprintf("%s: Total %d", tgt.name, tgt.total)
		if len(parts) > 0 {
			summary += fmt.Sprintf(" (%s)", strings.Join(parts, ", "))
		}
		out = append(out, summary)

		// Show CRITICAL and HIGH vuln details
		for _, v := range tgt.vulns {
			if v.severity != "CRITICAL" && v.severity != "HIGH" {
				continue
			}
			fixed := v.fixed
			if fixed == "" {
				fixed = "no fix"
			}
			out = append(out, fmt.Sprintf("  %s %s: %s (fixed: %s)", v.pkg, v.cveID, v.title, fixed))
		}
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(trimmed, result), nil
}

func getTrivyFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return func(raw string) (string, error) { return filterTrivy(raw) }
	}
	switch args[0] {
	case "image", "fs", "repo", "config":
		return func(raw string) (string, error) { return filterTrivy(raw) }
	default:
		return nil
	}
}
