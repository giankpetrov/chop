package filters

import (
	"strings"
)

// isDecorativeLine returns true for lines that are purely decorative
// (repeated dashes, asterisks, or similar characters).
func isDecorativeLine(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c != '-' && c != '*' && c != '=' && c != '#' && c != ' ' && c != '\t' {
			return false
		}
	}
	return len(strings.TrimSpace(s)) > 2
}

func filterSsh(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeSshOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")
	var out []string
	inPreamble := true

	for _, line := range lines {
		t := strings.TrimSpace(line)

		// Drop connection banner noise
		if strings.Contains(t, "Warning: Permanently added") ||
			strings.Contains(t, "The authenticity of host") ||
			strings.Contains(t, "key fingerprint is") ||
			strings.HasPrefix(t, "Welcome to") ||
			strings.HasPrefix(t, "Last login:") ||
			strings.HasPrefix(t, "System information") {
			continue
		}

		// Drop decorative lines in preamble
		if inPreamble && (isDecorativeLine(t) || t == "") {
			continue
		}

		// Once we see actual content, we're out of preamble
		inPreamble = false

		out = append(out, line)
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
