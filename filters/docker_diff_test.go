package filters

import (
	"fmt"
	"strings"
	"testing"
)

func TestFilterDockerDiff_Few(t *testing.T) {
	raw := "C /app\nA /app/newfile.txt\nA /app/other.txt\nD /app/old.txt\n"

	got, err := filterDockerDiff(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "A /app/newfile.txt") {
		t.Errorf("expected paths shown for few changes, got: %s", got)
	}
}

func TestFilterDockerDiff_Many(t *testing.T) {
	var lines []string
	for i := 0; i < 20; i++ {
		lines = append(lines, fmt.Sprintf("A /app/file%d.txt", i))
	}
	lines = append(lines, "C /app")
	lines = append(lines, "D /app/old.txt")
	raw := strings.Join(lines, "\n") + "\n"

	got, err := filterDockerDiff(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "added(20)") {
		t.Errorf("expected summary format, got: %s", got)
	}
}

func TestFilterDockerDiff_Empty(t *testing.T) {
	got, err := filterDockerDiff("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
