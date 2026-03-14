package filters

import (
	"fmt"
	"strings"
	"testing"
)

func TestFilterGrepManyFiles(t *testing.T) {
	// Generate grep output with 15+ files and many matches
	var lines []string
	for i := 1; i <= 18; i++ {
		file := fmt.Sprintf("src/module%d/handler.go", i)
		matchCount := 3 + i%5 // 3-7 matches per file
		for j := 1; j <= matchCount; j++ {
			lineNum := j * 10
			lines = append(lines, fmt.Sprintf("%s:%d:    fmt.Println(\"Processing item %d-%d\")", file, lineNum, i, j))
		}
	}
	raw := strings.Join(lines, "\n")

	got, err := filterGrep(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show max 10 files
	if strings.Contains(got, "module11") && !strings.Contains(got, "more files") {
		t.Errorf("should truncate after 10 files or show 'more files'")
	}

	// Should have summary
	if !strings.Contains(got, "matches across") {
		t.Errorf("expected summary line with match count, got:\n%s", got)
	}
	if !strings.Contains(got, "18 files") {
		t.Errorf("expected 18 files in summary, got:\n%s", got)
	}

	// Should mention extra files
	if !strings.Contains(got, "more files") {
		t.Errorf("expected 'more files' for truncated output, got:\n%s", got)
	}

	// Token savings >= 60%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 60.0 {
		t.Errorf("expected >=60%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterGrepFewFiles(t *testing.T) {
	raw := `src/main.go:10:    fmt.Println("hello")
src/main.go:20:    fmt.Println("world")
src/utils.go:5:    fmt.Println("util")`

	got, err := filterGrep(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With only 2 files and 3 matches, should passthrough
	if got != raw {
		t.Errorf("expected passthrough for few matches, got:\n%s", got)
	}
}

func TestFilterGrepEmpty(t *testing.T) {
	got, err := filterGrep("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestFilterGrepPlainOutput(t *testing.T) {
	// grep without filenames (piped input)
	raw := `hello world
this is a test
another line of output`

	got, err := filterGrep(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Few lines, should passthrough
	if got != raw {
		t.Errorf("expected passthrough for plain output with few lines")
	}
}

func TestFilterGrepPlainOutputNoExpand(t *testing.T) {
	// Plain output >10 lines should not expand (sanity check must apply)
	lines := make([]string, 12)
	for i := range lines {
		lines[i] = fmt.Sprintf("  line %d: some content here", i+1)
	}
	raw := strings.Join(lines, "\n")

	got, err := filterGrep(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	if filteredTokens > rawTokens {
		t.Errorf("filter expanded output: raw=%d tokens, filtered=%d tokens", rawTokens, filteredTokens)
	}
}

func TestFilterGrepWithAnsi(t *testing.T) {
	raw := "\x1b[32msrc/main.go\x1b[0m:\x1b[33m10\x1b[0m:    fmt.Println(\"hello\")\n" +
		"\x1b[32msrc/main.go\x1b[0m:\x1b[33m20\x1b[0m:    fmt.Println(\"world\")"

	got, err := filterGrep(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not contain ANSI codes
	if strings.Contains(got, "\x1b[") {
		t.Errorf("should strip ANSI codes from output")
	}
}

func TestFilterGrepMaxMatchesPerFile(t *testing.T) {
	var lines []string
	for i := 1; i <= 10; i++ {
		lines = append(lines, fmt.Sprintf("src/big_file.go:%d:    match line %d", i*10, i))
	}
	for i := 1; i <= 5; i++ {
		lines = append(lines, fmt.Sprintf("src/other.go:%d:    match line %d", i*10, i))
	}
	for i := 1; i <= 5; i++ {
		lines = append(lines, fmt.Sprintf("src/third.go:%d:    match line %d", i*10, i))
	}
	raw := strings.Join(lines, "\n")

	got, err := filterGrep(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show file groupings
	if !strings.Contains(got, "big_file.go") {
		t.Errorf("expected big_file.go in output")
	}
	if !strings.Contains(got, "10 matches") {
		t.Errorf("expected match count for big_file.go, got:\n%s", got)
	}
	// Should truncate to 2 matches per file
	if !strings.Contains(got, "+8 more") {
		t.Errorf("expected '+8 more' for truncated matches, got:\n%s", got)
	}
}
