package filters

import "strings"

func filterKubectlTop(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeKubectlTopOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	// Keep header + all data lines + add count
	var dataCount int
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) != "" {
			dataCount++
		}
	}

	if dataCount == 0 {
		return raw, nil
	}

	// Already compact, just pass through (sanity check would catch it anyway)
	return raw, nil
}
