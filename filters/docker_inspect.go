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
	if len(trimmed) < 500 {
		return raw, nil
	}
	// Always deep-compress docker inspect output (skip isSmallJSON path)
	var parsed interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return raw, nil
	}
	result := compressJSONValue(parsed, 0)
	return outputSanityCheck(raw, result), nil
}
