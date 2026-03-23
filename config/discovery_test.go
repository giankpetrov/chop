package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- DiscoveryPath ---

func TestDiscoveryPath_ReturnsPathJSONUnderDotChop(t *testing.T) {
	path, err := DiscoveryPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(path) != "path.json" {
		t.Errorf("expected file name 'path.json', got %q", filepath.Base(path))
	}
	if !strings.Contains(filepath.ToSlash(path), ".chop/") {
		t.Errorf("expected path to contain '.chop/', got %s", path)
	}
}

func TestDiscoveryPath_UnderHomeDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	path, err := DiscoveryPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(path, home) {
		t.Errorf("expected path to start with home dir %s, got %s", home, path)
	}
}

// --- WriteDiscoveryInfo ---

func TestWriteDiscoveryInfo_CreatesFileWithCorrectContent(t *testing.T) {
	// We cannot override os.UserHomeDir() or os.Executable() without changing the
	// source, so we call WriteDiscoveryInfo with a known version string and verify
	// the resulting file is valid JSON with the expected version field.
	const testVersion = "v1.2.3-test"

	err := WriteDiscoveryInfo(testVersion)
	if err != nil {
		t.Fatalf("WriteDiscoveryInfo returned error: %v", err)
	}

	// Locate the file using DiscoveryPath so the path logic is consistent.
	path, err := DiscoveryPath()
	if err != nil {
		t.Fatalf("DiscoveryPath error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read discovery file %s: %v", path, err)
	}

	var info DiscoveryInfo
	if err := json.Unmarshal(data, &info); err != nil {
		t.Fatalf("discovery file is not valid JSON: %v", err)
	}

	if info.Version != testVersion {
		t.Errorf("expected version %q, got %q", testVersion, info.Version)
	}
	if info.Path == "" {
		t.Error("expected non-empty Path in discovery info")
	}
}

func TestWriteDiscoveryInfo_FileIsValidJSON(t *testing.T) {
	err := WriteDiscoveryInfo("vtest")
	if err != nil {
		t.Fatalf("WriteDiscoveryInfo returned error: %v", err)
	}

	path, err := DiscoveryPath()
	if err != nil {
		t.Fatalf("DiscoveryPath error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Errorf("file is not valid JSON: %v\ncontent: %s", err, string(data))
	}

	if _, ok := raw["version"]; !ok {
		t.Error("expected 'version' key in JSON output")
	}
	if _, ok := raw["path"]; !ok {
		t.Error("expected 'path' key in JSON output")
	}
}

// --- ValidateFilters ---

func writeFilterTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "filters.yml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestValidateFilters_MissingFile(t *testing.T) {
	errs := ValidateFilters("/nonexistent/path/filters.yml")
	if len(errs) == 0 {
		t.Error("expected error for missing file")
	}
}

func TestValidateFilters_InvalidYAML(t *testing.T) {
	path := writeFilterTemp(t, "{{invalid yaml")
	errs := ValidateFilters(path)
	if len(errs) == 0 {
		t.Error("expected error for invalid YAML")
	}
}

func TestValidateFilters_ValidFilters(t *testing.T) {
	content := `
filters:
  "mycli build":
    keep: ["ERROR", "^BUILD"]
    drop: ["DEBUG"]
  terraform:
    head: 10
    tail: 5
`
	path := writeFilterTemp(t, content)
	errs := ValidateFilters(path)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid filters, got: %v", errs)
	}
}

func TestValidateFilters_InvalidKeepRegex(t *testing.T) {
	content := `
filters:
  "mycli build":
    keep: ["[invalid regex"]
`
	path := writeFilterTemp(t, content)
	errs := ValidateFilters(path)
	if len(errs) == 0 {
		t.Error("expected error for invalid keep regex")
	}
}

func TestValidateFilters_InvalidDropRegex(t *testing.T) {
	content := `
filters:
  "mycli build":
    drop: ["(unclosed"]
`
	path := writeFilterTemp(t, content)
	errs := ValidateFilters(path)
	if len(errs) == 0 {
		t.Error("expected error for invalid drop regex")
	}
}

func TestValidateFilters_MissingExecScript(t *testing.T) {
	content := `
filters:
  "mycli build":
    exec: "/nonexistent/path/to/script.sh"
`
	path := writeFilterTemp(t, content)
	errs := ValidateFilters(path)
	if len(errs) == 0 {
		t.Error("expected error for missing exec script")
	}
}

func TestValidateFilters_ValidExecScript(t *testing.T) {
	// Create an actual script file so validation passes.
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "filter.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\ncat\n"), 0700); err != nil {
		t.Fatal(err)
	}

	content := "filters:\n  \"mycli\":\n    exec: \"" + filepath.ToSlash(scriptPath) + "\"\n"
	filterPath := filepath.Join(dir, "filters.yml")
	if err := os.WriteFile(filterPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	errs := ValidateFilters(filterPath)
	if len(errs) != 0 {
		t.Errorf("expected no errors for existing exec script, got: %v", errs)
	}
}

func TestValidateFilters_EmptyFile(t *testing.T) {
	path := writeFilterTemp(t, "")
	errs := ValidateFilters(path)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty file, got: %v", errs)
	}
}

func TestValidateFilters_MultipleErrors(t *testing.T) {
	content := `
filters:
  "tool1":
    keep: ["[bad regex"]
  "tool2":
    drop: ["(unclosed"]
`
	path := writeFilterTemp(t, content)
	errs := ValidateFilters(path)
	if len(errs) < 2 {
		t.Errorf("expected at least 2 errors, got %d: %v", len(errs), errs)
	}
}
