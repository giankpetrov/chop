package filters

import (
	"strings"
	"testing"
)

func TestFilterPipInstall(t *testing.T) {
	raw := `Collecting flask
  Downloading flask-3.0.0-py3-none-any.whl (101 kB)
Collecting werkzeug>=3.0.0
  Downloading werkzeug-3.0.1-py3-none-any.whl (226 kB)
Collecting jinja2>=3.1.2
  Using cached jinja2-3.1.2-py3-none-any.whl (133 kB)
Installing collected packages: werkzeug, jinja2, flask
Successfully installed flask-3.0.0 jinja2-3.1.2 werkzeug-3.0.1`

	got, err := filterPipInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "installed 3 packages") {
		t.Errorf("expected package count, got: %s", got)
	}
}

func TestFilterPipList(t *testing.T) {
	raw := "Package    Version\n---------- -------\nflask      3.0.0\njinja2     3.1.2\nwerkzeug   3.0.1\n"

	got, err := filterPipList(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "3 packages installed") {
		t.Errorf("expected count, got: %s", got)
	}
}

func TestFilterPipInstall_Empty(t *testing.T) {
	got, err := filterPipInstall("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
