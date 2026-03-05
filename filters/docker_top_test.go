package filters

import (
	"strings"
	"testing"
)

func TestFilterDockerTop_Short(t *testing.T) {
	raw := "UID        PID  PPID  C STIME TTY      TIME     CMD\n" +
		"root         1     0  0 10:30 ?    00:00:05 node server.js\n" +
		"root        15     1  0 10:30 ?    00:00:01 npm\n"

	got, err := filterDockerTop(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != strings.TrimSpace(raw) {
		t.Error("short output should pass through")
	}
}

func TestFilterDockerTop_Empty(t *testing.T) {
	got, err := filterDockerTop("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
