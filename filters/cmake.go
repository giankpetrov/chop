package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reCmakeConfigDone = regexp.MustCompile(`(?i)--\s*Configuring done\s*\((.+?)\)`)
	reCmakeGenDone    = regexp.MustCompile(`(?i)--\s*Generating done`)
	reCmakeWritten    = regexp.MustCompile(`(?i)--\s*Build files have been written to:\s*(.+)`)
	reCmakeBuildPct   = regexp.MustCompile(`^\[\s*(\d+)%\]`)
	reCmakeBuiltTgt   = regexp.MustCompile(`(?i)^\[\s*\d+%\]\s+Built target\s+(.+)`)
	reCmakeBuilding   = regexp.MustCompile(`(?i)^\[\s*\d+%\]\s+Building`)
)

func filterCmake(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeCmakeOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	var configTime, outputDir string
	var targets []string
	buildFiles := 0
	var errors []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if m := reCmakeConfigDone.FindStringSubmatch(trimmed); m != nil {
			configTime = m[1]
			continue
		}
		if m := reCmakeWritten.FindStringSubmatch(trimmed); m != nil {
			outputDir = strings.TrimSpace(m[1])
			continue
		}
		if m := reCmakeBuiltTgt.FindStringSubmatch(trimmed); m != nil {
			targets = append(targets, strings.TrimSpace(m[1]))
			continue
		}
		if reCmakeBuilding.MatchString(trimmed) {
			buildFiles++
			continue
		}
		if strings.Contains(strings.ToLower(trimmed), "error") && !strings.HasPrefix(trimmed, "--") {
			errors = append(errors, trimmed)
		}
	}

	var out []string

	for _, e := range errors {
		out = append(out, e)
	}

	if configTime != "" {
		msg := "configured in " + configTime
		if outputDir != "" {
			msg += ", output: " + outputDir
		}
		out = append(out, msg)
	}

	if len(targets) > 0 {
		for _, t := range targets {
			msg := "built target " + t
			if buildFiles > 0 {
				msg += fmt.Sprintf(" (%d files compiled)", buildFiles)
			}
			out = append(out, msg)
		}
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
