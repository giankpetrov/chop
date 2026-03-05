package filters

import (
	"fmt"
	"strings"
)

func filterPsCmd(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikePsCmdOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	if len(lines) < 20 {
		return raw, nil
	}

	// Keep header + active processes (CPU > 0 or notable)
	header := lines[0]
	var active []string
	total := 0

	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		total++

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		// Keep processes with non-zero CPU (field index 2 for ps aux)
		cpu := fields[2]
		if cpu != "0.0" {
			active = append(active, line)
		}
	}

	if len(active) == 0 || len(active) >= total {
		return raw, nil
	}

	var out []string
	out = append(out, header)
	out = append(out, active...)
	out = append(out, fmt.Sprintf("%d active of %d total processes", len(active), total))

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
