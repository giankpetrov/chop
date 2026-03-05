package filters

import (
	"fmt"
	"strings"
)

func filterDockerNetworkLs(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeDockerNetworkLsOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")
	if len(lines) < 2 {
		return raw, nil
	}

	header := lines[0]
	nameIdx := strings.Index(header, "NAME")
	driverIdx := strings.Index(header, "DRIVER")
	scopeIdx := strings.Index(header, "SCOPE")

	if nameIdx == -1 || driverIdx == -1 {
		return raw, nil
	}

	nameBound := driverIdx
	driverBound := scopeIdx
	if driverBound == -1 {
		driverBound = len(header)
	}

	var out []string
	count := 0
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		count++

		name := extractColumn(line, nameIdx, nameBound)
		driver := extractColumn(line, driverIdx, driverBound)
		scope := ""
		if scopeIdx != -1 {
			scope = extractColumn(line, scopeIdx, len(line))
		}

		entry := fmt.Sprintf("%s (%s)", name, driver)
		if scope != "" {
			entry += " " + scope
		}
		out = append(out, entry)
	}

	if count == 0 {
		return "no networks", nil
	}

	out = append(out, fmt.Sprintf("%d networks", count))
	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
