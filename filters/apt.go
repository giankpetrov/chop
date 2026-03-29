package filters

import (
	"strings"
)

func filterApt(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeAptOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")
	var out []string

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		// Keep: Setting up, Upgrading, Installing
		if strings.HasPrefix(t, "Setting up") ||
			strings.HasPrefix(t, "Upgrading") ||
			strings.HasPrefix(t, "Installing") {
			out = append(out, t)
			continue
		}

		// Keep: summary lines
		if strings.Contains(t, "upgraded,") ||
			strings.Contains(t, "newly installed") ||
			strings.Contains(t, "to remove") ||
			strings.Contains(t, "not upgraded") {
			out = append(out, t)
			continue
		}

		// Keep: errors and warnings
		lower := strings.ToLower(t)
		if strings.HasPrefix(lower, "e:") ||
			strings.HasPrefix(lower, "err:") ||
			strings.HasPrefix(lower, "w:") ||
			strings.HasPrefix(lower, "warning:") ||
			strings.Contains(lower, "error") {
			out = append(out, t)
			continue
		}

		// Drop: network fetch lines
		if strings.HasPrefix(t, "Get:") ||
			strings.HasPrefix(t, "Fetched") ||
			strings.HasPrefix(t, "Reading database") ||
			strings.HasPrefix(t, "Preparing to unpack") ||
			strings.HasPrefix(t, "Processing triggers") ||
			strings.HasPrefix(t, "Unpacking") ||
			strings.HasPrefix(t, "Selecting previously") ||
			strings.HasPrefix(t, "Reading package lists") ||
			strings.HasPrefix(t, "Building dependency tree") ||
			strings.HasPrefix(t, "Scanning processes") ||
			strings.HasPrefix(t, "Scanning linux images") {
			continue
		}
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

func getAptFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "install", "upgrade", "update", "dist-upgrade":
		return filterApt
	default:
		return nil
	}
}
