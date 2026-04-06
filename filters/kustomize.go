package filters

import (
	"fmt"
	"strings"
)

func filterKustomizeBuild(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeKustomizeBuildOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	// Count each kind
	kindCounts := make(map[string]int)
	var kindOrder []string

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if !strings.HasPrefix(t, "kind:") {
			continue
		}
		kind := strings.TrimSpace(strings.TrimPrefix(t, "kind:"))
		if kind == "" {
			continue
		}
		if kindCounts[kind] == 0 {
			kindOrder = append(kindOrder, kind)
		}
		kindCounts[kind]++
	}

	total := 0
	for _, c := range kindCounts {
		total += c
	}

	if total == 0 {
		return raw, nil
	}

	// Short output: passthrough
	if total <= 3 {
		return raw, nil
	}

	var parts []string
	for _, k := range kindOrder {
		parts = append(parts, fmt.Sprintf("%s(%d)", k, kindCounts[k]))
	}

	result := fmt.Sprintf("%d resources: %s", total, strings.Join(parts, ", "))
	return outputSanityCheck(raw, result), nil
}

func filterKustomizeDiff(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeGitDiffOutput(trimmed) {
		return raw, nil
	}

	// Reuse kubectl diff approach: show resource headers and +/- summary
	raw = trimmed
	lines := strings.Split(raw, "\n")

	var changed []string
	var currentResource string
	added, removed := 0, 0

	for _, line := range lines {
		// Diff headers look like: diff -u .../Deployment.myapp.production ...
		if strings.HasPrefix(line, "diff ") {
			if currentResource != "" {
				changed = append(changed, fmt.Sprintf("%s (+%d/-%d)", currentResource, added, removed))
			}
			// Extract resource name from diff header
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				// take the path portion after last /
				path := parts[len(parts)-1]
				idx := strings.LastIndexByte(path, '/')
				if idx >= 0 {
					path = path[idx+1:]
				}
				currentResource = path
			} else {
				currentResource = line
			}
			added, removed = 0, 0
			continue
		}
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			added++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removed++
		}
	}
	if currentResource != "" {
		changed = append(changed, fmt.Sprintf("%s (+%d/-%d)", currentResource, added, removed))
	}

	if len(changed) == 0 {
		return raw, nil
	}

	result := fmt.Sprintf("%d changed resources:\n%s", len(changed), strings.Join(changed, "\n"))
	return outputSanityCheck(raw, result), nil
}

func getKustomizeFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "build":
		return filterKustomizeBuild
	case "diff":
		return filterKustomizeDiff
	default:
		return nil
	}
}
