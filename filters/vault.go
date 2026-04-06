package filters

import (
	"fmt"
	"strings"
)

var vaultSecretKeywords = []string{"password", "token", "secret", "key", "cert", "private"}

func isVaultSecretKey(k string) bool {
	lower := strings.ToLower(k)
	for _, kw := range vaultSecretKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func getVaultFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "read":
		return filterVaultRead
	case "kv":
		if len(args) >= 2 {
			switch args[1] {
			case "get":
				return filterVaultRead
			case "list":
				return filterVaultList
			}
		}
		return nil
	case "list":
		return filterVaultList
	case "secrets":
		if len(args) >= 2 && args[1] == "list" {
			return filterVaultMountList
		}
		return nil
	case "auth":
		if len(args) >= 2 && args[1] == "list" {
			return filterVaultMountList
		}
		return nil
	default:
		return nil
	}
}

// filterVaultRead handles `vault read` and `vault kv get`.
// Key=value pairs are shown; values for secret-like keys are redacted.
func filterVaultRead(raw string) (string, error) {
	raw = stripAnsi(strings.TrimSpace(raw))
	if raw == "" {
		return raw, nil
	}
	if !looksLikeVaultOutput(raw) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")
	var out []string
	inData := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Section headers like "====== Data ======"
		if strings.HasPrefix(trimmed, "=") && strings.HasSuffix(trimmed, "=") {
			label := strings.Trim(trimmed, "= ")
			if label != "" {
				out = append(out, "["+label+"]")
			}
			inData = true
			continue
		}
		// Skip separator lines (--- -----)
		if isSeparatorLine(trimmed) {
			continue
		}
		// Secret path lines (plain path, no spaces typically)
		if inData && !strings.Contains(trimmed, " ") && strings.Contains(trimmed, "/") {
			out = append(out, trimmed)
			inData = false
			continue
		}
		// Key-value table lines: "key    value"
		// Split on 2+ spaces to separate key from value
		if idx := indexTwoSpaces(trimmed); idx > 0 {
			k := strings.TrimSpace(trimmed[:idx])
			v := strings.TrimSpace(trimmed[idx:])
			if k == "Key" && v == "Value" {
				continue // skip header row
			}
			if isVaultSecretKey(k) {
				v = "[REDACTED]"
			}
			out = append(out, fmt.Sprintf("%s=%s", k, v))
		} else {
			// Might be a plain metadata value line (e.g. the path after a header)
			out = append(out, trimmed)
		}
	}

	if len(out) == 0 {
		return raw, nil
	}
	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

// filterVaultList handles `vault list` / `vault kv list`.
func filterVaultList(raw string) (string, error) {
	raw = stripAnsi(strings.TrimSpace(raw))
	if raw == "" {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")
	var items []string
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" || t == "Keys" || isSeparatorLine(t) {
			continue
		}
		items = append(items, t)
	}

	if len(items) == 0 {
		return raw, nil
	}
	if len(items) <= 10 {
		result := strings.Join(items, "\n")
		return outputSanityCheck(raw, result), nil
	}
	result := fmt.Sprintf("%d items: %s ... and %d more",
		len(items), strings.Join(items[:3], ", "), len(items)-3)
	return outputSanityCheck(raw, result), nil
}

// filterVaultMountList handles `vault secrets list` / `vault auth list`.
// Keeps only Path and Type columns.
func filterVaultMountList(raw string) (string, error) {
	raw = stripAnsi(strings.TrimSpace(raw))
	if raw == "" {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")
	if len(lines) == 0 {
		return raw, nil
	}

	// Find header line with "Path" and "Type"
	headerIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "Path") && strings.Contains(line, "Type") {
			headerIdx = i
			break
		}
	}
	if headerIdx < 0 {
		return raw, nil
	}

	header := lines[headerIdx]
	pathStart := strings.Index(header, "Path")
	typeStart := strings.Index(header, "Type")
	if pathStart < 0 || typeStart < 0 {
		return raw, nil
	}

	var out []string
	out = append(out, "Path\tType")

	for _, line := range lines[headerIdx+1:] {
		t := strings.TrimSpace(line)
		if t == "" || isSeparatorLine(t) {
			continue
		}
		pathVal := safeExtract(line, pathStart, typeStart)
		typeEnd := findColumnEnd(line, typeStart)
		typeVal := safeExtract(line, typeStart, typeEnd)
		if pathVal == "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s\t%s", pathVal, typeVal))
	}

	if len(out) <= 1 {
		return raw, nil
	}
	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

// findColumnEnd finds where a column value ends by looking for 2+ consecutive spaces.
func findColumnEnd(line string, start int) int {
	if start >= len(line) {
		return len(line)
	}
	inSpaces := false
	for i := start; i < len(line); i++ {
		if line[i] == ' ' {
			if inSpaces {
				return i - 1
			}
			inSpaces = true
		} else {
			inSpaces = false
		}
	}
	return len(line)
}

// indexTwoSpaces returns the index of the first run of 2+ spaces, or -1.
func indexTwoSpaces(s string) int {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == ' ' && s[i+1] == ' ' {
			return i
		}
	}
	return -1
}

func looksLikeVaultOutput(s string) bool {
	return strings.Contains(s, "lease_id") ||
		strings.Contains(s, "lease_duration") ||
		strings.Contains(s, "lease_renewable") ||
		strings.Contains(s, "Secret Path") ||
		strings.Contains(s, "cubbyhole/") ||
		(strings.Contains(s, "Path") && strings.Contains(s, "Accessor")) ||
		strings.Contains(s, "created_time") ||
		strings.Contains(s, "deletion_time")
}
