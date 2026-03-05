package filters

import (
	"strings"
	"testing"
)

func TestFilterCompiler(t *testing.T) {
	raw := "main.c: In function 'main':\n" +
		"main.c:10:5: warning: implicit declaration of function 'printf' [-Wimplicit-function-declaration]\n" +
		"   10 |     printf(\"hello\\n\");\n" +
		"      |     ^~~~~~\n" +
		"main.c:15:12: error: expected ';' before '}' token\n" +
		"   15 |     return 0\n" +
		"      |            ^\n"

	got, err := filterCompiler(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "errors(1)") {
		t.Error("expected error count")
	}
	if !strings.Contains(got, "warnings(1)") {
		t.Error("expected warning count")
	}
}

func TestFilterCompiler_Empty(t *testing.T) {
	got, err := filterCompiler("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
