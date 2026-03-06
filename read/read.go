package read

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Language represents a detected source language.
type Language struct {
	Name          string
	LineComment   string // e.g., "//"
	BlockStart    string // e.g., "/*"
	BlockEnd      string // e.g., "*/"
	DocPrefix     string // e.g., "///" for Rust/C# doc comments
	PreserveFirst bool   // preserve first line if shebang
}

// DetectLanguage returns the Language for a file extension.
func DetectLanguage(ext string) Language {
	ext = strings.ToLower(ext)
	switch ext {
	case ".go":
		return Language{Name: "Go", LineComment: "//", BlockStart: "/*", BlockEnd: "*/"}
	case ".rs":
		return Language{Name: "Rust", LineComment: "//", BlockStart: "/*", BlockEnd: "*/", DocPrefix: "///"}
	case ".py":
		return Language{Name: "Python", LineComment: "#", BlockStart: `"""`, BlockEnd: `"""`}
	case ".js", ".mjs", ".cjs", ".ts", ".tsx", ".jsx":
		return Language{Name: "JavaScript", LineComment: "//", BlockStart: "/*", BlockEnd: "*/"}
	case ".c", ".h", ".cpp", ".cc", ".hpp":
		return Language{Name: "C", LineComment: "//", BlockStart: "/*", BlockEnd: "*/"}
	case ".cs":
		return Language{Name: "CSharp", LineComment: "//", BlockStart: "/*", BlockEnd: "*/", DocPrefix: "///"}
	case ".java":
		return Language{Name: "Java", LineComment: "//", BlockStart: "/*", BlockEnd: "*/"}
	case ".rb":
		return Language{Name: "Ruby", LineComment: "#", BlockStart: "=begin", BlockEnd: "=end"}
	case ".sh", ".bash", ".zsh":
		return Language{Name: "Shell", LineComment: "#", PreserveFirst: true}
	case ".html", ".htm", ".xml", ".svg", ".xaml":
		return Language{Name: "HTML", BlockStart: "<!--", BlockEnd: "-->"}
	case ".css":
		return Language{Name: "CSS", BlockStart: "/*", BlockEnd: "*/"}
	case ".scss", ".less":
		return Language{Name: "SCSS", LineComment: "//", BlockStart: "/*", BlockEnd: "*/"}
	case ".sql":
		return Language{Name: "SQL", LineComment: "--", BlockStart: "/*", BlockEnd: "*/"}
	case ".yml", ".yaml":
		return Language{Name: "YAML", LineComment: "#"}
	case ".md":
		return Language{Name: "Markdown"}
	default:
		return Language{Name: "Unknown", LineComment: "//", BlockStart: "/*", BlockEnd: "*/"}
	}
}

// FilterMinimal removes comments and collapses blank lines.
// Doc comments are preserved. Shebangs are preserved for shell scripts.
func FilterMinimal(content string, lang Language) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBlock := false
	inDocstring := false
	blankCount := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Preserve shebang on first line
		if i == 0 && lang.PreserveFirst && strings.HasPrefix(trimmed, "#!") {
			result = append(result, strings.TrimRight(line, " \t\r"))
			blankCount = 0
			continue
		}

		// Python docstring handling — keep docstrings in minimal mode
		if lang.Name == "Python" && !inBlock {
			if strings.Contains(trimmed, `"""`) {
				// Check if docstring opens and closes on same line
				count := strings.Count(trimmed, `"""`)
				if inDocstring {
					// Closing docstring
					result = append(result, strings.TrimRight(line, " \t\r"))
					inDocstring = false
					blankCount = 0
					continue
				} else if count >= 2 {
					// Opens and closes on same line
					result = append(result, strings.TrimRight(line, " \t\r"))
					blankCount = 0
					continue
				} else {
					// Opening docstring
					inDocstring = true
					result = append(result, strings.TrimRight(line, " \t\r"))
					blankCount = 0
					continue
				}
			}
			if inDocstring {
				result = append(result, strings.TrimRight(line, " \t\r"))
				blankCount = 0
				continue
			}
		}

		// Block comment handling
		if inBlock {
			if lang.BlockEnd != "" && strings.Contains(trimmed, lang.BlockEnd) {
				inBlock = false
			}
			continue
		}

		if lang.BlockStart != "" && !inBlock && lang.Name != "Python" {
			if strings.HasPrefix(trimmed, lang.BlockStart) {
				// Check if block closes on same line
				if lang.BlockEnd != "" {
					afterStart := trimmed[len(lang.BlockStart):]
					if strings.Contains(afterStart, lang.BlockEnd) {
						// Single-line block comment, skip it
						continue
					}
				}
				inBlock = true
				continue
			}
		}

		// Doc comment — preserve in minimal mode
		if lang.DocPrefix != "" && strings.HasPrefix(trimmed, lang.DocPrefix) {
			result = append(result, strings.TrimRight(line, " \t\r"))
			blankCount = 0
			continue
		}

		// Rust //! doc comments — preserve
		if lang.Name == "Rust" && strings.HasPrefix(trimmed, "//!") {
			result = append(result, strings.TrimRight(line, " \t\r"))
			blankCount = 0
			continue
		}

		// Line comment removal
		if lang.LineComment != "" && strings.HasPrefix(trimmed, lang.LineComment) {
			continue
		}

		// Blank line collapsing: 3+ blank lines -> 1
		if trimmed == "" {
			blankCount++
			if blankCount <= 1 {
				result = append(result, "")
			}
			continue
		}

		blankCount = 0
		// Strip inline block comments (e.g., code /* comment */ more_code)
		cleaned := stripInlineBlockComments(line, lang)
		result = append(result, strings.TrimRight(cleaned, " \t\r"))
	}

	// Trim trailing blank lines
	for len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}

	return strings.Join(result, "\n")
}

// stripInlineBlockComments removes inline /* ... */ style comments from a line.
func stripInlineBlockComments(line string, lang Language) string {
	if lang.BlockStart == "" || lang.BlockEnd == "" {
		return line
	}
	// Only handle /* */ style inline comments (not <!-- --> or """)
	if lang.BlockStart != "/*" {
		return line
	}
	for {
		start := strings.Index(line, "/*")
		if start < 0 {
			break
		}
		end := strings.Index(line[start:], "*/")
		if end < 0 {
			break
		}
		line = line[:start] + line[start+end+2:]
	}
	return line
}

// FilterAggressive removes all comments, blank lines, and import blocks.
func FilterAggressive(content string, lang Language) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBlock := false
	inDocstring := false
	inImportBlock := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Preserve shebang on first line
		if i == 0 && lang.PreserveFirst && strings.HasPrefix(trimmed, "#!") {
			result = append(result, strings.TrimRight(line, " \t\r"))
			continue
		}

		// Python docstring handling — REMOVE in aggressive mode
		if lang.Name == "Python" && !inBlock {
			if strings.Contains(trimmed, `"""`) {
				count := strings.Count(trimmed, `"""`)
				if inDocstring {
					inDocstring = false
					continue
				} else if count >= 2 {
					continue
				} else {
					inDocstring = true
					continue
				}
			}
			if inDocstring {
				continue
			}
		}

		// Block comment handling
		if inBlock {
			if lang.BlockEnd != "" && strings.Contains(trimmed, lang.BlockEnd) {
				inBlock = false
			}
			continue
		}

		if lang.BlockStart != "" && !inBlock && lang.Name != "Python" {
			if strings.HasPrefix(trimmed, lang.BlockStart) {
				if lang.BlockEnd != "" {
					afterStart := trimmed[len(lang.BlockStart):]
					if strings.Contains(afterStart, lang.BlockEnd) {
						continue
					}
				}
				inBlock = true
				continue
			}
		}

		// Remove ALL doc comments in aggressive mode
		if lang.DocPrefix != "" && strings.HasPrefix(trimmed, lang.DocPrefix) {
			continue
		}
		if lang.Name == "Rust" && strings.HasPrefix(trimmed, "//!") {
			continue
		}

		// Line comment removal
		if lang.LineComment != "" && strings.HasPrefix(trimmed, lang.LineComment) {
			continue
		}

		// Remove blank lines
		if trimmed == "" {
			continue
		}

		// Track import blocks: import ( ... )
		if isImportLine(trimmed) {
			if strings.Contains(trimmed, "(") {
				inImportBlock = true
			}
			continue
		}
		if inImportBlock {
			if trimmed == ")" {
				inImportBlock = false
			}
			continue
		}

		cleaned := stripInlineBlockComments(line, lang)
		result = append(result, strings.TrimRight(cleaned, " \t\r"))
	}

	return strings.Join(result, "\n")
}

// isImportLine checks if a line is an import/use/require/include statement.
func isImportLine(trimmed string) bool {
	prefixes := []string{
		"import ",
		"from ",
		"use ",
		"require(",
		"require ",
		"#include ",
		"using ",
		"package ",
		"module ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(trimmed, p) {
			return true
		}
	}
	return false
}

// Run reads a file, applies the filter, and returns the result.
// level: "minimal" or "aggressive"
// maxLines: 0 means no limit
// lineNumbers: prepend line numbers
func Run(filePath string, level string, maxLines int, lineNumbers bool) (raw string, filtered string, err error) {
	ext := filepath.Ext(filePath)
	lang := DetectLanguage(ext)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", fmt.Errorf("read %s: %w", filePath, err)
	}

	raw = string(data)

	switch strings.ToLower(level) {
	case "aggressive":
		filtered = FilterAggressive(raw, lang)
	default:
		filtered = FilterMinimal(raw, lang)
	}

	// Smart truncation
	if maxLines > 0 {
		filtered = truncate(filtered, maxLines)
	}

	// Line numbers
	if lineNumbers {
		filtered = addLineNumbers(filtered)
	}

	return raw, filtered, nil
}

// truncate keeps the first half and last half of lines, omitting the middle.
func truncate(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}

	half := maxLines / 2
	omitted := len(lines) - maxLines

	var result []string
	result = append(result, lines[:half]...)
	result = append(result, fmt.Sprintf("// ... %d lines omitted", omitted))
	result = append(result, lines[len(lines)-half:]...)
	return strings.Join(result, "\n")
}

// addLineNumbers prepends line numbers in %4d | %s format.
func addLineNumbers(content string) string {
	lines := strings.Split(content, "\n")
	numbered := make([]string, len(lines))
	for i, line := range lines {
		numbered[i] = fmt.Sprintf("%4d | %s", i+1, line)
	}
	return strings.Join(numbered, "\n")
}
