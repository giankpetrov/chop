package filters

import (
	"strings"
	"testing"
)

func TestFilterYarnInstall(t *testing.T) {
	raw := "yarn install v1.22.19\n" +
		"[1/4] Resolving packages...\n" +
		"[2/4] Fetching packages...\n" +
		"[3/4] Linking dependencies...\n" +
		"[4/4] Building fresh packages...\n" +
		"Done in 10.5s.\n"

	got, err := filterYarnInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "Done in") {
		t.Errorf("expected done line, got: %s", got)
	}
}

func TestFilterYarnInstall_Empty(t *testing.T) {
	got, err := filterYarnInstall("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
