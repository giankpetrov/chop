package config

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	configDirOnce sync.Once
	configDirVal  string

	dataDirOnce sync.Once
	dataDirVal  string
)

// ConfigDir returns the platform-specific directory for configuration files.
// On Unix: $XDG_CONFIG_HOME/chop or ~/.config/chop
// On Windows: %AppData%\chop
func ConfigDir() string {
	configDirOnce.Do(func() {
		dir, err := os.UserConfigDir()
		if err != nil {
			home, _ := os.UserHomeDir()
			if runtime.GOOS == "windows" {
				configDirVal = filepath.Join(home, "AppData", "Roaming", "chop")
				return
			}
			configDirVal = filepath.Join(home, ".config", "chop")
			return
		}
		configDirVal = filepath.Join(dir, "chop")
	})
	return configDirVal
}

// DataDir returns the platform-specific directory for state, logs, and database.
// On Unix: $XDG_DATA_HOME/chop or ~/.local/share/chop
// On Windows: %LocalAppData%\chop
func DataDir() string {
	dataDirOnce.Do(func() {
		if runtime.GOOS == "windows" {
			dir, err := os.UserCacheDir()
			if err == nil {
				dataDirVal = filepath.Join(dir, "chop")
				return
			}
			home, _ := os.UserHomeDir()
			dataDirVal = filepath.Join(home, "AppData", "Local", "chop")
			return
		}

		// Unix-like
		if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
			dataDirVal = filepath.Join(dir, "chop")
			return
		}
		home, _ := os.UserHomeDir()
		dataDirVal = filepath.Join(home, ".local", "share", "chop")
	})
	return dataDirVal
}
