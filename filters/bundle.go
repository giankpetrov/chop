package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reBundleComplete  = regexp.MustCompile(`(?i)^Bundle complete!.*`)
	reBundleInstalled = regexp.MustCompile(`(?i)^Installing\s+`)
	reBundleErr       = regexp.MustCompile(`(?i)^Bundler::`)
)

func filterBundleInstall(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikeBundleInstallOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")

	var completeLine string
	installed := 0
	var errors []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if reBundleComplete.MatchString(trimmed) {
			completeLine = trimmed
			continue
		}
		if reBundleInstalled.MatchString(trimmed) {
			installed++
			continue
		}
		if reBundleErr.MatchString(trimmed) {
			errors = append(errors, trimmed)
			continue
		}
	}

	var out strings.Builder
	if installed > 0 {
		fmt.Fprintf(&out, "installed %d gems\n", installed)
	}
	if completeLine != "" {
		out.WriteString(completeLine)
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
