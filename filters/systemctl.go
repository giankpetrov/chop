package filters

import (
	"strings"
)

func filterSystemctlStatus(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeSystemctlStatus(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")
	var out []string
	var serviceLine string
	var activeLine string
	var pidLine string
	var stats []string
	var logLines []string

	inLogs := false

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		// Service name line (usually starts with ●, ○, or ×)
		if serviceLine == "" && (strings.HasPrefix(t, "●") || strings.HasPrefix(t, "○") || strings.HasPrefix(t, "×")) {
			serviceLine = t
			// Clean up the bullet points but keep the text
			serviceLine = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(serviceLine, "●"), "○"), "×"))
			continue
		} else if serviceLine == "" && (strings.Contains(t, ".service -") || strings.Contains(t, ".timer -") || strings.Contains(t, ".socket -")) {
			serviceLine = t
			continue
		}

		if strings.HasPrefix(t, "Active:") {
			activeLine = t
			continue
		}
		if strings.HasPrefix(t, "Main PID:") {
			pidLine = t
			continue
		}
		if strings.HasPrefix(t, "Memory:") {
			stats = append(stats, t)
			continue
		}
		if strings.HasPrefix(t, "CPU:") {
			stats = append(stats, t)
			continue
		}
		if strings.HasPrefix(t, "Tasks:") {
			stats = append(stats, t)
			continue
		}

		// Detect start of logs (usually lines starting with a timestamp or month)
		if !inLogs && (strings.Contains(t, " systemd[") || (len(t) > 15 && (t[3] == ' ' || t[3] == '-'))) {
			inLogs = true
		}

		if inLogs {
			logLines = append(logLines, t)
		}
	}

	if serviceLine != "" {
		out = append(out, serviceLine)
	}
	if activeLine != "" {
		out = append(out, activeLine)
	}
	if pidLine != "" {
		out = append(out, pidLine)
	}
	if len(stats) > 0 {
		out = append(out, strings.Join(stats, ", "))
	}

	// Keep last 5 log lines if they exist
	if len(logLines) > 0 {
		out = append(out, "") // separator
		showLogs := logLines
		if len(showLogs) > 5 {
			showLogs = logLines[len(logLines)-5:]
		}
		out = append(out, showLogs...)
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

func filterSystemctlListUnits(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	// Use table compressor for list-units as it's already quite good
	return compressTable(strings.Split(trimmed, "\n")), nil
}
