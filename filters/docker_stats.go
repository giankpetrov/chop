package filters

import (
	"fmt"
	"strings"
)

func filterDockerStats(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeDockerStatsOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")
	if len(lines) < 2 {
		return raw, nil
	}

	header := lines[0]
	nameIdx := strings.Index(header, "NAME")
	cpuIdx := strings.Index(header, "CPU %")
	memIdx := strings.Index(header, "MEM USAGE")
	pidsIdx := strings.Index(header, "PIDS")

	if nameIdx == -1 || cpuIdx == -1 || memIdx == -1 {
		return raw, nil
	}

	memPercIdx := strings.Index(header, "MEM %")
	memBound := memPercIdx
	if memBound == -1 {
		memBound = pidsIdx
	}
	if memBound == -1 {
		memBound = len(header)
	}

	var out []string
	count := 0
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		count++

		name := extractColumn(line, nameIdx, cpuIdx)
		cpu := extractColumn(line, cpuIdx, memIdx)
		memRaw := extractColumn(line, memIdx, memBound)
		mem := memRaw
		if slashIdx := strings.Index(memRaw, "/"); slashIdx != -1 {
			mem = strings.TrimSpace(memRaw[:slashIdx])
		}

		pids := ""
		if pidsIdx != -1 {
			pids = extractColumn(line, pidsIdx, len(line))
		}

		entry := fmt.Sprintf("%s %s %s", name, cpu, mem)
		if pids != "" {
			entry += fmt.Sprintf(" pids:%s", pids)
		}
		out = append(out, entry)
	}

	if count == 0 {
		return "no containers", nil
	}

	out = append(out, fmt.Sprintf("%d containers", count))
	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
