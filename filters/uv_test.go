package filters

import (
	"strings"
	"testing"
)

func TestFilterUvInstall(t *testing.T) {
	raw := "Resolved 25 packages in 200ms\n" +
		"Prepared 25 packages in 500ms\n" +
		"Installed 25 packages in 100ms\n" +
		" + flask==3.0.0\n" +
		" + jinja2==3.1.2\n" +
		" + werkzeug==3.0.1\n"

	got, err := filterUvInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "installed 25 packages") {
		t.Errorf("expected install summary, got: %s", got)
	}
}

func TestFilterUvInstall_Empty(t *testing.T) {
	got, err := filterUvInstall("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
