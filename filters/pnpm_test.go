package filters

import (
	"strings"
	"testing"
)

func TestFilterPnpmInstall(t *testing.T) {
	raw := "Packages: +150\n" +
		"++++++++++++++++++++++++++++++\n" +
		"Progress: resolved 200, reused 150, downloaded 50, added 150, done\n" +
		"\ndevDependencies:\n+ @types/node 20.0.0\n\ndependencies:\n+ express 4.18.2\n"

	got, err := filterPnpmInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "added 150 packages") {
		t.Errorf("expected package count, got: %s", got)
	}
}

func TestFilterPnpmInstall_Empty(t *testing.T) {
	got, err := filterPnpmInstall("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
