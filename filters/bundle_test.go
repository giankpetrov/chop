package filters

import (
	"strings"
	"testing"
)

func TestFilterBundleInstall(t *testing.T) {
	raw := "Fetching gem metadata from https://rubygems.org/...........\n" +
		"Resolving dependencies...\n" +
		"Installing rake 13.1.0\n" +
		"Installing concurrent-ruby 1.2.2\n" +
		"Installing i18n 1.14.1\n" +
		"Using bundler 2.4.22\n" +
		"Bundle complete! 15 Gemfile dependencies, 72 gems now installed.\n" +
		"Use `bundle info [gemname]` to see where a bundled gem is installed.\n"

	got, err := filterBundleInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "installed 3 gems") {
		t.Errorf("expected gem count, got: %s", got)
	}
	if !strings.Contains(got, "Bundle complete") {
		t.Error("expected complete message")
	}
}

func TestFilterBundleInstall_Empty(t *testing.T) {
	got, err := filterBundleInstall("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
