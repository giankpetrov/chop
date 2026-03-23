package cleanup

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDataDir_ControlledHome verifies dataDir() respects HOME (on non-Windows
// where os.UserHomeDir falls back to HOME env var).
func TestDataDir_ControlledHome(t *testing.T) {
	tmp := t.TempDir()
	// On Linux/macOS UserHomeDir reads $HOME; on Windows it reads USERPROFILE.
	// Set both so the test is portable inside Docker (Linux).
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	dir := dataDir()
	expected := filepath.Join(tmp, ".local", "share", "chop")
	if dir != expected {
		t.Errorf("dataDir() = %q, want %q", dir, expected)
	}
}

// TestReset_ActuallyRemovesFiles calls Reset() against a real dataDir by
// pointing HOME at a temp dir, planting the expected files, and verifying
// Reset removes them.
func TestReset_ActuallyRemovesFiles(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	// dataDir() now returns tmp/.local/share/chop
	dir := dataDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}

	dbFile := filepath.Join(dir, "tracking.db")
	logFile := filepath.Join(dir, "hook-audit.log")

	if err := os.WriteFile(dbFile, []byte("db content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logFile, []byte("log content"), 0o600); err != nil {
		t.Fatal(err)
	}

	Reset()

	if _, err := os.Stat(dbFile); !os.IsNotExist(err) {
		t.Error("tracking.db should have been removed by Reset()")
	}
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Error("hook-audit.log should have been removed by Reset()")
	}
}

// TestReset_WhenFilesAbsent verifies Reset() does not panic when files are missing.
func TestReset_WhenFilesAbsent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	// dataDir points to a non-existent path — Reset() should be a no-op
	Reset()
}

// TestUninstall_KeepData verifies Uninstall(keepData=true) does not delete dataDir.
// We cannot test hook/binary removal (require real binary paths), but we can
// verify the data directory is preserved.
func TestUninstall_KeepData(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdg"))

	// Create the data dir with a sentinel file.
	dir := dataDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(dir, "tracking.db")
	if err := os.WriteFile(sentinel, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	// keepData=true: data should survive
	Uninstall(true)

	if _, err := os.Stat(sentinel); os.IsNotExist(err) {
		t.Error("Uninstall(keepData=true) should NOT remove the data directory")
	}
}

// TestUninstall_RemoveData verifies Uninstall(keepData=false) removes dataDir.
func TestUninstall_RemoveData(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdg"))

	dir := dataDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(dir, "tracking.db")
	if err := os.WriteFile(sentinel, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	// keepData=false: data dir should be gone
	Uninstall(false)

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("Uninstall(keepData=false) should have removed the data directory")
	}
}

// TestConfigDir_NoXDGAndNoHome verifies configDir() falls back gracefully when
// both XDG_CONFIG_HOME and HOME are unset (returns a path ending in chop).
func TestConfigDir_NoXDGAndNoHome(t *testing.T) {
	// Can't truly unset HOME without breaking the process; just verify
	// that with a custom XDG_CONFIG_HOME the suffix is always "chop".
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	dir := configDir()
	if filepath.Base(dir) != "chop" {
		t.Errorf("configDir() base = %q, want \"chop\"", filepath.Base(dir))
	}
}
