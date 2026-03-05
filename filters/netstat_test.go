package filters

import (
	"fmt"
	"strings"
	"testing"
)

func TestFilterNetstat(t *testing.T) {
	lines := []string{
		"State    Recv-Q   Send-Q     Local Address:Port      Peer Address:Port   Process",
		"LISTEN   0        128        0.0.0.0:22              0.0.0.0:*           sshd",
		"LISTEN   0        128        0.0.0.0:80              0.0.0.0:*           nginx",
	}
	for i := 0; i < 25; i++ {
		lines = append(lines, fmt.Sprintf("ESTAB    0        0          10.0.0.5:80             192.168.1.%d:50000   nginx", i))
	}
	raw := strings.Join(lines, "\n") + "\n"

	got, err := filterNetstat(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "LISTEN") {
		t.Error("expected LISTEN entries")
	}
	if !strings.Contains(got, "total connections") {
		t.Error("expected total count")
	}
}

func TestFilterNetstat_Empty(t *testing.T) {
	got, err := filterNetstat("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
