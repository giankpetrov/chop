//go:build windows

package config

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// IsSecure checks if the given path is under the current user's profile
// directories (%APPDATA%, %LOCALAPPDATA%, or home). On Windows, files under
// these directories are owned and controlled by the current user, providing
// equivalent protection to the Unix ownership + permission checks.
func IsSecure(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	absPath = filepath.Clean(absPath)

	u, err := user.Current()
	if err != nil {
		return false
	}

	candidates := []string{
		u.HomeDir,
		os.Getenv("APPDATA"),
		os.Getenv("LOCALAPPDATA"),
	}

	for _, dir := range candidates {
		if dir == "" {
			continue
		}
		cleanDir := filepath.Clean(dir)
		// Case-insensitive comparison — Windows paths are case-insensitive.
		if strings.EqualFold(absPath, cleanDir) ||
			strings.HasPrefix(
				strings.ToLower(absPath),
				strings.ToLower(cleanDir+string(filepath.Separator)),
			) {
			return true
		}
	}

	return false
}
