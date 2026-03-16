package config

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the platform-specific directory for configuration.
// Respects XDG_CONFIG_HOME if set, otherwise uses os.UserConfigDir().
func ConfigDir() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		d, err := os.UserConfigDir()
		if err != nil {
			home, err := os.UserHomeDir()
			if err != nil {
				home = "."
			}
			d = filepath.Join(home, ".config")
		}
		dir = d
	}
	return filepath.Join(dir, "chop")
}

// DataDir returns the platform-specific directory for data files (tracking, logs, etc).
// Respects XDG_DATA_HOME if set, otherwise uses os.UserCacheDir().
func DataDir() string {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		d, err := os.UserCacheDir()
		if err != nil {
			home, err := os.UserHomeDir()
			if err != nil {
				home = "."
			}
			d = filepath.Join(home, ".local", "share")
		}
		dir = d
	}
	return filepath.Join(dir, "chop")
}

// SecureFileMode is used for private files (0o600).
const SecureFileMode = 0o600

// SecureDirMode is used for private directories (0o700).
const SecureDirMode = 0o700
