package updater

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestAutoUpdateConfiguration(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Default is off
	if IsAutoUpdateEnabled() {
		t.Error("IsAutoUpdateEnabled() = true, want false")
	}

	// Enable
	if err := SetAutoUpdate(true); err != nil {
		t.Fatalf("SetAutoUpdate(true) failed: %v", err)
	}
	if !IsAutoUpdateEnabled() {
		t.Error("IsAutoUpdateEnabled() = false, want true")
	}

	// Disable
	if err := SetAutoUpdate(false); err != nil {
		t.Fatalf("SetAutoUpdate(false) failed: %v", err)
	}
	if IsAutoUpdateEnabled() {
		t.Error("IsAutoUpdateEnabled() = true, want false")
	}

	// Disable again (should be no-op/success)
	if err := SetAutoUpdate(false); err != nil {
		t.Fatalf("SetAutoUpdate(false) second call failed: %v", err)
	}
}

func TestUpdateAvailable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	version := "v1.2.3"
	recordUpdateAvailable(version)

	p, err := updateAvailablePath()
	if err != nil {
		t.Fatalf("updateAvailablePath() failed: %v", err)
	}

	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("os.ReadFile() failed: %v", err)
	}
	if string(data) != version {
		t.Errorf("got %q, want %q", string(data), version)
	}

	clearUpdateAvailable()
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Error("clearUpdateAvailable() did not remove the file")
	}
}

func TestNotifyIfUpdateAvailable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	current := "v1.0.0"
	latest := "v1.1.0"

	// Scenario 1: No update recorded
	if output := captureStderr(func() { NotifyIfUpdateAvailable(current) }); output != "" {
		t.Errorf("expected no output when no update recorded, got %q", output)
	}

	// Scenario 2: Update recorded, auto-update off
	recordUpdateAvailable(latest)
	expected := fmt.Sprintf("openchop: update available %s -> %s (run 'openchop update')\n", current, latest)
	if output := captureStderr(func() { NotifyIfUpdateAvailable(current) }); output != expected {
		t.Errorf("got %q, want %q", output, expected)
	}

	// Scenario 3: Update recorded, auto-update on (should be silent)
	SetAutoUpdate(true)
	if output := captureStderr(func() { NotifyIfUpdateAvailable(current) }); output != "" {
		t.Errorf("expected no output when auto-update is on, got %q", output)
	}
	SetAutoUpdate(false)

	// Scenario 4: Dev version (should be silent)
	if output := captureStderr(func() { NotifyIfUpdateAvailable("dev") }); output != "" {
		t.Errorf("expected no output for dev version, got %q", output)
	}

	// Scenario 5: Already on latest version (should be silent)
	if output := captureStderr(func() { NotifyIfUpdateAvailable(latest) }); output != "" {
		t.Errorf("expected no output when already on latest, got %q", output)
	}
}

func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = old
	}()

	f()

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()
	return buf.String()
}
