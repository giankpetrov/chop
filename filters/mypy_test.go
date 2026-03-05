package filters

import (
	"strings"
	"testing"
)

func TestFilterMypy(t *testing.T) {
	raw := `src/app.py:12: error: Incompatible types in assignment (expression has type "str", variable has type "int")  [assignment]
src/app.py:25: error: "Dict[str, Any]" has no attribute "missing"  [attr-defined]
src/utils.py:8: note: Revealed type is "builtins.str"
Found 2 errors in 2 files (checked 10 source files)`

	got, err := filterMypy(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "[assignment]") {
		t.Error("expected error code group")
	}
	if !strings.Contains(got, "Found 2 errors") {
		t.Error("expected summary")
	}
	if strings.Contains(got, "note:") {
		t.Error("notes should be skipped")
	}
}

func TestFilterMypy_Success(t *testing.T) {
	raw := "Success: no issues found in 10 source files"
	got, err := filterMypy(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != raw {
		t.Errorf("expected passthrough for success, got: %s", got)
	}
}

func TestFilterMypy_Empty(t *testing.T) {
	got, err := filterMypy("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
