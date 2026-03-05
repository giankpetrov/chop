package filters

import (
	"strings"
	"testing"
)

func TestFilterRuff(t *testing.T) {
	raw := "src/app.py:1:1: F401 [*] `os` imported but unused\n" +
		"src/app.py:5:1: E302 [*] Expected 2 blank lines, found 1\n" +
		"src/utils.py:10:5: F841 Local variable `x` is assigned to but never used\n" +
		"Found 3 errors.\n" +
		"[*] 2 fixable with the `--fix` option.\n"

	got, err := filterRuff(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "F401") {
		t.Error("expected error code")
	}
	if !strings.Contains(got, "3 problems") {
		t.Error("expected problem count")
	}
}

func TestFilterPylint(t *testing.T) {
	raw := "************* Module app\n" +
		"src/app.py:1:0: C0114: Missing module docstring (missing-module-docstring)\n" +
		"src/app.py:5:0: C0116: Missing function or method docstring (missing-function-docstring)\n"

	got, err := filterPylint(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "C0114") {
		t.Error("expected pylint code")
	}
	if !strings.Contains(got, "2 problems") {
		t.Error("expected problem count")
	}
}

func TestFilterRuff_Empty(t *testing.T) {
	got, err := filterRuff("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "no problems" {
		t.Errorf("expected 'no problems', got %q", got)
	}
}
