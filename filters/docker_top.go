package filters

import (
	"fmt"
	"strings"
)

func filterDockerTop(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	// Short output, pass through
	if len(lines) < 20 {
		return raw, nil
	}

	// Keep header + first 15 + tail 5
	var out []string
	out = append(out, lines[0])
	dataLines := lines[1:]

	headCount := 15
	tailCount := 5
	if len(dataLines) <= headCount+tailCount {
		return raw, nil
	}

	for i := 0; i < headCount; i++ {
		out = append(out, dataLines[i])
	}
	hidden := len(dataLines) - headCount - tailCount
	out = append(out, fmt.Sprintf("... (%d processes hidden)", hidden))
	for i := len(dataLines) - tailCount; i < len(dataLines); i++ {
		out = append(out, dataLines[i])
	}
	out = append(out, fmt.Sprintf("%d processes total", len(dataLines)))

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
