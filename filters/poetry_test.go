package filters

import (
	"fmt"
	"strings"
	"testing"
)

func TestFilterPoetryInstallEmpty(t *testing.T) {
	got, err := filterPoetryInstall("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected passthrough for empty input, got %q", got)
	}
}

func TestFilterPoetryInstallNoChanges(t *testing.T) {
	raw := `Installing dependencies from Pipfile.lock...
All packages are already up to date.`

	got, err := filterPoetryInstall(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return "up to date" or passthrough — never expand
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	if filteredTokens > rawTokens {
		t.Errorf("filter expanded output: raw=%d tokens, filtered=%d tokens", rawTokens, filteredTokens)
	}
}

func TestFilterPoetryInstallWithUpdates(t *testing.T) {
	raw := `Creating virtualenv myapp-abc123-py3.11 in /home/user/.cache/pypoetry/virtualenvs
Updating dependencies
Resolving dependencies... (2.3s)

Package operations: 10 installs, 2 updates, 0 removals

  • Updating requests (2.28.0 -> 2.31.0)
  • Updating urllib3 (1.26.0 -> 2.0.4)
  • Installing certifi (2023.7.22)
  • Installing charset-normalizer (3.3.0)
  • Installing idna (3.4)
  • Installing pycparser (2.21)
  • Installing cffi (1.15.1)
  • Installing cryptography (41.0.3)
  • Installing pyOpenSSL (23.2.0)
  • Installing requests-toolbelt (1.0.0)
  • Installing six (1.16.0)
  • Installing urllib3 (2.0.4)

Writing lock file`

	got, err := filterPoetryInstall(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show ops summary
	if !strings.Contains(got, "Package operations:") {
		t.Errorf("expected ops summary line, got:\n%s", got)
	}
	// Should show updates
	if !strings.Contains(got, "Updating requests") {
		t.Errorf("expected update for requests, got:\n%s", got)
	}
	if !strings.Contains(got, "Updating urllib3") {
		t.Errorf("expected update for urllib3, got:\n%s", got)
	}
	// Should truncate installs with "more installs"
	if !strings.Contains(got, "more installs") {
		t.Errorf("expected truncation notice for installs, got:\n%s", got)
	}

	// Token savings >= 50%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterPoetryInstallOnlyInstalls(t *testing.T) {
	var pkgLines []string
	pkgNames := []string{
		"certifi", "charset-normalizer", "idna", "pycparser", "cffi",
		"cryptography", "pyOpenSSL", "requests-toolbelt", "six", "urllib3",
		"packaging", "pyparsing", "tomli", "typing-extensions", "zipp",
	}
	for _, name := range pkgNames {
		pkgLines = append(pkgLines, fmt.Sprintf("  • Installing %s (1.0.0)", name))
	}
	raw := "Package operations: 15 installs, 0 updates, 0 removals\n\n" +
		strings.Join(pkgLines, "\n") + "\n\nWriting lock file"

	got, err := filterPoetryInstall(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show ops line
	if !strings.Contains(got, "Package operations:") {
		t.Errorf("expected ops summary, got:\n%s", got)
	}
	// Should show first installs
	if !strings.Contains(got, "certifi") {
		t.Errorf("expected first install shown, got:\n%s", got)
	}
	// Should truncate the rest
	if !strings.Contains(got, "more installs") {
		t.Errorf("expected 'more installs' truncation, got:\n%s", got)
	}
}

func TestFilterPoetryShowFewPackages(t *testing.T) {
	raw := `certifi          2023.7.22 Python package for providing Mozilla's CA Bundle.
charset-normalizer 3.3.0   The Real First Universal Charset Detector.
requests         2.31.0    Python HTTP for Humans.
urllib3          2.0.4     HTTP library with thread-safe connection pooling.
idna             3.4       Internationalized Domain Names in Applications (IDNA).`

	got, err := filterPoetryShow(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5 packages is <= 10, should passthrough
	if got != raw {
		t.Errorf("expected passthrough for few packages, got:\n%s", got)
	}
}

func TestFilterPoetryShowManyPackages(t *testing.T) {
	var lines []string
	pkgs := []string{
		"certifi", "charset-normalizer", "requests", "urllib3", "idna",
		"pycparser", "cffi", "cryptography", "pyOpenSSL", "six",
		"packaging", "pyparsing", "tomli", "typing-extensions", "zipp",
		"attrs", "click", "colorama", "filelock", "iniconfig",
		"pluggy", "py", "pytest", "setuptools", "wheel",
	}
	for _, pkg := range pkgs {
		lines = append(lines, fmt.Sprintf("%-25s 1.0.0  A description for %s package.", pkg, pkg))
	}
	raw := strings.Join(lines, "\n")

	got, err := filterPoetryShow(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 25 packages > 10, should compress
	if got == raw {
		t.Errorf("expected compression for 25 packages, got passthrough")
	}
	// Should contain count
	if !strings.Contains(got, "25") {
		t.Errorf("expected package count (25) in output, got:\n%s", got)
	}
}

func TestFilterPoetrySanityCheck(t *testing.T) {
	// Install output
	raw := `Package operations: 3 installs, 1 update, 0 removals

  • Updating requests (2.28.0 -> 2.31.0)
  • Installing certifi (2023.7.22)
  • Installing charset-normalizer (3.3.0)
  • Installing idna (3.4)

Writing lock file`

	got, err := filterPoetryInstall(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) > len(raw) {
		t.Errorf("filter expanded output: raw=%d bytes, filtered=%d bytes", len(raw), len(got))
	}

	// Show output
	var showLines []string
	for i := 0; i < 20; i++ {
		showLines = append(showLines, fmt.Sprintf("package%-2d  1.0.%d  Description for package %d.", i, i, i))
	}
	rawShow := strings.Join(showLines, "\n")

	gotShow, err := filterPoetryShow(rawShow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotShow) > len(rawShow) {
		t.Errorf("show filter expanded output: raw=%d bytes, filtered=%d bytes", len(rawShow), len(gotShow))
	}
}
