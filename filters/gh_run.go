package filters

import (
	"fmt"
	"strings"
)

func getGhRunFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "list":
		return filterGhRunList
	case "view":
		return filterGhRunView
	default:
		return nil
	}
}

func filterGhRunList(raw string) (string, error) {
	raw = stripAnsi(strings.TrimSpace(raw))
	if raw == "" {
		return "no workflow runs", nil
	}

	lines := strings.Split(raw, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// gh run list outputs tab-separated:
		// STATUS\tCONCLUSION\tNAME\tWORKFLOW\tBRANCH\tEVENT\tID\tELAPSED\tAGE
		parts := strings.Split(line, "\t")
		if len(parts) >= 7 {
			status := strings.TrimSpace(parts[0])
			conclusion := strings.TrimSpace(parts[1])
			name := strings.TrimSpace(parts[2])
			workflow := strings.TrimSpace(parts[3])
			branch := strings.TrimSpace(parts[4])
			id := strings.TrimSpace(parts[6])
			elapsed := ""
			if len(parts) >= 8 {
				elapsed = strings.TrimSpace(parts[7])
			}

			displayStatus := status
			if conclusion != "" {
				displayStatus = conclusion
			}

			_ = name // workflow name often duplicates workflow
			entry := fmt.Sprintf("%s %s %s (%s)", id, displayStatus, workflow, branch)
			if elapsed != "" {
				entry += " " + elapsed
			}
			out = append(out, entry)
		} else {
			out = append(out, line)
		}
	}

	return strings.Join(out, "\n"), nil
}

func filterGhRunView(raw string) (string, error) {
	raw = stripAnsi(strings.TrimSpace(raw))
	if raw == "" {
		return "no run data", nil
	}

	lines := strings.Split(raw, "\n")
	var (
		status, workflow, conclusion string
		failedSteps                  []string
	)

	inJobs := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		// Detect indentation to distinguish top-level fields from nested job/step fields
		isIndented := len(line) > 0 && (line[0] == ' ' || line[0] == '\t')

		if !isIndented && strings.HasPrefix(lower, "status:") {
			status = strings.TrimSpace(trimmed[7:])
		} else if !isIndented && strings.HasPrefix(lower, "workflow:") {
			workflow = strings.TrimSpace(trimmed[9:])
		} else if !isIndented && strings.HasPrefix(lower, "conclusion:") {
			conclusion = strings.TrimSpace(trimmed[11:])
		} else if strings.HasPrefix(lower, "jobs:") || strings.HasPrefix(lower, "steps:") {
			inJobs = true
		} else if inJobs {
			// Look for failed steps/jobs
			if strings.Contains(lower, "fail") || strings.Contains(lower, "error") || strings.Contains(lower, "cancelled") || strings.Contains(lower, "timed out") {
				failedSteps = append(failedSteps, trimmed)
			}
		}
	}

	var out []string
	if workflow != "" {
		out = append(out, "workflow: "+workflow)
	}
	displayStatus := status
	if conclusion != "" {
		displayStatus = conclusion
	}
	if displayStatus != "" {
		out = append(out, "status: "+displayStatus)
	}
	if len(failedSteps) > 0 {
		out = append(out, "FAILED STEPS:")
		for _, s := range failedSteps {
			out = append(out, "  "+s)
		}
	}

	if len(out) == 0 {
		return raw, nil
	}
	return strings.Join(out, "\n"), nil
}
