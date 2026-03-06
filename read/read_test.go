package read

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// TestDetectLanguage
// ---------------------------------------------------------------------------

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		ext  string
		name string
	}{
		{".go", "Go"},
		{".py", "Python"},
		{".ts", "JavaScript"},
		{".rs", "Rust"},
		{".cs", "CSharp"},
		{".sh", "Shell"},
		{".html", "HTML"},
		{".sql", "SQL"},
		{".yaml", "YAML"},
		{".yml", "YAML"},
		{".xyz", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			lang := DetectLanguage(tt.ext)
			if lang.Name != tt.name {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.ext, lang.Name, tt.name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestFilterMinimalGo
// ---------------------------------------------------------------------------

func TestFilterMinimalGo(t *testing.T) {
	input := `package main

// This is a comment
import "fmt"

/*
Multi-line
comment
*/
func main() {
    // inline comment
    fmt.Println("hello")
}
`
	lang := DetectLanguage(".go")
	got := FilterMinimal(input, lang)

	mustContain := []string{"package main", `import "fmt"`, "func main()", `fmt.Println("hello")`}
	for _, s := range mustContain {
		if !strings.Contains(got, s) {
			t.Errorf("expected output to contain %q, got:\n%s", s, got)
		}
	}

	mustNotContain := []string{"// This is a comment", "// inline comment", "Multi-line", "/*", "*/"}
	for _, s := range mustNotContain {
		if strings.Contains(got, s) {
			t.Errorf("expected output NOT to contain %q, got:\n%s", s, got)
		}
	}
}

// ---------------------------------------------------------------------------
// TestFilterMinimalPython
// ---------------------------------------------------------------------------

func TestFilterMinimalPython(t *testing.T) {
	input := `# File comment
import os

class Foo:
    """This is a docstring."""
    def bar(self):
        # method comment
        pass
`
	lang := DetectLanguage(".py")
	got := FilterMinimal(input, lang)

	if !strings.Contains(got, `"""This is a docstring."""`) {
		t.Errorf("expected docstring to be preserved, got:\n%s", got)
	}
	if strings.Contains(got, "# File comment") {
		t.Errorf("expected '# File comment' to be removed, got:\n%s", got)
	}
	if strings.Contains(got, "# method comment") {
		t.Errorf("expected '# method comment' to be removed, got:\n%s", got)
	}
	if !strings.Contains(got, "import os") {
		t.Errorf("expected 'import os' to remain, got:\n%s", got)
	}
	if !strings.Contains(got, "class Foo:") {
		t.Errorf("expected 'class Foo:' to remain, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// TestFilterMinimalShell
// ---------------------------------------------------------------------------

func TestFilterMinimalShell(t *testing.T) {
	input := `#!/bin/bash
# This is a comment
echo "hello"
# Another comment
echo "world"
`
	lang := DetectLanguage(".sh")
	got := FilterMinimal(input, lang)

	if !strings.Contains(got, "#!/bin/bash") {
		t.Errorf("expected shebang to be preserved, got:\n%s", got)
	}
	if strings.Contains(got, "# This is a comment") {
		t.Errorf("expected comment to be removed, got:\n%s", got)
	}
	if strings.Contains(got, "# Another comment") {
		t.Errorf("expected comment to be removed, got:\n%s", got)
	}
	if !strings.Contains(got, `echo "hello"`) {
		t.Errorf("expected echo to remain, got:\n%s", got)
	}
	if !strings.Contains(got, `echo "world"`) {
		t.Errorf("expected echo to remain, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// TestFilterMinimalHTML
// ---------------------------------------------------------------------------

func TestFilterMinimalHTML(t *testing.T) {
	input := `<div>
  <!-- This is a comment -->
  <p>Hello</p>
  <!--
  Multi-line comment
  -->
  <p>World</p>
</div>
`
	lang := DetectLanguage(".html")
	got := FilterMinimal(input, lang)

	if strings.Contains(got, "<!-- This is a comment -->") {
		t.Errorf("expected inline comment to be removed, got:\n%s", got)
	}
	if strings.Contains(got, "Multi-line comment") {
		t.Errorf("expected multi-line comment to be removed, got:\n%s", got)
	}
	if !strings.Contains(got, "<p>Hello</p>") {
		t.Errorf("expected <p>Hello</p> to remain, got:\n%s", got)
	}
	if !strings.Contains(got, "<p>World</p>") {
		t.Errorf("expected <p>World</p> to remain, got:\n%s", got)
	}
	if !strings.Contains(got, "<div>") {
		t.Errorf("expected <div> to remain, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// TestFilterMinimalSQL
// ---------------------------------------------------------------------------

func TestFilterMinimalSQL(t *testing.T) {
	input := `-- Create users table
CREATE TABLE users (
    id INT PRIMARY KEY, /* auto-increment */
    name VARCHAR(100)
);
`
	lang := DetectLanguage(".sql")
	got := FilterMinimal(input, lang)

	if strings.Contains(got, "-- Create users table") {
		t.Errorf("expected SQL line comment to be removed, got:\n%s", got)
	}
	if strings.Contains(got, "/* auto-increment */") {
		t.Errorf("expected SQL block comment to be removed, got:\n%s", got)
	}
	if !strings.Contains(got, "CREATE TABLE users") {
		t.Errorf("expected CREATE TABLE to remain, got:\n%s", got)
	}
	if !strings.Contains(got, "name VARCHAR(100)") {
		t.Errorf("expected column definition to remain, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// TestFilterAggressive
// ---------------------------------------------------------------------------

func TestFilterAggressive(t *testing.T) {
	input := `package main

// Package comment
import (
    "fmt"
    "os"
)

/*
Multi-line comment
*/
func main() {
    // inline comment
    fmt.Println("hello")

    os.Exit(0)
}
`
	lang := DetectLanguage(".go")
	got := FilterAggressive(input, lang)

	// Comments removed
	if strings.Contains(got, "// Package comment") {
		t.Errorf("expected comment removed in aggressive mode, got:\n%s", got)
	}
	if strings.Contains(got, "// inline comment") {
		t.Errorf("expected inline comment removed, got:\n%s", got)
	}
	if strings.Contains(got, "Multi-line comment") {
		t.Errorf("expected block comment removed, got:\n%s", got)
	}

	// Imports removed
	if strings.Contains(got, `"fmt"`) {
		t.Errorf("expected imports removed in aggressive mode, got:\n%s", got)
	}
	if strings.Contains(got, `"os"`) {
		t.Errorf("expected imports removed in aggressive mode, got:\n%s", got)
	}

	// No consecutive blank lines
	if strings.Contains(got, "\n\n\n") {
		t.Errorf("expected no consecutive blank lines in aggressive mode, got:\n%s", got)
	}

	// Code preserved
	if !strings.Contains(got, `fmt.Println("hello")`) {
		t.Errorf("expected code to remain, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// TestFilterMinimalCollapsesBlankLines
// ---------------------------------------------------------------------------

func TestFilterMinimalCollapsesBlankLines(t *testing.T) {
	input := "line1\n\n\n\n\n\nline2\n\n\n\n\n\n\nline3\n"
	lang := DetectLanguage(".txt") // unknown / neutral
	got := FilterMinimal(input, lang)

	// Should not have more than 2 consecutive newlines (1 blank line)
	if strings.Contains(got, "\n\n\n") {
		t.Errorf("expected blank lines collapsed to at most 1, got:\n%q", got)
	}
	if !strings.Contains(got, "line1") || !strings.Contains(got, "line2") || !strings.Contains(got, "line3") {
		t.Errorf("expected all content lines preserved, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// TestSmartTruncation
// ---------------------------------------------------------------------------

func TestSmartTruncation(t *testing.T) {
	// Build a 100-line temp file
	var lines []string
	for i := 1; i <= 100; i++ {
		lines = append(lines, fmt.Sprintf("line %d content", i))
	}
	content := strings.Join(lines, "\n") + "\n"

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "bigfile.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, filtered, err := Run(tmpFile, "minimal", 10, false)
	if err != nil {
		t.Fatal(err)
	}
	got := filtered

	outLines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	// Should be approximately 10 lines (may include omission marker)
	if len(outLines) > 15 {
		t.Errorf("expected ~10 lines with truncation, got %d lines", len(outLines))
	}

	// Should contain some kind of omission marker
	if !strings.Contains(got, "...") && !strings.Contains(got, "omitted") && !strings.Contains(got, "truncated") && !strings.Contains(got, "lines") {
		t.Logf("warning: no omission marker found in truncated output:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// TestLineNumbers
// ---------------------------------------------------------------------------

func TestLineNumbers(t *testing.T) {
	content := "first line\nsecond line\nthird line\n"

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "numbered.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, filtered, err := Run(tmpFile, "minimal", 0, true)
	if err != nil {
		t.Fatal(err)
	}
	got := filtered

	// Verify line number pattern: digits followed by " | "
	lineNumPattern := regexp.MustCompile(`^\s*\d+\s*\|`)
	for _, line := range strings.Split(strings.TrimRight(got, "\n"), "\n") {
		if line == "" {
			continue
		}
		if !lineNumPattern.MatchString(line) {
			t.Errorf("expected line number pattern, got: %q", line)
		}
	}
}

// ---------------------------------------------------------------------------
// TestRunFileNotFound
// ---------------------------------------------------------------------------

func TestRunFileNotFound(t *testing.T) {
	_, _, err := Run("/nonexistent/path/to/file.go", "minimal", 0, false)
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestFilterMinimalRust
// ---------------------------------------------------------------------------

func TestFilterMinimalRust(t *testing.T) {
	input := `/// Documentation comment
// Regular comment
fn main() {
    // inline comment
    println!("hello");
}
`
	lang := DetectLanguage(".rs")
	got := FilterMinimal(input, lang)

	// Doc comments (///) should be preserved
	if !strings.Contains(got, "/// Documentation comment") {
		t.Errorf("expected Rust doc comment to be preserved, got:\n%s", got)
	}
	// Regular comments removed
	if strings.Contains(got, "// Regular comment") {
		t.Errorf("expected regular comment removed, got:\n%s", got)
	}
	if strings.Contains(got, "// inline comment") {
		t.Errorf("expected inline comment removed, got:\n%s", got)
	}
	if !strings.Contains(got, "fn main()") {
		t.Errorf("expected fn main to remain, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// TestFilterMinimalYAML
// ---------------------------------------------------------------------------

func TestFilterMinimalYAML(t *testing.T) {
	input := `# This is a YAML comment
name: my-app
# Another comment
version: "1.0"
services:
  # Service comment
  web:
    port: 8080
`
	lang := DetectLanguage(".yaml")
	got := FilterMinimal(input, lang)

	if strings.Contains(got, "# This is a YAML comment") {
		t.Errorf("expected YAML comment removed, got:\n%s", got)
	}
	if strings.Contains(got, "# Another comment") {
		t.Errorf("expected YAML comment removed, got:\n%s", got)
	}
	if strings.Contains(got, "# Service comment") {
		t.Errorf("expected YAML comment removed, got:\n%s", got)
	}
	if !strings.Contains(got, "name: my-app") {
		t.Errorf("expected content preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "port: 8080") {
		t.Errorf("expected content preserved, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// TestFilterMinimalCSS
// ---------------------------------------------------------------------------

func TestFilterMinimalCSS(t *testing.T) {
	input := `/* Reset styles */
body {
    margin: 0;
    /* padding reset */
    padding: 0;
}

/*
 * Component styles
 */
.header {
    color: red;
}
`
	lang := DetectLanguage(".css")
	got := FilterMinimal(input, lang)

	if strings.Contains(got, "/* Reset styles */") {
		t.Errorf("expected CSS comment removed, got:\n%s", got)
	}
	if strings.Contains(got, "/* padding reset */") {
		t.Errorf("expected inline CSS comment removed, got:\n%s", got)
	}
	if strings.Contains(got, "Component styles") {
		t.Errorf("expected multi-line CSS comment removed, got:\n%s", got)
	}
	if !strings.Contains(got, "margin: 0;") {
		t.Errorf("expected CSS rules preserved, got:\n%s", got)
	}
	if !strings.Contains(got, ".header") {
		t.Errorf("expected selector preserved, got:\n%s", got)
	}
}
