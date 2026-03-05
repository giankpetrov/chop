package filters

import (
	"strings"
	"testing"
)

func TestFilterKubectlTop(t *testing.T) {
	raw := "NAME                        CPU(cores)   MEMORY(bytes)\n" +
		"web-abc123-xyz              250m         512Mi\n" +
		"api-def456-uvw              100m         256Mi\n"

	got, err := filterKubectlTop(raw)
	if err != nil {
		t.Fatal(err)
	}
	// Function trims input, so compare trimmed
	if got != strings.TrimSpace(raw) {
		t.Errorf("expected passthrough for short output, got: %s", got)
	}
}

func TestFilterKubectlTop_Empty(t *testing.T) {
	got, err := filterKubectlTop("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
