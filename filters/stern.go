package filters

import (
	"fmt"
	"strings"
)

const sternMaxPatterns = 20

func filterStern(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	// Short output: passthrough
	if len(lines) < 20 {
		return raw, nil
	}

	// Each stern line: <pod> <container> <rest...>
	// We group by the message content (everything after pod and container).
	type podMsg struct {
		message string
		isError bool
	}

	type msgGroup struct {
		pods    map[string]struct{} // distinct pod names
		example string              // last full line seen
		isError bool
		order   int // insertion order
	}

	groups := make(map[string]*msgGroup)
	var groupOrder []string
	totalLines := 0

	for _, line := range lines {
		t := strings.TrimSpace(stripAnsi(line))
		if t == "" {
			continue
		}
		totalLines++

		// Split into fields: pod container [rest...]
		parts := strings.SplitN(t, " ", 3)
		if len(parts) < 3 {
			// Not stern format — treat entire line as message
			msg := t
			if g, ok := groups[msg]; ok {
				g.pods[""] = struct{}{}
				g.example = t
			} else {
				groups[msg] = &msgGroup{
					pods:    map[string]struct{}{"": {}},
					example: t,
					isError: isErrorLine(t),
					order:   len(groupOrder),
				}
				groupOrder = append(groupOrder, msg)
			}
			continue
		}

		pod := parts[0]
		// parts[1] is container, parts[2] is the rest of the log line
		msgBody := parts[2]
		isErr := isErrorLine(msgBody)

		// Use fingerprint for grouping similar messages across pods
		fp := patternFingerprint(msgBody)

		if g, ok := groups[fp]; ok {
			g.pods[pod] = struct{}{}
			g.example = t
			if isErr {
				g.isError = true
			}
		} else {
			groups[fp] = &msgGroup{
				pods:    map[string]struct{}{pod: {}},
				example: t,
				isError: isErr,
				order:   len(groupOrder),
			}
			groupOrder = append(groupOrder, fp)
		}
	}

	// Separate error and normal groups
	var errorGroups []string
	var normalGroups []string

	for _, fp := range groupOrder {
		g := groups[fp]
		if g.isError {
			errorGroups = append(errorGroups, fp)
		} else {
			normalGroups = append(normalGroups, fp)
		}
	}

	// Keep last sternMaxPatterns normal groups
	hidden := 0
	if len(normalGroups) > sternMaxPatterns {
		hidden = len(normalGroups) - sternMaxPatterns
		normalGroups = normalGroups[hidden:]
	}

	var out []string
	if hidden > 0 {
		out = append(out, fmt.Sprintf("(%d earlier message patterns hidden)", hidden))
	}

	formatGroup := func(fp string) string {
		g := groups[fp]
		podCount := len(g.pods)

		// For single pod or non-stern lines: show example unchanged
		if podCount <= 1 {
			return g.example
		}

		// Multiple pods: strip pod name from the example and prefix with pod count
		parts := strings.SplitN(g.example, " ", 3)
		if len(parts) >= 3 {
			return fmt.Sprintf("[%d pods] %s", podCount, parts[2])
		}
		return fmt.Sprintf("[%d pods] %s", podCount, g.example)
	}

	for _, fp := range errorGroups {
		out = append(out, formatGroup(fp))
	}
	for _, fp := range normalGroups {
		out = append(out, formatGroup(fp))
	}

	out = append(out, fmt.Sprintf("--- %d lines total ---", totalLines))

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
