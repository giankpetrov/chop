package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	rePnpmPackages = regexp.MustCompile(`(?i)^Packages:\s*\+(\d+)`)
	rePnpmDone     = regexp.MustCompile(`(?i)^Progress:.*done`)
	rePnpmWarn     = regexp.MustCompile(`(?i)^WARN|warn`)
	rePnpmErr      = regexp.MustCompile(`(?i)^ERR|ERROR`)
)

func filterPnpmInstall(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikePnpmInstallOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")

	var packageCount string
	var warnings int
	var errors []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if m := rePnpmPackages.FindStringSubmatch(trimmed); m != nil {
			packageCount = m[1]
			continue
		}
		if rePnpmErr.MatchString(trimmed) {
			errors = append(errors, trimmed)
			continue
		}
		if rePnpmWarn.MatchString(trimmed) {
			warnings++
			continue
		}
	}

	var out strings.Builder
	if packageCount != "" {
		fmt.Fprintf(&out, "added %s packages", packageCount)
	}

	for _, e := range errors {
		fmt.Fprintf(&out, "\n%s", e)
	}
	if warnings > 0 {
		fmt.Fprintf(&out, "\n%d warnings", warnings)
	}

	result := strings.TrimSpace(out.String())
	if result == "" {
		return raw, nil
	}
	return outputSanityCheck(raw, result), nil
}
