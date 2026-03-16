//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// IsSecure checks if the given path is secure (owned by current user or root,
// and not world-writable). Parent directories are also checked.
func IsSecure(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}

	uid := os.Getuid()
	// Must be owned by current user or root
	if stat.Uid != uint32(uid) && stat.Uid != 0 {
		return false
	}

	// Must not be world-writable
	if info.Mode()&0002 != 0 {
		// Exception for /tmp which is standard in some environments,
		// but we still want to be careful.
		if !strings.HasPrefix(path, "/tmp") {
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
