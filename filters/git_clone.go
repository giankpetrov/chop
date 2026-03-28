package filters

import (
	"strings"
)

func filterGitClone(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeGitCloneOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")
	var out []string

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		// Keep: "Cloning into"
		if strings.HasPrefix(t, "Cloning into") {
			out = append(out, t)
			continue
		}

		// Keep: "remote: Total"
		if strings.HasPrefix(t, "remote: Total") {
			out = append(out, t)
			continue
		}

		// Keep: "Resolving deltas" when done
		if strings.HasPrefix(t, "Resolving deltas:") && strings.Contains(t, "done") {
			out = append(out, t)
			continue
		}

		// Keep: error lines
		lower := strings.ToLower(t)
		if strings.Contains(lower, "error") ||
			strings.Contains(lower, "fatal") ||
			strings.Contains(lower, "warning") {
			out = append(out, t)
			continue
		}

		// Drop: progress/counting lines
		if strings.HasPrefix(t, "remote: Enumerating") ||
			strings.HasPrefix(t, "remote: Counting") ||
			strings.HasPrefix(t, "remote: Compressing") ||
			strings.HasPrefix(t, "Receiving objects:") ||
			strings.HasPrefix(t, "remote:") {
			continue
		}

		// Keep anything else
		out = append(out, t)
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
