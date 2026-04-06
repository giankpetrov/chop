package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reCondaSpecs    = regexp.MustCompile(`(?i)^\s{4}-\s+\S+`)
	reCondaTotal    = regexp.MustCompile(`(?i)Total:\s+(\S+\s+[MKG]B)`)
	reCondaNewCount = regexp.MustCompile(`(?i)The following NEW packages will be INSTALLED:`)
	reCondaDone     = regexp.MustCompile(`(?i)^(Preparing|Verifying|Executing) transaction:\s*(done|failed)`)
	reCondaErr      = regexp.MustCompile(`(?i)^(PackagesNotFoundError|CondaError|ERROR|error):`)
)

func filterCondaInstall(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikeCondaInstallOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")

	var specs []string
	var totalSize string
	newPkgCount := 0
	countingNew := false
	var finalStatus string
	var errors []string

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			countingNew = false
			continue
		}

		if reCondaNewCount.MatchString(t) {
			countingNew = true
			continue
		}
		if countingNew {
			// New package lines start with "  <name>" (two-space indent + name)
			if strings.HasPrefix(line, "  ") && !strings.HasPrefix(t, "-") && !strings.HasPrefix(t, "=") {
				newPkgCount++
				continue
			} else if !strings.HasPrefix(line, " ") {
				countingNew = false
			}
		}

		if reCondaSpecs.MatchString(line) {
			specs = append(specs, t[2:]) // strip leading "- "
			continue
		}

		if m := reCondaTotal.FindStringSubmatch(t); m != nil {
			totalSize = m[1]
			continue
		}

		if m := reCondaDone.FindStringSubmatch(t); m != nil {
			finalStatus = fmt.Sprintf("%s transaction: %s", m[1], m[2])
			continue
		}

		if reCondaErr.MatchString(t) {
			errors = append(errors, t)
			continue
		}
	}

	var out strings.Builder

	if len(specs) > 0 {
		fmt.Fprintf(&out, "specs: %s", strings.Join(specs, ", "))
	}

	if totalSize != "" {
		if out.Len() > 0 {
			out.WriteString("\n")
		}
		fmt.Fprintf(&out, "download: %s", totalSize)
	}

	if newPkgCount > 0 {
		if out.Len() > 0 {
			out.WriteString("\n")
		}
		fmt.Fprintf(&out, "%d new packages installed", newPkgCount)
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

func filterCondaList(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikeCondaListOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")
	count := 0
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		count++
	}

	if count <= 15 {
		return raw, nil
	}

	result := fmt.Sprintf("%d packages in environment", count)
	return outputSanityCheck(raw, result), nil
}

func getCondaFilter(args []string) FilterFunc {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "install", "update", "create":
		return filterCondaInstall
	case "list":
		return filterCondaList
	default:
		return nil
	}
}
