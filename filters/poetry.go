package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	rePoetryOps     = regexp.MustCompile(`(?i)^Package operations:\s*(.+)`)
	rePoetryUpdate  = regexp.MustCompile(`(?i)^\s*[•·\-]\s+Updating\s+\S+\s+\((.+?)\s*->\s*(.+?)\)`)
	rePoetryInstall = regexp.MustCompile(`(?i)^\s*[•·\-]\s+Installing\s+(\S+)\s+\((.+?)\)`)
	rePoetryErr     = regexp.MustCompile(`(?i)^(error|solveprob|could not solve)`)
)

func filterPoetryInstall(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikePoetryInstallOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")

	var opsLine string
	var updates []string
	var installs []string
	var errors []string
	wrotelock := false

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		if m := rePoetryOps.FindStringSubmatch(t); m != nil {
			opsLine = "Package operations: " + m[1]
			continue
		}
		if rePoetryUpdate.MatchString(t) {
			updates = append(updates, t)
			continue
		}
		if rePoetryInstall.MatchString(t) {
			installs = append(installs, t)
			continue
		}
		if rePoetryErr.MatchString(t) {
			errors = append(errors, t)
			continue
		}
		if strings.Contains(t, "Writing lock file") {
			wrotelock = true
			continue
		}
	}

	// Detect "nothing to do" output
	if opsLine == "" && len(updates) == 0 && len(installs) == 0 && len(errors) == 0 {
		if strings.Contains(trimmed, "up to date") || strings.Contains(trimmed, "No dependencies to install") {
			return outputSanityCheck(raw, "up to date"), nil
		}
		return raw, nil
	}

	var out strings.Builder

	if opsLine != "" {
		out.WriteString(opsLine)
	}

	for _, u := range updates {
		if out.Len() > 0 {
			out.WriteString("\n")
		}
		out.WriteString(u)
	}

	const maxInstalls = 2
	if len(installs) > 0 {
		shown := installs
		if len(installs) > maxInstalls {
			shown = installs[:maxInstalls]
		}
		for _, s := range shown {
			if out.Len() > 0 {
				out.WriteString("\n")
			}
			out.WriteString(s)
		}
		if len(installs) > maxInstalls {
			fmt.Fprintf(&out, "\n... and %d more installs", len(installs)-maxInstalls)
		}
	}

	if wrotelock {
		if out.Len() > 0 {
			out.WriteString("\n")
		}
		out.WriteString("Writing lock file")
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

func filterPoetryShow(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikePoetryShowOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	if count <= 10 {
		return raw, nil
	}

	result := fmt.Sprintf("%d packages installed", count)
	return outputSanityCheck(raw, result), nil
}

func getPoetryFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "install", "update", "add", "remove":
		return filterPoetryInstall
	case "show":
		return filterPoetryShow
	default:
		return nil
	}
}
