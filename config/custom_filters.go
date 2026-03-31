package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// CustomFilter defines a user-configured output filter for a specific command.
type CustomFilter struct {
	Keep    []string `yaml:"keep"`    // regex patterns - only lines matching at least one are kept
	Drop    []string `yaml:"drop"`    // regex patterns - lines matching any are removed
	Head    int      `yaml:"head"`    // keep only first N lines (applied after keep/drop)
	Tail    int      `yaml:"tail"`    // keep only last N lines (applied after keep/drop)
	Exec    string   `yaml:"exec"`    // external script: pipe output through this command
	Trusted bool     `yaml:"-"`       // Internal: true if loaded from a trusted source
}

// CustomFiltersConfig holds the full filters.yml content.
type CustomFiltersConfig struct {
	Filters map[string]CustomFilter `yaml:"filters"`
}

var (
	globalFiltersOnce sync.Once
	globalFilters     map[string]CustomFilter
)

// FiltersConfigPath returns the path to the custom filters file.
func FiltersConfigPath() string {
	return filepath.Join(ConfigDir(), "filters.yml")
}

// LoadCustomFilters reads the custom filters config file.
// Global filters are trusted only if the config file is secure.
// Result is cached for the lifetime of the process.
func LoadCustomFilters() map[string]CustomFilter {
	globalFiltersOnce.Do(func() {
		path := FiltersConfigPath()
		trusted := IsSecure(path)
		globalFilters = loadCustomFiltersWithTrust(path, trusted)
	})
	return globalFilters
}

// LoadCustomFiltersFrom reads custom filters from a specific path as untrusted.
func LoadCustomFiltersFrom(path string) map[string]CustomFilter {
	return loadCustomFiltersWithTrust(path, false)
}

func loadCustomFiltersWithTrust(path string, trusted bool) map[string]CustomFilter {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return parseCustomFiltersWithTrust(data, trusted)
}

// ParseCustomFilters parses YAML bytes into a custom filters map as untrusted.
func ParseCustomFilters(data []byte) map[string]CustomFilter {
	return parseCustomFiltersWithTrust(data, false)
}

func parseCustomFiltersWithTrust(data []byte, trusted bool) map[string]CustomFilter {
	var cfg CustomFiltersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	if cfg.Filters == nil {
		return nil
	}
	normalized := make(map[string]CustomFilter, len(cfg.Filters))
	for k, v := range cfg.Filters {
		// Security: strictly enforce the Trusted flag by completely rejecting exec
		// directives from local .chop-filters.yml files
		if !trusted && v.Exec != "" {
			v.Exec = ""
		}
		v.Trusted = trusted
		normalized[strings.ToLower(k)] = v
	}
	return normalized
}

// ValidateFilters checks a filters.yml file for issues: invalid YAML, bad regex
// patterns, and missing exec scripts. Returns nil if everything looks good.
func ValidateFilters(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("cannot read file: %v", err)}
	}

	var cfg CustomFiltersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return []string{fmt.Sprintf("invalid YAML: %v", err)}
	}

	var errs []string
	for cmd, f := range cfg.Filters {
		for _, p := range f.Keep {
			if _, err := regexp.Compile(p); err != nil {
				errs = append(errs, fmt.Sprintf("%q keep pattern %q: %v", cmd, p, err))
			}
		}
		for _, p := range f.Drop {
			if _, err := regexp.Compile(p); err != nil {
				errs = append(errs, fmt.Sprintf("%q drop pattern %q: %v", cmd, p, err))
			}
		}
		if f.Exec != "" {
			// Expand ~ for the existence check
			execPath := f.Exec
			if strings.HasPrefix(execPath, "~/") || strings.HasPrefix(execPath, "~\\") {
				if home, err := os.UserHomeDir(); err == nil {
					execPath = filepath.Join(home, execPath[2:])
				}
			}
			if _, err := os.Stat(execPath); err != nil {
				errs = append(errs, fmt.Sprintf("%q exec script %q: not found", cmd, f.Exec))
			}
		}
	}
	return errs
}

// LookupCustomFilter finds a custom filter for the given command and args.
// It checks "command subcommand" first, then falls back to "command".
// Keys are stored lowercased (see parseCustomFiltersWithTrust), so lookups
// are O(1) map accesses.
func LookupCustomFilter(filters map[string]CustomFilter, command string, args []string) *CustomFilter {
	if len(filters) == 0 {
		return nil
	}

	// Try "command subcommand" match first
	if len(args) > 0 {
		if f, ok := filters[strings.ToLower(command+" "+args[0])]; ok {
			return &f
		}
	}

	// Try base command match
	if f, ok := filters[strings.ToLower(command)]; ok {
		return &f
	}

	return nil
}

// LoadCustomFiltersWithLocal loads global custom filters, then overlays
// a local .chop-filters.yml from the given directory (if it exists).
// Global filters are trusted; local ones are NOT.
// Local filters are merged on top of global ones (local wins on conflict).
func LoadCustomFiltersWithLocal(cwd string) map[string]CustomFilter {
	global := LoadCustomFilters()

	if cwd == "" {
		return global
	}

	localPath := filepath.Join(cwd, ".chop-filters.yml")
	// Local filters are untrusted
	local := loadCustomFiltersWithTrust(localPath, false)

	if len(local) == 0 {
		return global
	}

	// Merge: local overrides global. Copy global first to avoid mutating the cache.
	merged := make(map[string]CustomFilter, len(global)+len(local))
	for k, v := range global {
		merged[k] = v
	}
	for k, v := range local {
		merged[k] = v
	}
	return merged
}
