package filters

import (
	"fmt"
	"strings"
	"testing"
)

func TestFilterGlabMrListEmpty(t *testing.T) {
	got, err := filterGlabMrList("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no merge requests" {
		t.Errorf("expected 'no merge requests', got %q", got)
	}
}

func TestFilterGlabMrListFew(t *testing.T) {
	raw := `MR    TITLE                           BRANCH          CREATED_AT
!42   Fix authentication bug          fix/auth-bug    about 2 hours ago
!41   Add user profile endpoint       feat/profile    about 5 hours ago
!40   Update dependencies             chore/deps      2 days ago`

	got, err := filterGlabMrList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// <=5 MRs — all shown (passthrough or reformatted, not truncated)
	if !strings.Contains(got, "!42") {
		t.Errorf("expected '!42' in output")
	}
	if !strings.Contains(got, "!41") {
		t.Errorf("expected '!41' in output")
	}
	if !strings.Contains(got, "!40") {
		t.Errorf("expected '!40' in output")
	}
	if strings.Contains(got, "more") {
		t.Errorf("should not truncate when <=5 MRs, got:\n%s", got)
	}
}

func TestFilterGlabMrListMany(t *testing.T) {
	var lines []string
	lines = append(lines, "MR    TITLE                           BRANCH          CREATED_AT")
	for i := 1; i <= 10; i++ {
		lines = append(lines, fmt.Sprintf("!%d   Fix issue %d                    fix/issue-%d     about %d hours ago", 100-i, i, i, i))
	}
	raw := strings.Join(lines, "\n")

	got, err := filterGlabMrList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show first 5 and "and 5 more"
	if !strings.Contains(got, "5 more") {
		t.Errorf("expected '5 more' in output, got:\n%s", got)
	}

	// Should show total count
	if !strings.Contains(got, "total: 10") {
		t.Errorf("expected 'total: 10' in output, got:\n%s", got)
	}

	// Should show first MR
	if !strings.Contains(got, "!99") {
		t.Errorf("expected first MR '!99' in output, got:\n%s", got)
	}

	// Should not show last MR (it would be beyond the first 5)
	if strings.Contains(got, "!90") {
		t.Errorf("should not show 10th MR when truncated, got:\n%s", got)
	}
}

func TestFilterGlabIssueListEmpty(t *testing.T) {
	got, err := filterGlabIssueList("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no issues" {
		t.Errorf("expected 'no issues', got %q", got)
	}
}

func TestFilterGlabIssueListMany(t *testing.T) {
	var lines []string
	lines = append(lines, "#   TITLE                           LABELS      CREATED_AT")
	for i := 1; i <= 8; i++ {
		lines = append(lines, fmt.Sprintf("#%d  Issue title %d                 bug         about %d days ago", i, i, i))
	}
	raw := strings.Join(lines, "\n")

	got, err := filterGlabIssueList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 8 issues > 5 threshold — should truncate to 5 + count
	if !strings.Contains(got, "more") {
		t.Errorf("expected truncation indicator in output, got:\n%s", got)
	}
	if !strings.Contains(got, "8") {
		t.Errorf("expected total count 8 in output, got:\n%s", got)
	}

	// First issue should be shown
	if !strings.Contains(got, "#1") {
		t.Errorf("expected '#1' in output, got:\n%s", got)
	}
}

func TestFilterGlabCiStatusAllPassed(t *testing.T) {
	raw := `• Pipeline #12345 triggered 5 minutes ago by user@example.com
• Status: passed

STAGE     NAME            STATUS    DURATION
build     compile         passed    1m 23s
build     lint            passed    45s
test      unit-tests      passed    2m 11s
deploy    staging         passed    30s`

	got, err := filterGlabCiStatus(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All jobs passed — all shown (short output)
	if !strings.Contains(got, "compile") {
		t.Errorf("expected 'compile' job in output")
	}
	if !strings.Contains(got, "lint") {
		t.Errorf("expected 'lint' job in output")
	}
	if !strings.Contains(got, "unit-tests") {
		t.Errorf("expected 'unit-tests' job in output")
	}
	if !strings.Contains(got, "staging") {
		t.Errorf("expected 'staging' job in output")
	}
}

func TestFilterGlabCiStatusWithFailure(t *testing.T) {
	raw := `• Pipeline #12345 triggered 5 minutes ago by user@example.com
• Status: running

STAGE     NAME            STATUS    DURATION
build     compile         passed    1m 23s
build     lint            passed    45s
test      unit-tests      failed    2m 11s
deploy    staging         pending`

	got, err := filterGlabCiStatus(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Failed job must appear
	if !strings.Contains(got, "unit-tests") {
		t.Errorf("expected failed job 'unit-tests' in output")
	}
	if !strings.Contains(got, "failed") {
		t.Errorf("expected 'failed' status in output")
	}

	// Failed job must appear before passing jobs (first in output)
	unitTestIdx := strings.Index(got, "unit-tests")
	compileIdx := strings.Index(got, "compile")
	if compileIdx != -1 && unitTestIdx > compileIdx {
		t.Errorf("expected failed job 'unit-tests' to appear before passing job 'compile', got:\n%s", got)
	}
}

func TestFilterGlabSanityCheck(t *testing.T) {
	// Large MR list
	var lines []string
	lines = append(lines, "MR    TITLE                                 BRANCH           CREATED_AT")
	for i := 1; i <= 50; i++ {
		lines = append(lines, fmt.Sprintf("!%d   Some merge request title here %d   feat/branch-%d   about %d days ago", i, i, i, i))
	}
	raw := strings.Join(lines, "\n")

	got, err := filterGlabMrList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) > len(raw) {
		t.Errorf("output longer than input: raw=%d bytes, filtered=%d bytes", len(raw), len(got))
	}
}
