//go:build !windows

package config

import (
	"os"
	"syscall"
)

// IsSecure checks if a file or directory has secure permissions and is owned by the current user.
// Secure means only the owner has read/write (and search for dirs) access.
func IsSecure(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	mode := info.Mode()
	// Check that group and others have no permissions (0077 mask)
	if mode&0077 != 0 {
		return false
	}

	// On Unix, also verify ownership
	stat, ok := info.Sys().(*syscall.Stat_t)
	if ok {
		if int(stat.Uid) != os.Getuid() {
			return false
		}
	}

	return true
}

// SecureFileMode is the recommended permission for configuration files.
const SecureFileMode = 0o600

// SecureDirMode is the recommended permission for configuration directories.
const SecureDirMode = 0o700
