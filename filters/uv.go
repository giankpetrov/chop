package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reUvInstalled = regexp.MustCompile(`(?i)^Installed\s+(\d+)\s+packages?\s+in\s+(.+)`)
	reUvResolved  = regexp.MustCompile(`(?i)^Resolved\s+(\d+)\s+packages?`)
	reUvErr       = regexp.MustCompile(`(?i)^error:`)
)

func filterUvInstall(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikeUvInstallOutput(trimmed) {
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

		if m := reUvInstalled.FindStringSubmatch(trimmed); m != nil {
			installedLine = fmt.Sprintf("installed %s packages in %s", m[1], m[2])
			continue
		}
		if reUvErr.MatchString(trimmed) {
			errors = append(errors, trimmed)
			continue
		}
	}

	var out strings.Builder
	if installedLine != "" {
		out.WriteString(installedLine)
	}
	for _, e := range errors {
		fmt.Fprintf(&out, "\n%s", e)
	}

	result := strings.TrimSpace(out.String())
	if result == "" {
		return raw, nil
	}
	return outputSanityCheck(raw, result), nil
}
