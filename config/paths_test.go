package config

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestConfigDir(t *testing.T) {
	ResetCacheForTest()
	t.Cleanup(ResetCacheForTest)

	// We use a temporary directory for HOME/USERPROFILE to ensure isolation
	tmpHome := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", tmpHome)
		t.Setenv("AppData", filepath.Join(tmpHome, "AppData", "Roaming"))
	} else {
		t.Setenv("HOME", tmpHome)
		// Clear XDG_CONFIG_HOME to test the fallback
		t.Setenv("XDG_CONFIG_HOME", "")
	}

	dir := ConfigDir()
	if !strings.Contains(dir, "chop") {
		t.Errorf("expected path to contain 'chop', got %q", dir)
	}

	if runtime.GOOS == "windows" {
		expectedSuffix := filepath.Join("AppData", "Roaming", "chop")
		if !strings.HasSuffix(dir, expectedSuffix) {
			t.Errorf("expected path to end with %q, got %q", expectedSuffix, dir)
		}
	} else if runtime.GOOS == "darwin" {
		expectedSuffix := filepath.Join("Library", "Application Support", "chop")
		if !strings.HasSuffix(dir, expectedSuffix) {
			t.Errorf("expected path to end with %q, got %q", expectedSuffix, dir)
		}
	} else {
		// On Unix (Linux, etc.), if XDG_CONFIG_HOME is empty, os.UserConfigDir() returns $HOME/.config
		expectedSuffix := filepath.Join(".config", "chop")
		if !strings.HasSuffix(dir, expectedSuffix) {
			t.Errorf("expected path to end with %q, got %q", expectedSuffix, dir)
		}

		// Test with XDG_CONFIG_HOME set (Go ignores this on Darwin/Windows in os.UserConfigDir)
		xdgConfig := filepath.Join(tmpHome, "custom_xdg")
		t.Setenv("XDG_CONFIG_HOME", xdgConfig)
		ResetCacheForTest()
		dir = ConfigDir()
		expected := filepath.Join(xdgConfig, "chop")
		if dir != expected {
			t.Errorf("expected %q, got %q", expected, dir)
		}
	}
}

func TestDataDir(t *testing.T) {
	ResetCacheForTest()
	t.Cleanup(ResetCacheForTest)

	tmpHome := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", tmpHome)
		localAppData := filepath.Join(tmpHome, "AppData", "Local")
		t.Setenv("LocalAppData", localAppData)
	} else {
		t.Setenv("HOME", tmpHome)
		t.Setenv("XDG_DATA_HOME", "")
	}

	dir := DataDir()
	if !strings.Contains(dir, "chop") {
		t.Errorf("expected path to contain 'chop', got %q", dir)
	}

	if runtime.GOOS == "windows" {
		// On Windows, DataDir uses os.UserCacheDir() which usually points to %LocalAppData%
		expectedSuffix := filepath.Join("AppData", "Local", "chop")
		if !strings.HasSuffix(dir, expectedSuffix) {
			t.Errorf("expected path to end with %q, got %q", expectedSuffix, dir)
		}
	} else {
		// Default Unix path: ~/.local/share/chop
		expectedSuffix := filepath.Join(".local", "share", "chop")
		if !strings.HasSuffix(dir, expectedSuffix) {
			t.Errorf("expected path to end with %q, got %q", expectedSuffix, dir)
		}

		// Test with XDG_DATA_HOME set (DataDir explicitly checks this env var on non-Windows)
		xdgData := filepath.Join(tmpHome, "custom_xdg_data")
		t.Setenv("XDG_DATA_HOME", xdgData)
		ResetCacheForTest()
		dir = DataDir()
		expected := filepath.Join(xdgData, "chop")
		if dir != expected {
			t.Errorf("expected %q, got %q", expected, dir)
		}
	}
}
