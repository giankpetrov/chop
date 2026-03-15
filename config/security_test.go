package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsSecure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	dir := t.TempDir()

	t.Run("secure file", func(t *testing.T) {
		path := filepath.Join(dir, "secure.txt")
		if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
			t.Fatal(err)
		}
		if !IsSecure(path) {
			t.Errorf("expected %s to be secure", path)
		}
	})

	t.Run("insecure file (group readable)", func(t *testing.T) {
		path := filepath.Join(dir, "insecure_group.txt")
		if err := os.WriteFile(path, []byte("data"), 0o640); err != nil {
			t.Fatal(err)
		}
		if IsSecure(path) {
			t.Errorf("expected %s to be insecure", path)
		}
	})

	t.Run("insecure file (world readable)", func(t *testing.T) {
		path := filepath.Join(dir, "insecure_world.txt")
		if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
		if IsSecure(path) {
			t.Errorf("expected %s to be insecure", path)
		}
	})

	t.Run("secure directory", func(t *testing.T) {
		path := filepath.Join(dir, "secure_dir")
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
		if !IsSecure(path) {
			t.Errorf("expected %s to be secure", path)
		}
	})

	t.Run("insecure directory", func(t *testing.T) {
		path := filepath.Join(dir, "insecure_dir")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
		if IsSecure(path) {
			t.Errorf("expected %s to be insecure", path)
		}
	})
}

func TestLoadCustomFilters_Insecure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	configPath := filepath.Join(tmpDir, "chop", "filters.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}

	// Write with insecure permissions
	content := `
filters:
  "pwned":
    exec: "touch /tmp/pwned"
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	result := LoadCustomFilters()
	if result == nil {
		t.Fatal("expected filters to be loaded even if insecure")
	}

	f, ok := result["pwned"]
	if !ok {
		t.Fatal("missing 'pwned' filter")
	}

	if f.Trusted {
		t.Error("expected insecure global filter to be untrusted")
	}
}
