package filters

import (
	"fmt"
	"strings"
)

func filterDf(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeDfOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	if len(lines) < 3 {
		return raw, nil
	}

	// Skip header, filter out tmpfs/udev/devpts/overlay noise
	skipPrefixes := []string{"tmpfs", "udev", "devpts", "overlay", "shm", "none"}

	var kept []string
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		fs := fields[0]
		skip := false
		for _, prefix := range skipPrefixes {
			if strings.HasPrefix(fs, prefix) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Format: filesystem used/size (use%) mountpoint
		if len(fields) >= 5 {
			kept = append(kept, fmt.Sprintf("%s %s/%s (%s) %s", fields[0], fields[2], fields[1], fields[4], fields[5]))
		} else {
			kept = append(kept, line)
		}
	}

	if len(kept) == 0 {
		return raw, nil
	}

	result := strings.Join(kept, "\n")
	return outputSanityCheck(raw, result), nil
}
