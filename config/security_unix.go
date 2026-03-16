//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"syscall"
)

// IsSecure checks if the given path is secure (owned by current user or root,
// and not world-writable). Parent directories are also checked recursively up to the root.
func IsSecure(path string) bool {
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}

	uid := os.Getuid()
	// Must be owned by current user or root.
	if stat.Uid != uint32(uid) && stat.Uid != 0 {
		return false
	}

	// Must not be world-writable.
	if info.Mode()&0002 != 0 {
		// We allow world-writable sticky bit directories like /tmp and /var/tmp
		// themselves to be part of the parent hierarchy, but any files or subdirectories
		// within them must still be secure (not world-writable).
		if path != "/tmp" && path != "/var/tmp" {
			return false
		}
	}

	// Check parent directory
	parent := filepath.Dir(path)
	if parent == path {
		return true // reached root
	}

	return IsSecure(parent)
}
