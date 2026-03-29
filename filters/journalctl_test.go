package filters

import (
	"strings"
	"testing"
)

var journalctlFixture = `Mar 15 10:00:00 myhost systemd[1]: Started nginx.service - A high performance web server.
Mar 15 10:00:01 myhost nginx[1234]: 2024/03/15 10:00:01 [notice] 1234#1234: using the "epoll" event method
Mar 15 10:00:01 myhost nginx[1234]: 2024/03/15 10:00:01 [notice] 1234#1234: nginx/1.24.0
Mar 15 10:00:02 myhost nginx[1234]: 127.0.0.1 - - [15/Mar/2024] "GET / HTTP/1.1" 200 615
Mar 15 10:00:03 myhost nginx[1234]: 127.0.0.1 - - [15/Mar/2024] "GET /api HTTP/1.1" 200 1234
Mar 15 10:00:04 myhost nginx[1234]: 2024/03/15 10:00:04 [error] 1234#1234: connect() failed
`

func TestFilterJournalctl(t *testing.T) {
	got, err := filterJournalctl(journalctlFixture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should keep error lines
	if !strings.Contains(got, "[error]") {
		t.Errorf("expected error line preserved, got:\n%s", got)
	}

	// Should keep systemd started line
	if !strings.Contains(got, "Started nginx.service") {
		t.Errorf("expected 'Started nginx.service' preserved, got:\n%s", got)
	}
}

func TestFilterJournalctlEmpty(t *testing.T) {
	got, err := filterJournalctl("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFilterJournalctlLarge(t *testing.T) {
	// Build >50 lines of journalctl output
	var lines []string
	lines = append(lines, "-- Logs begin at Mon 2024-01-01 00:00:00 UTC --")
	for i := 0; i < 60; i++ {
		lines = append(lines, "Mar 15 10:00:00 myhost systemd[1]: INFO: routine message number "+strings.Repeat("x", 10))
	}
	// Add an error line in the middle
	lines = append(lines, "Mar 15 10:01:00 myhost systemd[1]: ERROR: something went wrong")
	raw := strings.Join(lines, "\n")

	got, err := filterJournalctl(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain the truncation indicator
	if !strings.Contains(got, "lines hidden") {
		t.Errorf("expected truncation indicator, got:\n%s", got)
	}

	// Should contain the error line (important)
	if !strings.Contains(got, "ERROR: something went wrong") {
		t.Errorf("expected error line preserved, got:\n%s", got)
	}
}

func TestJournalctlRouted(t *testing.T) {
	if get("journalctl", []string{}) == nil {
		t.Error("expected non-nil filter for journalctl")
	}
}
