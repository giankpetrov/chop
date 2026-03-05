package filters

import (
	"fmt"
	"strings"
)

func getGhIssueFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "list":
		return filterGhIssueList
	case "view":
		return filterGhIssueView
	default:
		return nil
	}
}

func filterGhIssueList(raw string) (string, error) {
	raw = stripAnsi(strings.TrimSpace(raw))
	if raw == "" {
		return "no issues", nil
	}

	lines := strings.Split(raw, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// gh issue list outputs tab-separated: NUMBER\tTITLE\tLABELS\tDATE
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			num := strings.TrimSpace(parts[0])
			title := strings.TrimSpace(parts[1])
			labels := ""
			if len(parts) >= 3 {
				labels = strings.TrimSpace(parts[2])
			}
			entry := fmt.Sprintf("#%s %s", num, title)
			if labels != "" {
				entry += " [" + labels + "]"
			}
			out = append(out, entry)
		} else {
			out = append(out, line)
		}
	}

	return strings.Join(out, "\n"), nil
}

func filterGhIssueView(raw string) (string, error) {
	raw = stripAnsi(strings.TrimSpace(raw))
	if raw == "" {
		return "no issue data", nil
	}

	lines := strings.Split(raw, "\n")
	var (
		title, state, author string
		labels               []string
	)

	inBody := false
	bodyFirstLine := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(lower, "title:") {
			title = strings.TrimSpace(trimmed[6:])
			inBody = false
		} else if strings.HasPrefix(lower, "state:") {
			state = strings.TrimSpace(trimmed[6:])
			inBody = false
		} else if strings.HasPrefix(lower, "author:") {
			author = strings.TrimSpace(trimmed[7:])
			inBody = false
		} else if strings.HasPrefix(lower, "labels:") {
			labelStr := strings.TrimSpace(trimmed[7:])
			if labelStr != "" {
				for _, l := range strings.Split(labelStr, ",") {
					l = strings.TrimSpace(l)
					if l != "" {
						labels = append(labels, l)
					}
				}
			}
			inBody = false
		} else if strings.HasPrefix(lower, "--") {
			inBody = true
		} else if inBody && bodyFirstLine == "" && trimmed != "" {
			bodyFirstLine = trimmed
			if len(bodyFirstLine) > 120 {
				bodyFirstLine = bodyFirstLine[:120] + "..."
			}
			inBody = false
		}
	}

	// Handle format where title is the first line
	if title == "" && len(lines) > 0 {
		first := strings.TrimSpace(lines[0])
		if first != "" && !strings.Contains(first, ":") {
			title = first
		}
	}

	var out []string
	if title != "" {
		line := title
		if state != "" {
			line += " [" + state + "]"
		}
		out = append(out, line)
	}
	if author != "" {
		out = append(out, "author: "+author)
	}
	if len(labels) > 0 {
		out = append(out, "labels: "+strings.Join(labels, ", "))
	}
	if bodyFirstLine != "" {
		out = append(out, bodyFirstLine)
	}

	if len(out) == 0 {
		return raw, nil
	}
	return strings.Join(out, "\n"), nil
}
