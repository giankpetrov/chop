package filters

import (
	"regexp"
	"strings"
)

var (
	reComposerOps    = regexp.MustCompile(`(?i)^Package operations:\s*(.+)`)
	reComposerGen    = regexp.MustCompile(`(?i)^Generating\s+`)
	reComposerErr    = regexp.MustCompile(`(?i)^\s*\[.+?Exception\]`)
	reComposerWarn   = regexp.MustCompile(`(?i)^Warning:`)
)

func filterComposerInstall(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikeComposerOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(raw, "\n")

	var opsLine string
	var genLine string
	var errors []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if m := reComposerOps.FindStringSubmatch(trimmed); m != nil {
			opsLine = m[1]
			continue
		}
		if reComposerGen.MatchString(trimmed) {
			genLine = trimmed
			continue
		}
		if reComposerErr.MatchString(trimmed) {
			errors = append(errors, trimmed)
			continue
		}
	}

	var out strings.Builder
	if opsLine != "" {
		out.WriteString(opsLine)
	}
	if genLine != "" {
		if out.Len() > 0 {
			out.WriteString("\n")
		}
		out.WriteString(genLine)
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
