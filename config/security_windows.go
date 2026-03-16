//go:build windows

package config

import (
	"os"
)

// IsSecure checks if the given path is secure. On Windows, we perform a
// simplified check as UID/GID and fine-grained permission bits are different.
func IsSecure(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
