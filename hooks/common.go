package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var supportedCommands = map[string]bool{
	"git": true, "npm": true, "npx": true, "pnpm": true, "yarn": true, "bun": true,
	"docker": true, "docker-compose": true, "dotnet": true, "kubectl": true, "helm": true, "terraform": true,
	"ansible-playbook": true,
	"cargo": true, "go": true, "tsc": true, "eslint": true, "biome": true,
	"gh": true, "grep": true, "rg": true, "curl": true, "http": true,
	"aws": true, "az": true, "gcloud": true, "mvn": true, "gradle": true, "gradlew": true,
	"ng": true, "nx": true, "pytest": true, "pip": true, "pip3": true, "uv": true,
	"mypy": true, "ruff": true, "flake8": true, "pylint": true,
	"bundle": true, "bundler": true, "rspec": true, "rubocop": true,
	"composer": true, "make": true, "cmake": true,
	"gcc": true, "g++": true, "cc": true, "c++": true, "clang": true, "clang++": true,
	"ping": true, "ps": true, "ss": true, "netstat": true, "df": true, "du": true,
	"cat": true, "tail": true, "less": true, "more": true,
	"find": true, "node": true, "node16": true, "node18": true, "node20": true, "node22": true,
	"acli": true,
}

// shellBuiltins are commands that should never be wrapped.
var shellBuiltins = []string{
	"cd ", "export ", "source ", "echo ", "printf ", "set ", "unset ", "alias ", "eval ",
}

// pipeRedirectOperators make wrapping ambiguous — skip the entire command.
var pipeRedirectOperators = []string{" | ", " > ", " >> ", " < "}

// logicalSeparators chain independent commands — split and wrap each segment.
var logicalSeparators = []string{" && ", " || ", " ; "}

// quoteState tracks parser position relative to shell quoting.
type quoteState int

const (
	quoteNone   quoteState = iota
	quoteSingle            // inside '...'
	quoteDouble            // inside "..."
)

// indexOutsideQuotes returns the index of the first occurrence of needle that appears
// outside of any quoted region. Returns -1 if not found.
func indexOutsideQuotes(s, needle string) int {
	state := quoteNone
	for i := 0; i < len(s); {
		ch := s[i]
		switch state {
		case quoteNone:
			if ch == '\'' {
				state = quoteSingle
				i++
			} else if ch == '"' {
				state = quoteDouble
				i++
			} else if strings.HasPrefix(s[i:], needle) {
				return i
			} else {
				i++
			}
		case quoteSingle:
			if ch == '\'' {
				state = quoteNone
			}
			i++
		case quoteDouble:
			if ch == '\\' && i+1 < len(s) {
				i += 2
			} else if ch == '"' {
				state = quoteNone
				i++
			} else {
				i++
			}
		}
	}
	return -1
}

func containsOutsideQuotes(s, needle string) bool {
	return indexOutsideQuotes(s, needle) != -1
}

// WrapCommand determines if a command should be wrapped with chop and returns
// the modified command and a boolean indicating if it was modified.
func WrapCommand(command string) (string, bool) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", false
	}

	if strings.HasPrefix(command, "chop ") || strings.HasPrefix(command, ". ") {
		return command, false
	}

	for _, op := range logicalSeparators {
		if containsOutsideQuotes(command, op) {
			return wrapCompoundGeneric(command)
		}
	}

	for _, prefix := range shellBuiltins {
		if strings.HasPrefix(command, prefix) {
			return command, false
		}
	}

	for _, op := range pipeRedirectOperators {
		if containsOutsideQuotes(command, op) {
			return command, false
		}
	}

	if shouldWrap(command) {
		return "chop " + command, true
	}

	return command, false
}

func shouldWrap(command string) bool {
	if strings.HasPrefix(command, "chop ") || strings.HasPrefix(command, ". ") {
		return false
	}
	for _, prefix := range shellBuiltins {
		if strings.HasPrefix(command, prefix) {
			return false
		}
	}
	baseCmd := command

	state := quoteNone
	endIdx := len(command)
	for i := 0; i < len(command); i++ {
		ch := command[i]
		switch state {
		case quoteNone:
			if ch == '\'' {
				state = quoteSingle
			} else if ch == '"' {
				state = quoteDouble
			} else if ch == ' ' {
				endIdx = i
				goto foundEnd
			}
		case quoteSingle:
			if ch == '\'' {
				state = quoteNone
			}
		case quoteDouble:
			if ch == '\\' && i+1 < len(command) {
				i++
			} else if ch == '"' {
				state = quoteNone
			}
		}
	}
foundEnd:
	baseCmd = command[:endIdx]

	if len(baseCmd) >= 2 && ((baseCmd[0] == '"' && baseCmd[len(baseCmd)-1] == '"') || (baseCmd[0] == '\'' && baseCmd[len(baseCmd)-1] == '\'')) {
		baseCmd = baseCmd[1 : len(baseCmd)-1]
	}
	if lastSlash := strings.LastIndexAny(baseCmd, `/\`); lastSlash != -1 {
		baseCmd = baseCmd[lastSlash+1:]
	}
	if strings.HasSuffix(strings.ToLower(baseCmd), ".exe") {
		baseCmd = baseCmd[:len(baseCmd)-4]
	}
	return supportedCommands[baseCmd]
}

func splitLogical(command string) (segments []string, operators []string) {
	rest := command
	for {
		earliest := -1
		earliestOp := ""
		for _, op := range logicalSeparators {
			if idx := indexOutsideQuotes(rest, op); idx != -1 && (earliest == -1 || idx < earliest) {
				earliest = idx
				earliestOp = op
			}
		}
		if earliest == -1 {
			segments = append(segments, rest)
			break
		}
		segments = append(segments, rest[:earliest])
		operators = append(operators, earliestOp)
		rest = rest[earliest+len(earliestOp):]
	}
	return
}

func wrapCompoundGeneric(command string) (string, bool) {
	segments, operators := splitLogical(command)
	modified := false
	result := make([]string, len(segments))
	for i, seg := range segments {
		seg = strings.TrimSpace(seg)
		if shouldWrap(seg) {
			result[i] = "chop " + seg
			modified = true
		} else {
			result[i] = seg
		}
	}
	if !modified {
		return command, false
	}
	var sb strings.Builder
	for i, seg := range result {
		if i > 0 {
			sb.WriteString(operators[i-1])
		}
		sb.WriteString(seg)
	}
	return sb.String(), true
}

func auditLog(original, rewritten string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, ".local", "share", "chop")
	os.MkdirAll(dir, 0o700)
	path := filepath.Join(dir, "hook-audit.log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	ts := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "%s  rewrite  %s -> %s\n", ts, original, rewritten)
}

func AuditLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "chop", "hook-audit.log"), nil
}
