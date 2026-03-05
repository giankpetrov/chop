package filters

import (
	"strings"
	"testing"
)

func TestFilterGhIssueList(t *testing.T) {
	raw := "100\tLogin page crashes on mobile\tbug, priority:high\t2026-03-01T10:00:00Z\n" +
		"101\tAdd dark mode support\tenhancement\t2026-03-02T11:00:00Z\n" +
		"102\tUpdate documentation for v2\tdocs\t2026-03-03T09:00:00Z\n" +
		"103\tPerformance regression in search\tbug, performance\t2026-03-03T12:00:00Z\n" +
		"104\tRefactor database layer\trefactor\t2026-03-04T08:00:00Z\n"

	got, err := filterGhIssueList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d:\n%s", len(lines), got)
	}

	if !strings.Contains(got, "#100") {
		t.Errorf("expected issue number #100 in output")
	}
	if !strings.Contains(got, "Login page crashes on mobile") {
		t.Errorf("expected title in output")
	}
	if !strings.Contains(got, "bug, priority:high") {
		t.Errorf("expected labels in output")
	}

	// Should not contain timestamps
	if strings.Contains(got, "2026-03") {
		t.Errorf("should not contain timestamps")
	}
}

func TestFilterGhIssueView(t *testing.T) {
	raw := `title:	Implement OAuth2 authentication
state:	OPEN
author:	janedoe
labels:	enhancement, security
assignees:	johndoe
milestone:	v2.0
number:	200
url:	https://github.com/org/repo/issues/200
created:	2026-02-15T08:00:00Z
updated:	2026-03-04T10:00:00Z
--
body
--
We need to implement OAuth2 authentication to replace our current session-based auth.

## Requirements
- Support Google and GitHub as OAuth providers
- Implement token refresh flow
- Add PKCE for public clients
- Migrate existing users seamlessly

## Acceptance Criteria
- [ ] Users can log in with Google
- [ ] Users can log in with GitHub
- [ ] Existing sessions are migrated
- [ ] Token refresh works automatically

## References
- RFC 6749
- https://oauth.net/2/
`

	got, err := filterGhIssueView(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Implement OAuth2 authentication") {
		t.Errorf("expected title in output")
	}
	if !strings.Contains(got, "OPEN") {
		t.Errorf("expected state in output")
	}
	if !strings.Contains(got, "janedoe") {
		t.Errorf("expected author in output")
	}
	if !strings.Contains(got, "enhancement") {
		t.Errorf("expected labels in output")
	}

	// Should not contain full body
	if strings.Contains(got, "Acceptance Criteria") {
		t.Errorf("should not contain full body")
	}
	if strings.Contains(got, "RFC 6749") {
		t.Errorf("should not contain references")
	}
}

func TestFilterGhIssueListEmpty(t *testing.T) {
	got, err := filterGhIssueList("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no issues" {
		t.Errorf("expected 'no issues', got %q", got)
	}
}
