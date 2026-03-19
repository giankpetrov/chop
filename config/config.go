package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds user preferences loaded from ~/.config/chop/config.yml.
type Config struct {
	Disabled []string
}

// Path returns the config file path.
func Path() string {
	return filepath.Join(ConfigDir(), "config.yml")
}

// Load reads the config file and returns a Config.
// Returns defaults if the file doesn't exist or can't be parsed.
func Load() Config {
	return LoadFrom(Path())
}

// LoadFrom reads config from a specific path. Exported for testing.
func LoadFrom(path string) Config {
	cfg := Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	return parse(string(data))
}

// IsDisabled returns true if the given base command (and optional subcommand)
// matches an entry in the disabled list.
//
// Matching rules:
//   - "git" disables all git subcommands
//   - "git diff" disables only "git diff"
//
// Call with: IsDisabled("git", "diff") or IsDisabled("git")
func (c Config) IsDisabled(command string, args ...string) bool {
	fullCmd := command
	if len(args) > 0 && args[0] != "" {
		fullCmd = command + " " + args[0]
	}
	for _, d := range c.Disabled {
		// Exact match on full "cmd subcmd" or base "cmd"
		if strings.EqualFold(d, fullCmd) || strings.EqualFold(d, command) {
			return true
		}
	}
	return false
}

// LoadWithLocal loads the global config, then overlays a local .chop.yml
// from the given directory (if it exists). The local disabled list fully
// replaces the global one.
func LoadWithLocal(cwd string) Config {
	cfg := Load()
	if cwd == "" {
		return cfg
	}
	localPath := filepath.Join(cwd, ".chop.yml")
	if _, err := os.Stat(localPath); err != nil {
		return cfg
	}
	local := LoadFrom(localPath)
	// Local disabled list overrides global entirely
	cfg.Disabled = local.Disabled
	return cfg
}

// Validate checks a config file for structural issues and returns a list of
// human-readable error strings. Returns nil if the file is valid.
func Validate(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("cannot read file: %v", err)}
	}

	var errs []string
	knownKeys := map[string]bool{"disabled": true}

	for i, line := range strings.Split(string(data), "\n") {
		// Strip comments
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, value, ok := parseKV(line)
		if !ok {
			errs = append(errs, fmt.Sprintf("line %d: invalid syntax %q", i+1, line))
			continue
		}
		if !knownKeys[key] {
			errs = append(errs, fmt.Sprintf("line %d: unknown key %q", i+1, key))
		}
		if key == "disabled" {
			items := parseList(value)
			for _, item := range items {
				if item == "" {
					errs = append(errs, fmt.Sprintf("line %d: empty entry in disabled list", i+1))
				}
			}
		}
	}
	return errs
}

// parse does simple line-by-line parsing of the config YAML.
func parse(content string) Config {
	cfg := Config{}

	for _, line := range strings.Split(content, "\n") {
		// Strip comments
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		key, value, ok := parseKV(line)
		if !ok {
			continue
		}

		switch key {
		case "disabled":
			cfg.Disabled = parseList(value)
		}
	}

	return cfg
}

// parseKV splits "key: value" into key and value.
func parseKV(line string) (string, string, bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}
	key := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+1:])
	return key, value, true
}

// parseList parses an inline YAML list like "[git, docker]" or "[]".
func parseList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "[]" || value == "" {
		return nil
	}

	// Strip brackets
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")

	var items []string
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		// Strip quotes if present
		item = strings.Trim(item, "\"'")
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}
