package filters

import (
	"fmt"
	"strings"
)

func filterHelmInstall(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeHelmInstallOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	var name, namespace, status, revision string
	inNotes := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "NOTES:") {
			inNotes = true
			continue
		}
		if inNotes {
			continue
		}

		if strings.HasPrefix(trimmed, "NAME:") {
			name = strings.TrimSpace(strings.TrimPrefix(trimmed, "NAME:"))
		} else if strings.HasPrefix(trimmed, "NAMESPACE:") {
			namespace = strings.TrimSpace(strings.TrimPrefix(trimmed, "NAMESPACE:"))
		} else if strings.HasPrefix(trimmed, "STATUS:") {
			status = strings.TrimSpace(strings.TrimPrefix(trimmed, "STATUS:"))
		} else if strings.HasPrefix(trimmed, "REVISION:") {
			revision = strings.TrimSpace(strings.TrimPrefix(trimmed, "REVISION:"))
		}
	}

	if name == "" && status == "" {
		return raw, nil
	}

	parts := []string{name}
	if status != "" {
		parts = append(parts, status)
	}

	var details []string
	if revision != "" {
		details = append(details, "revision "+revision)
	}
	if namespace != "" {
		details = append(details, "namespace "+namespace)
	}

	result := strings.Join(parts, " ")
	if len(details) > 0 {
		result += " (" + strings.Join(details, ", ") + ")"
	}

	return outputSanityCheck(raw, result), nil
}

func filterHelmList(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeHelmListOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")
	if len(lines) < 2 {
		return raw, nil
	}

	header := lines[0]
	nameIdx := strings.Index(header, "NAME")
	nsIdx := strings.Index(header, "NAMESPACE")
	statusIdx := strings.Index(header, "STATUS")
	chartIdx := strings.Index(header, "CHART")
	appVerIdx := strings.Index(header, "APP VERSION")

	if nameIdx == -1 {
		return raw, nil
	}

	nameBound := nsIdx
	if nameBound == -1 {
		nameBound = len(header)
	}

	var out []string
	count := 0
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		count++

		name := extractColumn(line, nameIdx, nameBound)
		ns := ""
		if nsIdx != -1 {
			nsBound := statusIdx
			if nsBound == -1 {
				nsBound = len(header)
			}
			// Find the revision idx to bound namespace
			revIdx := strings.Index(header, "REVISION")
			if revIdx != -1 && revIdx < nsBound {
				nsBound = revIdx
			}
			ns = extractColumn(line, nsIdx, nsBound)
		}
		status := ""
		if statusIdx != -1 {
			sBound := chartIdx
			if sBound == -1 {
				sBound = len(header)
			}
			status = extractColumn(line, statusIdx, sBound)
		}
		chart := ""
		if chartIdx != -1 {
			cBound := appVerIdx
			if cBound == -1 {
				cBound = len(header)
			}
			chart = extractColumn(line, chartIdx, cBound)
		}

		entry := name
		if ns != "" {
			entry += " (" + ns + ")"
		}
		if status != "" {
			entry += " " + status
		}
		if chart != "" {
			entry += " " + chart
		}
		out = append(out, entry)
	}

	if count == 0 {
		return "no releases", nil
	}

	out = append(out, fmt.Sprintf("%d releases", count))
	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
