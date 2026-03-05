package filters

import (
	"strings"
	"testing"
)

func TestFilterDockerVolumeLs(t *testing.T) {
	raw := "DRIVER    VOLUME NAME\n" +
		"local     my-data\n" +
		"local     db-data\n" +
		"local     cache-vol\n"

	got, err := filterDockerVolumeLs(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "my-data (local)") {
		t.Errorf("expected compact format, got: %s", got)
	}
	if !strings.Contains(got, "3 volumes") {
		t.Error("expected volume count")
	}
}

func TestFilterDockerVolumeLs_Empty(t *testing.T) {
	got, err := filterDockerVolumeLs("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
