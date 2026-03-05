package filters

import (
	"regexp"
	"strings"
)

var (
	reBunInstalled = regexp.MustCompile(`(?i)^Installed\s+\d+\s+packages?\s+in`)
	reBunResolved  = regexp.MustCompile(`(?i)^Resolved\s+\d+\s+packages?`)
)

func filterBunInstall(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikeBunInstallOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")

	var installedLine string
	var errors []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if reBunInstalled.MatchString(trimmed) {
			installedLine = trimmed
			continue
		}
		if strings.HasPrefix(strings.ToLower(trimmed), "error") {
			errors = append(errors, trimmed)
			continue
		}
	}

	var out strings.Builder
	if installedLine != "" {
		out.WriteString(installedLine)
	}
	for _, e := range errors {
		out.WriteString("\n")
		out.WriteString(e)
	}

	result := strings.TrimSpace(out.String())
	if result == "" {
		return raw, nil
	}
	return outputSanityCheck(raw, result), nil
}
