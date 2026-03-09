package filters

import (
	"fmt"
	"strings"
)

type gitSection int

const (
	sectionNone gitSection = iota
	sectionStaged
	sectionUnstaged
	sectionUntracked
)

func filterGitStatus(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikeGitStatusOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")

	var staged, unstaged, untracked []string
	section := sectionNone

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Detect section headers
		if strings.HasPrefix(trimmedLine, "Changes to be committed") {
			section = sectionStaged
			continue
		}
		if strings.HasPrefix(trimmedLine, "Changes not staged for commit") {
			section = sectionUnstaged
			continue
		}
		if strings.HasPrefix(trimmedLine, "Untracked files") {
			section = sectionUntracked
			continue
		}

		// Skip hint lines and empty lines
		if strings.HasPrefix(trimmedLine, "(use ") || trimmedLine == "" ||
			strings.HasPrefix(trimmedLine, "On branch") ||
			strings.HasPrefix(trimmedLine, "Your branch") ||
			strings.HasPrefix(trimmedLine, "no changes added") ||
			strings.HasPrefix(trimmedLine, "nothing to commit") {
			continue
		}

		switch section {
		case sectionStaged:
			if f := parseStatusFile(trimmedLine); f != "" {
				staged = append(staged, f)
			}
		case sectionUnstaged:
			if f := parseStatusFile(trimmedLine); f != "" {
				unstaged = append(unstaged, f)
			}
		case sectionUntracked:
			// Untracked files are just indented names (no prefix like "modified:")
			if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "  ") {
				untracked = append(untracked, trimmedLine)
			}
		}
	}

	// Detect clean working tree
	if len(staged) == 0 && len(unstaged) == 0 && len(untracked) == 0 {
		if strings.Contains(raw, "nothing to commit") {
			return "clean", nil
		}
		return raw, nil
	}

	var out strings.Builder

	if len(staged) > 0 {
		fmt.Fprintf(&out, "staged(%d): %s\n", len(staged), strings.Join(staged, ", "))
	}
	if len(unstaged) > 0 {
		fmt.Fprintf(&out, "unstaged(%d): %s\n", len(unstaged), strings.Join(unstaged, ", "))
	}
	if len(untracked) > 0 {
		fmt.Fprintf(&out, "untracked(%d): %s\n", len(untracked), strings.Join(untracked, ", "))
	}

	result := strings.TrimSpace(out.String())
	return outputSanityCheck(raw, result), nil
}

// parseStatusFile extracts the filename from a git status line like
// "modified:   src/app.ts" or "new file:   README.md"
func parseStatusFile(line string) string {
	prefixes := []string{"new file:", "modified:", "deleted:", "renamed:", "copied:", "typechange:"}
	for _, p := range prefixes {
		if strings.HasPrefix(line, p) {
			name := strings.TrimSpace(strings.TrimPrefix(line, p))
			if p == "deleted:" {
				return name + " (deleted)"
			}
			if p == "new file:" {
				return name + " (new)"
			}
			return name
		}
	}
	return ""
}
