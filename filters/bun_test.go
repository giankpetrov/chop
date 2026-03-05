package filters

import (
	"strings"
	"testing"
)

func TestFilterBunInstall(t *testing.T) {
	raw := "bun install v1.0.0\n" +
		"Resolving dependencies...\n" +
		"Resolved 150 packages\n" +
		"Installed 150 packages in 2.5s\n"

	got, err := filterBunInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "Installed 150 packages") {
		t.Errorf("expected install summary, got: %s", got)
	}
}

func TestFilterBunInstall_Empty(t *testing.T) {
	got, err := filterBunInstall("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
