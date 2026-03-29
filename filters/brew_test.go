package filters

import (
	"strings"
	"testing"
)

var brewFixture = `==> Downloading https://formulae.brew.sh/api/formula.jws.json
Already downloaded: /Users/user/Library/Caches/Homebrew/api/formula.jws.json
==> Fetching dependencies for wget: libunistring, libidn2 and openssl@3
==> Fetching wget
==> Downloading https://ghcr.io/v2/homebrew/core/wget/manifests/1.21.4
######################################################################### 100.0%
==> Downloading https://ghcr.io/v2/homebrew/core/wget/blobs/sha256:abcdef1234567890
######################################################################### 100.0%
==> Installing dependencies for wget: libunistring, libidn2 and openssl@3
==> Installing wget dependency: libunistring
🍺  /usr/local/Cellar/libunistring/1.1: 56 files, 3MB
==> Installing wget dependency: libidn2
🍺  /usr/local/Cellar/libidn2/2.3.4: 78 files, 2.5MB
==> Installing wget
🍺  /usr/local/Cellar/wget/1.21.4: 89 files, 4.1MB
==> Running brew cleanup wget...
`

func TestFilterBrew(t *testing.T) {
	got, err := filterBrew(brewFixture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should keep section headers
	if !strings.Contains(got, "==> Installing wget") {
		t.Errorf("expected '==> Installing wget' preserved, got:\n%s", got)
	}

	// Should keep beer emoji lines
	if !strings.Contains(got, "🍺") {
		t.Errorf("expected beer emoji lines preserved, got:\n%s", got)
	}

	// Should drop download lines
	if strings.Contains(got, "Downloading") {
		t.Errorf("expected 'Downloading' lines dropped, got:\n%s", got)
	}

	// Should drop progress bar lines
	if strings.Contains(got, "###") {
		t.Errorf("expected progress bar lines dropped, got:\n%s", got)
	}

	// Should drop sha256 checksum lines
	if strings.Contains(got, "sha256") {
		t.Errorf("expected sha256 lines dropped, got:\n%s", got)
	}
}

func TestFilterBrewEmpty(t *testing.T) {
	got, err := filterBrew("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFilterBrewAlreadyInstalled(t *testing.T) {
	raw := "Warning: wget 1.21.4 is already installed and up-to-date.\nTo reinstall 1.21.4, run:\n  brew reinstall wget\n"
	got, err := filterBrew(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "already installed") {
		t.Errorf("expected 'already installed' preserved, got:\n%s", got)
	}
}

func TestBrewRouted(t *testing.T) {
	for _, sub := range []string{"install", "upgrade", "reinstall", "update"} {
		if get("brew", []string{sub}) == nil {
			t.Errorf("expected non-nil filter for brew %s", sub)
		}
	}
}
