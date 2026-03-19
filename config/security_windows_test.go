//go:build windows

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsSecure_Windows(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get user home dir: %v", err)
	}

	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")

	// Helper: create a real file under a given directory.
	makeFile := func(t *testing.T, dir string) string {
		t.Helper()
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("MkdirAll %s: %v", dir, err)
		}
		f, err := os.CreateTemp(dir, "chop-sec-test-*")
		if err != nil {
			t.Fatalf("CreateTemp under %s: %v", dir, err)
		}
		f.Close()
		t.Cleanup(func() { os.Remove(f.Name()) })
		return f.Name()
	}

	t.Run("nonexistent path returns false", func(t *testing.T) {
		path := filepath.Join(homeDir, "chop-nonexistent-file-that-does-not-exist-xyz.tmp")
		if IsSecure(path) {
			t.Error("expected false for non-existent path")
		}
	})

	t.Run("file under UserHomeDir returns true", func(t *testing.T) {
		subDir := filepath.Join(homeDir, "chop-sec-test-home")
		path := makeFile(t, subDir)
		t.Cleanup(func() { os.RemoveAll(subDir) })
		if !IsSecure(path) {
			t.Errorf("expected true for file under home dir %s, got false for %s", homeDir, path)
		}
	})

	t.Run("file under LOCALAPPDATA returns true", func(t *testing.T) {
		if localAppData == "" {
			t.Skip("LOCALAPPDATA not set")
		}
		subDir := filepath.Join(localAppData, "chop-sec-test-local")
		path := makeFile(t, subDir)
		t.Cleanup(func() { os.RemoveAll(subDir) })
		if !IsSecure(path) {
			t.Errorf("expected true for file under LOCALAPPDATA %s, got false for %s", localAppData, path)
		}
	})

	t.Run("file under APPDATA returns true", func(t *testing.T) {
		if appData == "" {
			t.Skip("APPDATA not set")
		}
		subDir := filepath.Join(appData, "chop-sec-test-roaming")
		path := makeFile(t, subDir)
		t.Cleanup(func() { os.RemoveAll(subDir) })
		if !IsSecure(path) {
			t.Errorf("expected true for file under APPDATA %s, got false for %s", appData, path)
		}
	})

	t.Run("nonexistent path outside user dirs returns false", func(t *testing.T) {
		// os.Stat will fail, so IsSecure returns false before checking the prefix.
		path := `C:\Windows\System32\chop-nonexistent-test-file.tmp`
		if IsSecure(path) {
			t.Errorf("expected false for path outside user dirs: %s", path)
		}
	})

	t.Run("t.TempDir under LOCALAPPDATA Temp returns true", func(t *testing.T) {
		if localAppData == "" {
			t.Skip("LOCALAPPDATA not set")
		}
		// t.TempDir() on Windows lives under %LOCALAPPDATA%\Temp.
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "chop-sec-test.tmp")
		if err := os.WriteFile(tmpFile, []byte("test"), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		if !IsSecure(tmpFile) {
			t.Errorf("expected true for t.TempDir() file %s (should be under LOCALAPPDATA)", tmpFile)
		}
	})
}
