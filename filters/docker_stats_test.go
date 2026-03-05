package filters

import (
	"strings"
	"testing"
)

func TestFilterDockerStats(t *testing.T) {
	raw := "CONTAINER ID   NAME      CPU %     MEM USAGE / LIMIT     MEM %     NET I/O           BLOCK I/O         PIDS\n" +
		"abc123def456   web       0.50%     150MiB / 1GiB         15.00%    1.2kB / 500B      8.19kB / 0B       10\n" +
		"def456abc789   db        2.30%     500MiB / 2GiB         25.00%    2.5kB / 1.1kB     16.4kB / 12.3kB   25\n"

	got, err := filterDockerStats(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "web") || !strings.Contains(got, "db") {
		t.Error("expected container names")
	}
	if !strings.Contains(got, "2 containers") {
		t.Error("expected container count")
	}
}

func TestFilterDockerStats_Empty(t *testing.T) {
	got, err := filterDockerStats("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
