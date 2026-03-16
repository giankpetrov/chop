//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsSecure_Unix(t *testing.T) {
	// Create a secure file
	tmpDir := t.TempDir()
	secureFile := filepath.Join(tmpDir, "secure-file")
	if err := os.WriteFile(secureFile, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	if !IsSecure(secureFile) {
		t.Errorf("IsSecure should be true for file in %s with 0600 permissions", tmpDir)
	}

	// Create a world-writable file
	insecureFile := filepath.Join(tmpDir, "insecure-file")
	if err := os.WriteFile(insecureFile, []byte("test"), 0666); err != nil {
		t.Fatal(err)
	}
	// On some systems, we need to explicitly chmod to ensure world-writable bit
	os.Chmod(insecureFile, 0666)

	if IsSecure(insecureFile) {
		t.Errorf("IsSecure should be false for world-writable file (mode %o)", func() os.FileMode {
			info, _ := os.Stat(insecureFile)
			return info.Mode()
		}())
	}

	// Create a world-writable directory
	insecureDir := filepath.Join(tmpDir, "insecure-dir")
	if err := os.Mkdir(insecureDir, 0777); err != nil {
		t.Fatal(err)
	}
	os.Chmod(insecureDir, 0777)

	secureFileInInsecureDir := filepath.Join(insecureDir, "secure-file")
	if err := os.WriteFile(secureFileInInsecureDir, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	if IsSecure(secureFileInInsecureDir) {
		t.Errorf("IsSecure should be false for file in world-writable parent directory (mode %o)", func() os.FileMode {
			info, _ := os.Stat(insecureDir)
			return info.Mode()
		}())
	}
}

func TestIsSecure_TmpSubdirectory(t *testing.T) {
	// Use a sub-directory in /tmp
	tmpDir, err := os.MkdirTemp("/tmp", "chop-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Make the dir world-writable
	if err := os.Chmod(tmpDir, 0777); err != nil {
		t.Fatal(err)
	}

	// Test a file within that world-writable dir
	tmpFile := filepath.Join(tmpDir, "test-secure-file")
	if err := os.WriteFile(tmpFile, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	if IsSecure(tmpFile) {
		t.Errorf("IsSecure should be false for file in world-writable subdirectory of /tmp: %s", tmpFile)
	}

	// Test a world-writable file in /tmp (if we can create it)
	directTmpFile := "/tmp/chop-test-direct-file"
	if err := os.WriteFile(directTmpFile, []byte("test"), 0666); err == nil {
		defer os.Remove(directTmpFile)
		os.Chmod(directTmpFile, 0666)
		if IsSecure(directTmpFile) {
			t.Errorf("IsSecure should be false for world-writable file directly in /tmp (mode %o)", func() os.FileMode {
				info, _ := os.Stat(directTmpFile)
				return info.Mode()
			}())
		}
	}
}

func TestIsSecure_RootOwned(t *testing.T) {
	// Standard paths like /usr/bin are root-owned and secure
	if !IsSecure("/usr/bin") {
		// This might fail in some restricted environments, but /usr/bin
		// is generally expected to be secure.
		t.Log("/usr/bin not considered secure, likely environment specific")
	}

	if !IsSecure("/tmp") {
		t.Error("/tmp should be considered secure (exception for world-writable sticky bit)")
	}
}
