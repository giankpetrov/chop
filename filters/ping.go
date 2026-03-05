package filters

import (
	"regexp"
	"strings"
)

var (
	rePingStats = regexp.MustCompile(`(?i)(\d+)\s+packets?\s+transmitted,\s*(\d+)\s+(?:packets?\s+)?received,\s*(\d+%)\s+(?:packet\s+)?loss`)
	rePingRtt   = regexp.MustCompile(`(?i)(?:rtt|round trip).*?=\s*(.+)`)
	rePingHost  = regexp.MustCompile(`(?i)^(?:PING|Pinging)\s+(\S+)\s+(?:\((\S+)\)|[\[\(](\S+)[\]\)])`)
	// Windows: Packets: Sent = 4, Received = 4, Lost = 0 (0% loss)
	rePingWinStats = regexp.MustCompile(`(?i)Packets:.*Sent\s*=\s*(\d+).*Received\s*=\s*(\d+).*Lost\s*=\s*\d+\s*\((\d+%)\s*loss\)`)
	rePingWinRtt   = regexp.MustCompile(`(?i)(?:Minimum|Average|Maximum)\s*=\s*\d+ms`)
)

func filterPing(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikePingOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	var host, ip, statsLine, rttLine string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if m := rePingHost.FindStringSubmatch(trimmed); m != nil {
			host = m[1]
			if m[2] != "" {
				ip = m[2]
			} else if m[3] != "" {
				ip = m[3]
			}
		}

		if m := rePingStats.FindStringSubmatch(trimmed); m != nil {
			statsLine = m[1] + " sent, " + m[2] + " received, " + m[3] + " loss"
		}
		if m := rePingWinStats.FindStringSubmatch(trimmed); m != nil {
			statsLine = m[1] + " sent, " + m[2] + " received, " + m[3] + " loss"
		}

		if m := rePingRtt.FindStringSubmatch(trimmed); m != nil {
			rttLine = "rtt " + strings.TrimSpace(m[1])
		}
	}

	if statsLine == "" {
		return raw, nil
	}

	var out []string
	target := host
	if ip != "" && ip != host {
		target = host + " (" + ip + ")"
	}
	if target != "" {
		out = append(out, target+": "+statsLine)
	} else {
		out = append(out, statsLine)
	}
	if rttLine != "" {
		out = append(out, rttLine)
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
