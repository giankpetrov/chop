package filters

import (
	"fmt"
	"strings"
)

func filterDockerVolumeLs(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeDockerVolumeLsOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")
	if len(lines) < 2 {
		return raw, nil
	}

	header := lines[0]
	driverIdx := strings.Index(header, "DRIVER")
	nameIdx := strings.Index(header, "VOLUME NAME")

	if driverIdx == -1 || nameIdx == -1 {
		return raw, nil
	}

	var out []string
	count := 0
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		count++

		driver := extractColumn(line, driverIdx, nameIdx)
		name := extractColumn(line, nameIdx, len(line))
		out = append(out, fmt.Sprintf("%s (%s)", name, driver))
	}

	if count == 0 {
		return "", nil
	}

	out = append(out, fmt.Sprintf("%d volumes", count))
	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
