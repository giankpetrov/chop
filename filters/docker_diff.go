package filters

import (
	"fmt"
	"strings"
)

func filterDockerDiff(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeDockerDiffOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	added := 0
	changed := 0
	deleted := 0
	var addedPaths, changedPaths, deletedPaths []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(line) < 3 {
			continue
		}

		switch line[0] {
		case 'A':
			added++
			if added <= 10 {
				addedPaths = append(addedPaths, strings.TrimSpace(line[1:]))
			}
		case 'C':
			changed++
			if changed <= 10 {
				changedPaths = append(changedPaths, strings.TrimSpace(line[1:]))
			}
		case 'D':
			deleted++
			if deleted <= 10 {
				deletedPaths = append(deletedPaths, strings.TrimSpace(line[1:]))
			}
		}
	}

	total := added + changed + deleted
	if total == 0 {
		return raw, nil
	}

	var out []string

	if total <= 10 {
		// Show all paths when few changes
		for _, p := range addedPaths {
			out = append(out, "A "+p)
		}
		for _, p := range changedPaths {
			out = append(out, "C "+p)
		}
		for _, p := range deletedPaths {
			out = append(out, "D "+p)
		}
	} else {
		// Summary only
		var parts []string
		if added > 0 {
			parts = append(parts, fmt.Sprintf("added(%d)", added))
		}
		if changed > 0 {
			parts = append(parts, fmt.Sprintf("changed(%d)", changed))
		}
		if deleted > 0 {
			parts = append(parts, fmt.Sprintf("deleted(%d)", deleted))
		}
		out = append(out, strings.Join(parts, " "))
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
