package filters

import (
	"encoding/json"
	"strings"
)

func filterDockerInspect(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return raw, nil
	}

	// Always redact sensitive data even in small inspect output
	parsed = redactJSON(parsed)

	// If large enough, summarize structure. If small, preserve redacted values.
	if len(trimmed) < 500 && isSmallJSON(parsed) {
		b, err := json.MarshalIndent(parsed, "", "  ")
		if err != nil {
			return raw, nil
		}
		return string(b), nil
	}

	// Always deep-compress large docker inspect output
	result := compressJSONValue(parsed, 0)
	return outputSanityCheck(raw, result), nil
}
