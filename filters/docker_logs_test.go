package filters

import (
	"strings"
	"testing"
)

func TestFilterDockerLogs_PlainText(t *testing.T) {
	raw := strings.Repeat("2024-01-15T10:30:00Z INFO Starting server on port 8080\n", 100) +
		"2024-01-15T10:30:01Z ERROR Connection refused to database\n"

	got, err := filterDockerLogs(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression, got %d >= %d", len(got), len(raw))
	}
	if !strings.Contains(got, "ERROR") {
		t.Error("expected error line preserved")
	}
}

func TestFilterDockerLogs_JSON(t *testing.T) {
	raw := `{"timestamp":"2024-01-15T10:30:00Z","level":"INFO","message":"request handled"}
{"timestamp":"2024-01-15T10:30:00Z","level":"INFO","message":"request handled"}
{"timestamp":"2024-01-15T10:30:00Z","level":"INFO","message":"request handled"}
{"timestamp":"2024-01-15T10:30:01Z","level":"ERROR","message":"database timeout"}
`
	got, err := filterDockerLogs(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "ERROR") {
		t.Error("expected error line preserved")
	}
}

func TestFilterDockerLogs_Empty(t *testing.T) {
	got, err := filterDockerLogs("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
