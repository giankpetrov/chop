package config

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Validate ---

func TestValidate_ValidConfig(t *testing.T) {
	tmp := writeTemp(t, "disabled: [git, docker]\n")
	errs := Validate(tmp)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidate_EmptyFile(t *testing.T) {
	tmp := writeTemp(t, "")
	errs := Validate(tmp)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty file, got: %v", errs)
	}
}

func TestValidate_CommentsOnly(t *testing.T) {
	tmp := writeTemp(t, "# just a comment\n")
	errs := Validate(tmp)
	if len(errs) != 0 {
		t.Errorf("expected no errors for comments-only file, got: %v", errs)
	}
}

func TestValidate_UnknownKey(t *testing.T) {
	tmp := writeTemp(t, "unknown_key: value\n")
	errs := Validate(tmp)
	if len(errs) == 0 {
		t.Error("expected error for unknown key")
	}
}

func TestValidate_MissingFile(t *testing.T) {
	errs := Validate("/nonexistent/path/config.yml")
	if len(errs) == 0 {
		t.Error("expected error for missing file")
	}
}

func TestValidate_EmptyDisabledEntry(t *testing.T) {
	// parseList strips empty items, so "[git, , docker]" should not produce empties
	// But "[,]" would. Let's test what actually triggers the empty-entry path.
	// Looking at parseList: items with trimmed value == "" are added if item=="",
	// but the loop skips empty items. So empty entries can't come from bracket lists.
	// The path is only reachable if parseList returns an empty string item -
	// which happens when the list contains only commas/spaces inside brackets.
	// Actually reading parseList again: it only appends if item != "". So empty
	// entries in disabled list are unreachable via the bracket syntax.
	// Verify that a valid disabled list passes cleanly.
	tmp := writeTemp(t, "disabled: [git, docker, kubectl]\n")
	errs := Validate(tmp)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidate_DisabledEmptyList(t *testing.T) {
	tmp := writeTemp(t, "disabled: []\n")
	errs := Validate(tmp)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty disabled list, got: %v", errs)
	}
}

func TestValidate_MultipleUnknownKeys(t *testing.T) {
	content := "disabled: [git]\nunknown1: foo\nunknown2: bar\n"
	tmp := writeTemp(t, content)
	errs := Validate(tmp)
	if len(errs) != 2 {
		t.Errorf("expected 2 errors (one per unknown key), got %d: %v", len(errs), errs)
	}
}

func TestValidate_InlineComment(t *testing.T) {
	tmp := writeTemp(t, "disabled: [git] # this is fine\n")
	errs := Validate(tmp)
	if len(errs) != 0 {
		t.Errorf("expected no errors with inline comment, got: %v", errs)
	}
}

// --- parseKV (same package, directly accessible) ---

func TestParseKV_ValidLine(t *testing.T) {
	key, value, ok := parseKV("disabled: [git, docker]")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if key != "disabled" {
		t.Errorf("expected key 'disabled', got %q", key)
	}
	if value != "[git, docker]" {
		t.Errorf("expected value '[git, docker]', got %q", value)
	}
}

func TestParseKV_NoColon(t *testing.T) {
	_, _, ok := parseKV("no colon here")
	if ok {
		t.Error("expected ok=false for line without colon")
	}
}

func TestParseKV_EmptyValue(t *testing.T) {
	key, value, ok := parseKV("disabled:")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if key != "disabled" {
		t.Errorf("expected key 'disabled', got %q", key)
	}
	if value != "" {
		t.Errorf("expected empty value, got %q", value)
	}
}

func TestParseKV_WhitespaceAround(t *testing.T) {
	key, value, ok := parseKV("  key  :  value  ")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if key != "key" {
		t.Errorf("expected 'key', got %q", key)
	}
	if value != "value" {
		t.Errorf("expected 'value', got %q", value)
	}
}

// --- parseList (same package, directly accessible) ---

func TestParseList_EmptyString(t *testing.T) {
	result := parseList("")
	if result != nil {
		t.Errorf("expected nil for empty string, got %v", result)
	}
}

func TestParseList_EmptyBrackets(t *testing.T) {
	result := parseList("[]")
	if result != nil {
		t.Errorf("expected nil for '[]', got %v", result)
	}
}

func TestParseList_SingleItem(t *testing.T) {
	result := parseList("[git]")
	if len(result) != 1 || result[0] != "git" {
		t.Errorf("expected [git], got %v", result)
	}
}

func TestParseList_MultipleItems(t *testing.T) {
	result := parseList("[git, docker, kubectl]")
	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d: %v", len(result), result)
	}
	if result[0] != "git" || result[1] != "docker" || result[2] != "kubectl" {
		t.Errorf("unexpected items: %v", result)
	}
}

func TestParseList_QuotedItems(t *testing.T) {
	result := parseList(`["git diff", "docker ps"]`)
	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d: %v", len(result), result)
	}
	if result[0] != "git diff" {
		t.Errorf("expected 'git diff', got %q", result[0])
	}
	if result[1] != "docker ps" {
		t.Errorf("expected 'docker ps', got %q", result[1])
	}
}

func TestParseList_SingleQuotedItems(t *testing.T) {
	result := parseList("['git', 'docker']")
	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d: %v", len(result), result)
	}
	if result[0] != "git" || result[1] != "docker" {
		t.Errorf("unexpected items: %v", result)
	}
}

func TestParseList_ExtraWhitespace(t *testing.T) {
	result := parseList("[  git  ,  docker  ]")
	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d: %v", len(result), result)
	}
	if result[0] != "git" || result[1] != "docker" {
		t.Errorf("unexpected items: %v", result)
	}
}

// --- ConfigDir / DataDir ---

func TestConfigDir_WithXDG(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir := ConfigDir()
	expected := filepath.Join(tmp, "chop")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestDataDir_WithXDG(t *testing.T) {
	// Only meaningful on non-Windows (XDG_DATA_HOME is a Unix concept).
	// But DataDir() reads XDG_DATA_HOME on Linux — on Windows it ignores it.
	// We test what we can: the function returns a non-empty path.
	dir := DataDir()
	if dir == "" {
		t.Error("expected non-empty DataDir")
	}
	if filepath.Base(dir) != "chop" {
		t.Errorf("expected path to end in 'chop', got %s", dir)
	}
}

func TestConfigDir_ReturnsChopSubdir(t *testing.T) {
	dir := ConfigDir()
	if filepath.Base(dir) != "chop" {
		t.Errorf("expected ConfigDir to end in 'chop', got %s", dir)
	}
}

// --- LoadWithLocal ---

func TestLoadWithLocal_EmptyCwd(t *testing.T) {
	// Should not panic; just returns the global config.
	cfg := LoadWithLocal("")
	_ = cfg
}

func TestLoadWithLocal_CwdWithLocalFile(t *testing.T) {
	dir := t.TempDir()
	localPath := filepath.Join(dir, ".chop.yml")
	if err := os.WriteFile(localPath, []byte("disabled: [npm]\n"), 0600); err != nil {
		t.Fatal(err)
	}

	// Point global config to a temp dir with no config to ensure predictable state.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := LoadWithLocal(dir)
	if len(cfg.Disabled) != 1 || cfg.Disabled[0] != "npm" {
		t.Errorf("expected [npm] from local file, got %v", cfg.Disabled)
	}
}
