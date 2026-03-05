package filters

import (
	"strings"
	"testing"
)

func TestFilterRubocop(t *testing.T) {
	raw := "Inspecting 25 files\n" +
		".....C.W..C.....C........\n\n" +
		"Offenses:\n\n" +
		"app/controllers/users_controller.rb:10:5: C: Style/StringLiterals: Prefer single-quoted strings.\n" +
		"    \"hello\"\n" +
		"    ^^^^^^^\n" +
		"app/controllers/users_controller.rb:15:3: W: Lint/UselessAssignment: Useless assignment.\n" +
		"    temp = 1\n" +
		"    ^^^^\n" +
		"app/models/user.rb:5:1: C: Layout/TrailingEmptyLines: Final newline missing.\n\n" +
		"25 files inspected, 3 offenses detected, 2 offenses autocorrectable\n"

	got, err := filterRubocop(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "Style/StringLiterals") {
		t.Error("expected cop name")
	}
	if !strings.Contains(got, "3 offenses") {
		t.Error("expected offense count")
	}
}

func TestFilterRubocop_Empty(t *testing.T) {
	got, err := filterRubocop("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "no offenses" {
		t.Errorf("expected 'no offenses', got %q", got)
	}
}
