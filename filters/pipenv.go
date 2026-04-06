package filters

import (
	"regexp"
	"strings"
)

var (
	rePipenvVenv    = regexp.MustCompile(`(?i)^Virtualenv location:\s*(.+)`)
	rePipenvLockRev = regexp.MustCompile(`(?i)^Installing dependencies from Pipfile\.lock`)
	rePipenvErr     = regexp.MustCompile(`(?i)^(ERROR|error|Error|✘|Failed)`)
	rePipenvSuccess = regexp.MustCompile(`(?i)^[✔✓]?\s*(Success!|Successfully)`)
)

func filterPipenv(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikePipenvOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")

	var venvLocation string
	installingDeps := false
	var errors []string
	var finalStatus string

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		if m := rePipenvVenv.FindStringSubmatch(t); m != nil {
			venvLocation = strings.TrimSpace(m[1])
			continue
		}

		if rePipenvLockRev.MatchString(t) {
			installingDeps = true
			continue
		}

		if rePipenvErr.MatchString(t) {
			errors = append(errors, t)
			continue
		}

		if rePipenvSuccess.MatchString(t) {
			finalStatus = t
			continue
		}
	}

	var out strings.Builder

	if venvLocation != "" {
		out.WriteString("Virtualenv location: ")
		out.WriteString(venvLocation)
	}

	if installingDeps {
		if out.Len() > 0 {
			out.WriteString("\n")
		}
		out.WriteString("Installing dependencies from Pipfile.lock")
	}

	if finalStatus != "" {
		if out.Len() > 0 {
			out.WriteString("\n")
		}
		out.WriteString(finalStatus)
	}

	for _, e := range errors {
		if out.Len() > 0 {
			out.WriteString("\n")
		}
		out.WriteString(e)
	}

	result := strings.TrimSpace(out.String())
	if result == "" {
		return raw, nil
	}
	return outputSanityCheck(raw, result), nil
}
