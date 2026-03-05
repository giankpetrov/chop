package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	rePipInstalled = regexp.MustCompile(`(?i)^Successfully installed\s+(.+)`)
	rePipUpToDate  = regexp.MustCompile(`(?i)already satisfied`)
	rePipWarn      = regexp.MustCompile(`(?i)^WARNING:`)
	rePipErr       = regexp.MustCompile(`(?i)^ERROR:`)
)

func filterPipInstall(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikePipInstallOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")

	var installedLine string
	var warnings []string
	var errors []string
	upToDate := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if m := rePipInstalled.FindStringSubmatch(trimmed); m != nil {
			installedLine = m[1]
			continue
		}
		if rePipUpToDate.MatchString(trimmed) {
			upToDate++
			continue
		}
		if rePipWarn.MatchString(trimmed) {
			warnings = append(warnings, trimmed)
			continue
		}
		if rePipErr.MatchString(trimmed) {
			errors = append(errors, trimmed)
			continue
		}
	}

	var out strings.Builder

	if installedLine != "" {
		pkgs := strings.Fields(installedLine)
		fmt.Fprintf(&out, "installed %d packages: %s", len(pkgs), installedLine)
	} else if upToDate > 0 {
		fmt.Fprintf(&out, "%d packages already up to date", upToDate)
	}

	for _, e := range errors {
		fmt.Fprintf(&out, "\n%s", e)
	}
	if len(warnings) > 0 {
		fmt.Fprintf(&out, "\n%d warnings", len(warnings))
	}

	result := strings.TrimSpace(out.String())
	if result == "" {
		return raw, nil
	}
	return outputSanityCheck(raw, result), nil
}

func filterPipList(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikePipListOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	// Count non-header, non-separator lines
	count := 0
	var pkgs []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "Package") {
			continue
		}
		count++
		fields := strings.Fields(trimmed)
		if len(fields) >= 1 && count <= 10 {
			pkgs = append(pkgs, fields[0])
		}
	}

	if count == 0 {
		return raw, nil
	}

	var out strings.Builder
	fmt.Fprintf(&out, "%d packages installed", count)
	if len(pkgs) > 0 {
		fmt.Fprintf(&out, ": %s", strings.Join(pkgs, ", "))
		if count > 10 {
			fmt.Fprintf(&out, " ... and %d more", count-10)
		}
	}

	result := out.String()
	return outputSanityCheck(raw, result), nil
}
