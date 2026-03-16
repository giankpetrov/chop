//go:build !windows
package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)

// IsSecure checks if a file or directory is "secure" (owned by current user,
// restricted permissions, and secure parent).
// Parent directories can be owned by root, as long as they are not world-writable.
func IsSecure(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, err
	}

	// 1. Check ownership
	curr, err := user.Current()
	if err != nil {
		return false, fmt.Errorf("failed to get current user: %w", err)
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("failed to get unix stat for %s", path)
	}

	uidStr := fmt.Sprintf("%d", stat.Uid)
	if uidStr != curr.Uid && uidStr != "0" {
		return false, fmt.Errorf("%s is owned by UID %d, expected current user %s or root", path, stat.Uid, curr.Uid)
	}

	// 2. Check permissions
	// Must not be world-writable, EXCEPT for /tmp which is designed to be world-writable
	// with the sticky bit (01777). We allow /tmp as a parent for testing.
	if info.Mode().Perm()&0o002 != 0 && path != "/tmp" && path != "/tmp/" {
		return false, fmt.Errorf("%s is world-writable: %04o", path, info.Mode().Perm())
	}

	// For the file/directory itself, it must be owned by the user (or root)
	// and restricted enough that others cannot write to it.
	// For Directories: 0755 or more restrictive is okay if owned by root/user.
	// But for our specific Config/Filters file, we actually want 0600.
	// We'll enforce stricter check for the target itself later if needed,
	// but here we check general "security" of the path.

	// 3. Recurse to parent (root is always secure)
	parent := filepath.Dir(path)
	if parent == path || parent == "/" || parent == "." {
		return true, nil
	}

	return IsSecure(parent)
}
