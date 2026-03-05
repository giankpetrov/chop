package filters

import (
	"strings"
	"testing"
)

func TestFilterComposerInstall(t *testing.T) {
	raw := "Installing dependencies from lock file (including require-dev)\n" +
		"Verifying lock file contents can be installed on current platform.\n" +
		"Package operations: 25 installs, 3 updates, 0 removals\n" +
		"  - Installing symfony/console (v6.4.0): Extracting archive\n" +
		"  - Installing symfony/http-kernel (v6.4.0): Extracting archive\n" +
		"  - Updating laravel/framework (v10.0.0 => v10.1.0): Extracting archive\n" +
		"Generating optimized autoload files\n"

	got, err := filterComposerInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "25 installs") {
		t.Errorf("expected ops summary, got: %s", got)
	}
}

func TestFilterComposerInstall_Empty(t *testing.T) {
	got, err := filterComposerInstall("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
