package filters

import (
	"fmt"
	"strings"
)

const glabListThreshold = 5

func getGlabFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "mr":
		if len(args) >= 2 && args[1] == "list" {
			return filterGlabMrList
		}
		return nil
	case "issue":
		if len(args) >= 2 && args[1] == "list" {
			return filterGlabIssueList
		}
		return nil
	case "ci":
		return filterGlabCiStatus
	case "pipeline":
		return filterGlabCiStatus
	default:
		return nil
	}
}

func filterGlabMrList(raw string) (string, error) {
	raw = stripAnsi(strings.TrimSpace(raw))
	if raw == "" {
		return "no merge requests", nil
	}
	if !looksLikeGlabOutput(raw) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")
	var header string
	var items []string

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if isGlabHeaderLine(t) {
			header = t
			continue
		}
		items = append(items, t)
	}

	if len(items) <= glabListThreshold {
		return raw, nil
	}

	var out []string
	if header != "" {
		out = append(out, header)
	}
	for i := 0; i < glabListThreshold; i++ {
		out = append(out, items[i])
	}
	out = append(out, fmt.Sprintf("... and %d more", len(items)-glabListThreshold))
	out = append(out, fmt.Sprintf("total: %d", len(items)))

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

func filterGlabIssueList(raw string) (string, error) {
	raw = stripAnsi(strings.TrimSpace(raw))
	if raw == "" {
		return "no issues", nil
	}
	if !looksLikeGlabOutput(raw) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")
	var header string
	var items []string

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if isGlabHeaderLine(t) {
			header = t
			continue
		}
		items = append(items, t)
	}

	if len(items) <= glabListThreshold {
		return raw, nil
	}

	var out []string
	if header != "" {
		out = append(out, header)
	}
	for i := 0; i < glabListThreshold; i++ {
		out = append(out, items[i])
	}
	out = append(out, fmt.Sprintf("... and %d more", len(items)-glabListThreshold))
	out = append(out, fmt.Sprintf("total: %d", len(items)))

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

// filterGlabCiStatus filters `glab ci status` / `glab pipeline` output.
func filterGlabCiStatus(raw string) (string, error) {
	raw = stripAnsi(strings.TrimSpace(raw))
	if raw == "" {
		return raw, nil
	}
	if !looksLikeGlabOutput(raw) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")

	type job struct {
		stage  string
		name   string
		status string
	}

	var pipelineLines []string
	var jobs []job
	inTable := false

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		lower := strings.ToLower(t)

		if strings.HasPrefix(t, "•") || strings.HasPrefix(lower, "pipeline") {
			pipelineLines = append(pipelineLines, t)
			continue
		}
		if isGlabCiHeaderLine(t) {
			inTable = true
			continue
		}
		if !inTable {
			continue
		}
		parts := splitByDoubleSpace(t)
		if len(parts) >= 3 {
			j := job{
				stage:  strings.ToLower(parts[0]),
				name:   parts[1],
				status: strings.ToLower(parts[2]),
			}
			jobs = append(jobs, j)
		}
	}

	var out []string
	out = append(out, pipelineLines...)

	priority := func(status string) int {
		switch status {
		case "failed":
			return 0
		case "running":
			return 1
		default:
			return 2
		}
	}
	sortedJobs := make([]job, len(jobs))
	copy(sortedJobs, jobs)
	for i := 1; i < len(sortedJobs); i++ {
		for j := i; j > 0 && priority(sortedJobs[j].status) < priority(sortedJobs[j-1].status); j-- {
			sortedJobs[j], sortedJobs[j-1] = sortedJobs[j-1], sortedJobs[j]
		}
	}

	for _, j := range sortedJobs {
		out = append(out, fmt.Sprintf("%s/%s: %s", j.stage, j.name, j.status))
	}

	if len(out) == 0 {
		return raw, nil
	}
	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

func isGlabHeaderLine(s string) bool {
	caps := 0
	for _, word := range strings.Fields(s) {
		// Only count words that contain at least one letter and are all-uppercase.
		// This excludes numbers (!99, 10) and symbols that trivially equal their ToUpper.
		if len(word) > 1 && containsLetter(word) && word == strings.ToUpper(word) {
			caps++
		}
	}
	return caps >= 2
}

func containsLetter(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

func isGlabCiHeaderLine(s string) bool {
	upper := strings.ToUpper(s)
	return strings.Contains(upper, "STAGE") && strings.Contains(upper, "STATUS")
}

func splitByDoubleSpace(s string) []string {
	var parts []string
	var cur strings.Builder
	spaceRun := 0

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' {
			spaceRun++
			if spaceRun == 1 {
				cur.WriteByte(c)
			}
		} else {
			if spaceRun >= 2 {
				field := strings.TrimSpace(cur.String())
				if field != "" {
					parts = append(parts, field)
				}
				cur.Reset()
			}
			spaceRun = 0
			cur.WriteByte(c)
		}
	}
	if field := strings.TrimSpace(cur.String()); field != "" {
		parts = append(parts, field)
	}
	return parts
}

func looksLikeGlabOutput(s string) bool {
	return strings.Contains(s, "OPENED") ||
		strings.Contains(s, "CLOSED") ||
		strings.Contains(s, "MERGED") ||
		strings.Contains(s, "CREATED_AT") ||
		(strings.Contains(s, "MR") && strings.Contains(s, "BRANCH")) ||
		(strings.Contains(s, "Pipeline") && strings.Contains(s, "triggered")) ||
		(strings.Contains(s, "STAGE") && strings.Contains(s, "STATUS")) ||
		strings.Contains(s, "glab")
}
