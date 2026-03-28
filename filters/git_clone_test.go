package filters

import (
	"strings"
	"testing"
)

var gitCloneFixture = `Cloning into 'myrepo'...
remote: Enumerating objects: 1234, done.
remote: Counting objects: 100% (1234/1234), done.
remote: Compressing objects: 100% (567/567), done.
remote: Total 1234 (delta 456), reused 789 (delta 123), pack-reused 0
Receiving objects: 100% (1234/1234), 5.67 MiB | 10.23 MiB/s, done.
Resolving deltas: 100% (456/456), done.
`

func TestFilterGitClone(t *testing.T) {
	got, err := filterGitClone(gitCloneFixture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should keep "Cloning into"
	if !strings.Contains(got, "Cloning into 'myrepo'") {
		t.Errorf("expected 'Cloning into' preserved, got:\n%s", got)
	}

	// Should keep "remote: Total"
	if !strings.Contains(got, "remote: Total") {
		t.Errorf("expected 'remote: Total' preserved, got:\n%s", got)
	}

	// Should keep "Resolving deltas: ... done"
	if !strings.Contains(got, "Resolving deltas:") {
		t.Errorf("expected 'Resolving deltas:' preserved, got:\n%s", got)
	}

	// Should drop "remote: Enumerating"
	if strings.Contains(got, "remote: Enumerating") {
		t.Errorf("expected 'remote: Enumerating' dropped, got:\n%s", got)
	}

	// Should drop "remote: Counting"
	if strings.Contains(got, "remote: Counting") {
		t.Errorf("expected 'remote: Counting' dropped, got:\n%s", got)
	}

	// Should drop "remote: Compressing"
	if strings.Contains(got, "remote: Compressing") {
		t.Errorf("expected 'remote: Compressing' dropped, got:\n%s", got)
	}

	// Should drop "Receiving objects:"
	if strings.Contains(got, "Receiving objects:") {
		t.Errorf("expected 'Receiving objects:' dropped, got:\n%s", got)
	}
}

func TestFilterGitCloneEmpty(t *testing.T) {
	got, err := filterGitClone("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestGitCloneRouted(t *testing.T) {
	if get("git", []string{"clone"}) == nil {
		t.Error("expected non-nil filter for git clone")
	}
}
