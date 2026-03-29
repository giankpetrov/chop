package filters

import (
	"strings"
)

func filterDockerPull(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeDockerPullOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")
	var out []string

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		// Keep: Status and Digest lines
		if strings.HasPrefix(t, "Status:") ||
			strings.HasPrefix(t, "Digest:") {
			out = append(out, t)
			continue
		}

		// Keep: final image reference (docker.io/... or other registry)
		if strings.Contains(t, "docker.io/") && !strings.Contains(t, "Pulling from") {
			out = append(out, t)
			continue
		}

		// Keep: error lines
		lower := strings.ToLower(t)
		if strings.Contains(lower, "error") ||
			strings.Contains(lower, "denied") ||
			strings.Contains(lower, "unauthorized") {
			out = append(out, t)
			continue
		}

		// Drop: individual layer lines
		if strings.Contains(t, ": Pull complete") ||
			strings.Contains(t, ": Waiting") ||
			strings.Contains(t, ": Downloading") ||
			strings.Contains(t, ": Extracting") ||
			strings.Contains(t, ": Verifying Checksum") ||
			strings.Contains(t, ": Already exists") ||
			strings.Contains(t, ": Download complete") ||
			strings.HasPrefix(t, "Using default tag:") ||
			strings.HasPrefix(t, "Pulling from") ||
			strings.Contains(t, "Pulling from") {
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
