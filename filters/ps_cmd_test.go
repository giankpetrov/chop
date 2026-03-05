package filters

import (
	"fmt"
	"strings"
	"testing"
)

func TestFilterPsCmd(t *testing.T) {
	header := "USER         PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND\n"
	var lines []string
	lines = append(lines, header)
	lines = append(lines, "user        1234  2.5  1.2 456789 98765 pts/0    Sl+  10:30   1:23 node server.js\n")
	for i := 0; i < 30; i++ {
		lines = append(lines, fmt.Sprintf("root        %d  0.0  0.0   1000   500 ?        S    Jan10   0:00 [kthread%d]\n", 100+i, i))
	}

	raw := strings.Join(lines, "")
	got, err := filterPsCmd(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "active of") {
		t.Error("expected active count")
	}
}

func TestFilterPsCmd_Empty(t *testing.T) {
	got, err := filterPsCmd("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
