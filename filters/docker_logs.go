package filters

import "strings"

func filterDockerLogs(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	lines := strings.Split(trimmed, "\n")
	var result string
	if isJSONLogs(lines) {
		result = filterJSONLogs(lines)
	} else {
		result = filterTextLogs(lines)
	}
	return outputSanityCheck(raw, result), nil
}
