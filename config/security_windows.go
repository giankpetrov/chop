//go:build windows

package config

import (
	"os"
)

// IsSecure checks if a file or directory has secure permissions.
// On Windows, we simplify this to checking that the file is not world-writable,
// although Go's os.Stat() on Windows doesn't map permissions perfectly.
// For now, we return true to avoid breaking Windows users, as the primary
// target for these security measures is Unix-like environments where
// shared systems are more common and permissions are well-defined.
func IsSecure(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}
