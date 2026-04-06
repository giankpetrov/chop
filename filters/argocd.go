package filters

import (
	"fmt"
	"strings"
)

// filterArgoCDAppList keeps NAME, STATUS, HEALTH columns only.
func filterArgoCDAppList(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeArgoCDAppListOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	if len(lines) < 2 {
		return raw, nil
	}

	header := lines[0]
	fields := strings.Fields(header)
	keepNames := []string{"NAME", "STATUS", "HEALTH"}
	keepCols := findIndices(fields, keepNames...)
	if len(keepCols) == 0 {
		return raw, nil
	}

	colBounds := findColumnBoundaries(header, fields)

	var out []string
	var headerParts []string
	for _, idx := range keepCols {
		if idx < len(fields) {
			headerParts = append(headerParts, fields[idx])
		}
	}
	out = append(out, strings.Join(headerParts, "\t"))

	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var parts []string
		for _, colIdx := range keepCols {
			if colIdx < len(colBounds) {
				val := safeExtract(line, colBounds[colIdx].start, colBounds[colIdx].end)
				parts = append(parts, val)
			}
		}
		out = append(out, strings.Join(parts, "\t"))
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

// filterArgoCDAppSync extracts resource changes and final sync/health status.
func filterArgoCDAppSync(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeArgoCDSyncOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	var resources []string
	var syncStatus string
	var healthStatus string

	// First non-empty, non-header line detection:
	// resource table has TIMESTAMP GROUP KIND NAMESPACE NAME STATUS HEALTH ...
	inTable := false

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		// Detect table header
		if strings.HasPrefix(t, "TIMESTAMP") {
			inTable = true
			continue
		}

		if inTable {
			// Once we hit a non-table line (e.g. "Name:"), stop table parsing
			if strings.Contains(t, ":") && !strings.HasPrefix(t, "20") {
				inTable = false
			} else {
				fields := strings.Fields(t)
				// TIMESTAMP GROUP KIND NAMESPACE NAME STATUS HEALTH HOOK MESSAGE
				// We want KIND(col2), NAME(col4), STATUS(col5)
				if len(fields) >= 6 {
					kind := fields[2]
					name := fields[4]
					status := fields[5]
					resources = append(resources, fmt.Sprintf("%s/%s %s", kind, name, status))
				}
				continue
			}
		}

		// Key: Value lines after the table
		if strings.HasPrefix(t, "Sync Status:") {
			syncStatus = strings.TrimSpace(strings.TrimPrefix(t, "Sync Status:"))
		} else if strings.HasPrefix(t, "Health Status:") {
			healthStatus = strings.TrimSpace(strings.TrimPrefix(t, "Health Status:"))
		}
	}

	var out []string
	if len(resources) > 0 {
		out = append(out, strings.Join(resources, "\n"))
	}
	if syncStatus != "" {
		out = append(out, "Sync: "+syncStatus)
	}
	if healthStatus != "" {
		out = append(out, "Health: "+healthStatus)
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

// filterArgoCDAppGet keeps Name, Sync Status, Health Status, and degraded/error conditions.
func filterArgoCDAppGet(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeArgoCDAppGetOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	keepPrefixes := []string{
		"Name:", "Sync Status:", "Health Status:", "Message:",
		"Conditions:", "Condition:", "Error:", "Warning:",
	}

	var out []string
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		for _, prefix := range keepPrefixes {
			if strings.HasPrefix(t, prefix) {
				out = append(out, t)
				break
			}
		}
		// Also keep lines that indicate degraded/error state
		lower := strings.ToLower(t)
		if strings.Contains(lower, "degraded") || strings.Contains(lower, "error") || strings.Contains(lower, "failed") {
			// avoid duplicates
			if len(out) == 0 || out[len(out)-1] != t {
				out = append(out, t)
			}
		}
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

func getArgoCDAppFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "list":
		return filterArgoCDAppList
	case "sync":
		return filterArgoCDAppSync
	case "get":
		return filterArgoCDAppGet
	case "diff":
		return filterKustomizeDiff // same diff format
	default:
		return nil
	}
}

func getArgoCDFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "app":
		return getArgoCDAppFilter(args[1:])
	default:
		return nil
	}
}
