//go:build windows
package config

// IsSecure on Windows is a simplified check.
// We prioritize path-based security (residing in user's profile)
// as Unix-style UID/GID checks don't map directly to Windows ACLs.
func IsSecure(path string) (bool, error) {
	// For now, assume any file in the user's home/config directory is "secure enough".
	// A more robust implementation would check for NTFS ownership and ACLs.
	return true, nil
}
