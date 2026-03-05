package filters

import (
	"strings"
	"testing"
)

func TestFilterDockerNetworkLs(t *testing.T) {
	raw := "NETWORK ID     NAME      DRIVER    SCOPE\n" +
		"abc123def456   bridge    bridge    local\n" +
		"def456abc789   host      host      local\n" +
		"ghi789jkl012   none      null      local\n"

	got, err := filterDockerNetworkLs(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "bridge (bridge) local") {
		t.Errorf("expected compact format, got: %s", got)
	}
	if !strings.Contains(got, "3 networks") {
		t.Error("expected network count")
	}
}

func TestFilterDockerNetworkLs_Empty(t *testing.T) {
	got, err := filterDockerNetworkLs("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
