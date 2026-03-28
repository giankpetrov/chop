package filters

import (
	"strings"
)

func filterBrew(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeBrewOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")
	var out []string

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		// Drop: ==> Fetching and ==> Downloading first (before general ==> keep)
		if strings.HasPrefix(t, "==> Fetching") ||
			strings.HasPrefix(t, "==> Downloading") {
			continue
		}

		// Keep other section headers
		if strings.HasPrefix(t, "==>") {
			out = append(out, t)
			continue
		}

		// Keep beer emoji lines (successful install)
		if strings.HasPrefix(t, "🍺") {
			out = append(out, t)
			continue
		}

		// Keep warnings and errors
		if strings.HasPrefix(t, "Warning:") ||
			strings.HasPrefix(t, "Error") ||
			strings.HasPrefix(t, "error") {
			out = append(out, t)
			continue
		}

		// Keep "already installed" and "up-to-date"
		if strings.Contains(t, "already installed") ||
			strings.Contains(t, "up-to-date") {
			out = append(out, t)
			continue
		}

		// Keep update summary lines
		if strings.Contains(t, "Updated") && strings.Contains(t, "tap") {
			out = append(out, t)
			continue
		}
		if strings.Contains(t, "new formulae") ||
			strings.Contains(t, "updated formulae") {
			out = append(out, t)
			continue
		}

		// Drop: Downloading, Already downloaded, ==> Fetching, progress bars, checksums, bottle paths
		if strings.HasPrefix(t, "Downloading") ||
			strings.HasPrefix(t, "Already downloaded:") ||
			strings.HasPrefix(t, "==> Fetching") ||
			strings.HasPrefix(t, "#") ||
			strings.HasPrefix(t, "###") ||
			strings.Contains(t, "sha256") ||
			strings.Contains(t, ".tar.gz") ||
			strings.Contains(t, ".bottle.") ||
			strings.Contains(t, "Cellar/") {
			continue
		}

		// Keep anything else that wasn't explicitly dropped
		out = append(out, t)
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

func getBrewFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "install", "upgrade", "reinstall", "update":
		return filterBrew
	default:
		return nil
	}
}
