package filters

import (
	"strings"
)

// filterFluxGet keeps NAME, READY, MESSAGE columns; highlights not-ready items.
func filterFluxGet(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeFluxGetOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	if len(lines) < 2 {
		return raw, nil
	}

	header := lines[0]
	fields := strings.Fields(header)
	keepNames := []string{"NAME", "READY", "MESSAGE"}
	keepCols := findIndices(fields, keepNames...)
	if len(keepCols) == 0 {
		return raw, nil
	}

	// Also keep SUSPENDED column if present, but only when value is True
	suspendedIdx := -1
	for i, f := range fields {
		if f == "SUSPENDED" {
			suspendedIdx = i
			break
		}
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

		// Check if suspended — show suspended items regardless
		isSuspended := false
		if suspendedIdx >= 0 && suspendedIdx < len(colBounds) {
			suspendedVal := safeExtract(line, colBounds[suspendedIdx].start, colBounds[suspendedIdx].end)
			isSuspended = strings.EqualFold(suspendedVal, "true")
		}

		var parts []string
		for _, colIdx := range keepCols {
			if colIdx < len(colBounds) {
				val := safeExtract(line, colBounds[colIdx].start, colBounds[colIdx].end)
				parts = append(parts, val)
			}
		}

		row := strings.Join(parts, "\t")

		// Annotate not-ready or suspended rows
		if isSuspended {
			row += "\t[SUSPENDED]"
		} else if len(parts) >= 2 && !strings.EqualFold(parts[1], "true") {
			row += "\t[NOT READY]"
		}

		out = append(out, row)
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

// filterFluxReconcile shows final status lines (✔); drops annotating/waiting lines on success.
func filterFluxReconcile(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeFluxReconcileOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	var successes []string
	var errors []string
	succeeded := true

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		// ✔ lines are success status
		if strings.HasPrefix(t, "✔") {
			successes = append(successes, t)
			continue
		}

		// ✗ or lines containing "failed" / "error" indicate failure
		lower := strings.ToLower(t)
		if strings.HasPrefix(t, "✗") || strings.Contains(lower, "failed") || strings.Contains(lower, "error") {
			errors = append(errors, t)
			succeeded = false
			continue
		}

		// ► and ◎ lines: include only on failure (they provide context)
		if !succeeded {
			if strings.HasPrefix(t, "►") || strings.HasPrefix(t, "◎") {
				errors = append(errors, t)
			}
		}
	}

	var out []string
	if len(errors) > 0 {
		out = append(out, errors...)
	} else {
		out = append(out, successes...)
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

func getFluxFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "get":
		return filterFluxGet
	case "reconcile":
		return filterFluxReconcile
	default:
		return nil
	}
}
