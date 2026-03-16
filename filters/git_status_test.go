package filters

import (
	"strings"
	"testing"
)

func countTokens(s string) int {
	return len(strings.Fields(s))
}

func TestGitStatusClean(t *testing.T) {
	raw := `On branch main
Your branch is up to date with 'origin/main'.

nothing to commit, working tree clean
`
	got, err := filterGitStatus(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != "clean" {
		t.Errorf("expected 'clean', got %q", got)
	}
}

func TestGitStatusUnstagedOnly(t *testing.T) {
	raw := `On branch feature/login
Your branch is up to date with 'origin/feature/login'.

Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
	modified:   src/app.ts
	modified:   src/auth/login.ts
	deleted:    src/old-config.json

Untracked files:
  (use "git add <file>..." to include in what will be committed)
	src/new-feature.ts
	docs/notes.md

no changes added to commit (use "git add" to track)
`
	got, err := filterGitStatus(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "unstaged(3)") {
		t.Errorf("expected unstaged count 3, got: %s", got)
	}
	if !strings.Contains(got, "untracked(2)") {
		t.Errorf("expected untracked count 2, got: %s", got)
	}
	if strings.Contains(got, "\nstaged(") || strings.HasPrefix(got, "staged(") {
		t.Errorf("should not have staged section, got: %s", got)
	}

	// Verify token savings
	rawTokens := countTokens(raw)
	filteredTokens := countTokens(got)
	savings := 100.0 - (float64(filteredTokens) / float64(rawTokens) * 100.0)
	if savings < 60.0 {
		t.Errorf("expected >=60%% savings, got %.1f%%", savings)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
}

func TestGitStatusStagedAndUnstaged(t *testing.T) {
	raw := `On branch main
Changes to be committed:
  (use "git restore --staged <file>..." to unstage)
	new file:   src/feature.go
	new file:   src/feature_test.go

Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
	modified:   README.md
`
	got, err := filterGitStatus(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "staged(2)") {
		t.Errorf("expected staged count 2, got: %s", got)
	}
	if !strings.Contains(got, "unstaged(1)") {
		t.Errorf("expected unstaged count 1, got: %s", got)
	}
}

func TestGitStatusStagedOnly(t *testing.T) {
	raw := `On branch main
Changes to be committed:
  (use "git restore --staged <file>..." to unstage)
	modified:   src/app.ts
	new file:   src/new.go
`
	got, err := filterGitStatus(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "staged(2)") {
		t.Errorf("expected staged count 2, got: %s", got)
	}
	if strings.Contains(got, "unstaged(") {
		t.Errorf("should not have unstaged section, got: %s", got)
	}
	if !strings.Contains(got, "(new)") {
		t.Errorf("expected (new) marker, got: %s", got)
	}
}

func TestGitStatusAllSections(t *testing.T) {
	raw := `On branch main
Your branch is up to date with 'origin/main'.

Changes to be committed:
  (use "git restore --staged <file>..." to unstage)
	modified:   config/config.go
	modified:   config/config_test.go

Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
	modified:   README.md
	modified:   main.go
	modified:   hooks/hook.go

Untracked files:
  (use "git add <file>..." to include in what will be committed)
	.openchop.yml
`
	got, err := filterGitStatus(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "staged(2)") {
		t.Errorf("expected staged count 2, got: %s", got)
	}
	if !strings.Contains(got, "unstaged(3)") {
		t.Errorf("expected unstaged count 3, got: %s", got)
	}
	if !strings.Contains(got, "untracked(1)") {
		t.Errorf("expected untracked count 1, got: %s", got)
	}

	// Verify order: staged before unstaged before untracked
	stagedIdx := strings.Index(got, "staged(")
	unstagedIdx := strings.Index(got, "unstaged(")
	untrackedIdx := strings.Index(got, "untracked(")
	if stagedIdx > unstagedIdx || unstagedIdx > untrackedIdx {
		t.Errorf("sections should be in order staged/unstaged/untracked, got: %s", got)
	}

	t.Logf("output: %s", got)
}

func TestGitStatusDeletedFile(t *testing.T) {
	raw := `On branch main
Changes not staged for commit:
  (use "git add/rm <file>..." to update what will be committed)
	deleted:    old-file.txt
`
	got, err := filterGitStatus(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "(deleted)") {
		t.Errorf("expected (deleted) marker, got: %s", got)
	}
}
