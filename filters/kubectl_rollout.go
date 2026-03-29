package filters

import (
	"strings"
)

func filterKubectlRolloutStatus(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	lines := strings.Split(trimmed, "\n")
	var out []string

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		// Keep: "Waiting for", "successfully rolled out", errors
		if strings.HasPrefix(t, "Waiting for") ||
			strings.Contains(t, "successfully rolled out") {
			out = append(out, t)
			continue
		}

		// Keep errors
		lower := strings.ToLower(t)
		if strings.Contains(lower, "error") ||
			strings.Contains(lower, "failed") {
			out = append(out, t)
			continue
		}

		// Drop other lines (progress details)
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
