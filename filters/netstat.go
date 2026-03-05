package filters

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var reNetstatPort = regexp.MustCompile(`:(\d+)\s`)

func filterNetstat(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeNetstatOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	if len(lines) < 20 {
		return raw, nil
	}

	type connGroup struct {
		port  string
		count int
	}

	var listenEntries []string
	estabPorts := make(map[string]int)
	total := 0

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		state := ""
		localAddr := ""
		// ss format: State Recv-Q Send-Q Local:Port Peer:Port
		// netstat format: Proto Recv-Q Send-Q Local:Port Foreign:Port State
		for _, f := range fields {
			switch strings.ToUpper(f) {
			case "LISTEN", "ESTAB", "ESTABLISHED", "TIME_WAIT", "CLOSE_WAIT", "SYN_SENT":
				state = strings.ToUpper(f)
			}
		}

		if state == "" {
			continue
		}
		total++

		// Find local address (contains :port)
		for _, f := range fields {
			if strings.Contains(f, ":") && !strings.HasPrefix(f, "LISTEN") {
				localAddr = f
				break
			}
		}

		switch state {
		case "LISTEN":
			listenEntries = append(listenEntries, localAddr)
		case "ESTAB", "ESTABLISHED":
			if m := reNetstatPort.FindStringSubmatch(localAddr + " "); m != nil {
				estabPorts[":" + m[1]]++
			} else {
				estabPorts[localAddr]++
			}
		}
	}

	if total == 0 {
		return raw, nil
	}

	var out []string
	if len(listenEntries) > 0 {
		out = append(out, "LISTEN: "+strings.Join(listenEntries, ", "))
	}

	if len(estabPorts) > 0 {
		var eParts []string
		ports := make([]string, 0, len(estabPorts))
		for p := range estabPorts {
			ports = append(ports, p)
		}
		sort.Strings(ports)
		for _, p := range ports {
			eParts = append(eParts, fmt.Sprintf("%s (%d)", p, estabPorts[p]))
		}
		out = append(out, "ESTAB: "+strings.Join(eParts, ", "))
	}

	out = append(out, fmt.Sprintf("%d total connections", total))

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
