package filters

import (
	"strings"
	"testing"
)

func TestFilterGhPrList(t *testing.T) {
	raw := "123\tFix login bug\tfix/login\t2026-03-01T10:00:00Z\n" +
		"124\tAdd dashboard feature\tfeat/dashboard\t2026-03-02T11:00:00Z\n" +
		"125\tRefactor auth module\trefactor/auth\t2026-03-03T09:00:00Z\n" +
		"126\tUpdate dependencies\tchore/deps\t2026-03-03T12:00:00Z\n" +
		"127\tFix CI pipeline\tfix/ci\t2026-03-04T08:00:00Z\n"

	got, err := filterGhPrList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d:\n%s", len(lines), got)
	}

	if !strings.Contains(got, "#123") {
		t.Errorf("expected PR number #123 in output")
	}
	if !strings.Contains(got, "Fix login bug") {
		t.Errorf("expected title in output")
	}
	// Should strip timestamps
	if strings.Contains(got, "2026-03") {
		t.Errorf("should not contain timestamps")
	}
}

func TestFilterGhPrViewVerbose(t *testing.T) {
	raw := `title:	Add new authentication flow
state:	OPEN
author:	johndoe
labels:	enhancement, security, priority:high
assignees:	janedoe, bobsmith
reviewers:	alice, charlie
milestone:	v2.0
project:	Backend Improvements
number:	456
url:	https://github.com/org/repo/pull/456
head:	feat/new-auth
base:	main
additions:	342
deletions:	89
changed files:	12
review decision:	CHANGES_REQUESTED
created:	2026-02-28T10:15:00Z
updated:	2026-03-04T16:30:00Z
--
body
--
This PR implements a completely new authentication flow using OAuth2.

## Changes
- Added OAuth2 provider integration
- Refactored token management
- Updated middleware to support new auth headers
- Added comprehensive tests for all auth endpoints

## Testing
- Unit tests added for all new modules
- Integration tests with mock OAuth provider
- Manual testing against staging environment

## Screenshots
[screenshot of login page]
[screenshot of oauth flow]

Closes #400, #401, #402
`

	got, err := filterGhPrView(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Add new authentication flow") {
		t.Errorf("expected title in output")
	}
	if !strings.Contains(got, "OPEN") {
		t.Errorf("expected state in output")
	}
	if !strings.Contains(got, "johndoe") {
		t.Errorf("expected author in output")
	}
	if !strings.Contains(got, "feat/new-auth") {
		t.Errorf("expected branch in output")
	}
	if !strings.Contains(got, "enhancement") {
		t.Errorf("expected labels in output")
	}
	if !strings.Contains(got, "12 files") {
		t.Errorf("expected changed files count in output")
	}
	if !strings.Contains(got, "CHANGES_REQUESTED") {
		t.Errorf("expected review status in output")
	}

	// Should NOT contain verbose body content
	if strings.Contains(got, "Screenshots") {
		t.Errorf("should not contain full body sections")
	}
	if strings.Contains(got, "Closes #400") {
		t.Errorf("should not contain body references")
	}

	// Token savings >= 60%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 60.0 {
		t.Errorf("expected >=60%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterGhPrChecksWithFailures(t *testing.T) {
	raw := `build (ubuntu-latest)	pass	2m30s	https://github.com/org/repo/runs/1
build (macos-latest)	pass	3m15s	https://github.com/org/repo/runs/2
lint	pass	1m05s	https://github.com/org/repo/runs/3
test (unit)	pass	4m20s	https://github.com/org/repo/runs/4
test (integration)	fail	5m10s	https://github.com/org/repo/runs/5
deploy-preview	pass	2m00s	https://github.com/org/repo/runs/6
security-scan	fail	1m30s	https://github.com/org/repo/runs/7
codecov	pass	0m45s	https://github.com/org/repo/runs/8
`

	got, err := filterGhPrChecks(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "FAILING") {
		t.Errorf("expected FAILING header in output")
	}
	if !strings.Contains(got, "integration") {
		t.Errorf("expected failed test (integration) in output")
	}
	if !strings.Contains(got, "security-scan") {
		t.Errorf("expected failed security-scan in output")
	}
	if !strings.Contains(got, "6 passed") {
		t.Errorf("expected passed count in summary, got:\n%s", got)
	}
	if !strings.Contains(got, "2 failed") {
		t.Errorf("expected failed count in summary, got:\n%s", got)
	}
}

func TestFilterGhPrChecksAllPassing(t *testing.T) {
	raw := `build	pass	2m30s	https://github.com/org/repo/runs/1
lint	pass	1m05s	https://github.com/org/repo/runs/2
test	pass	4m20s	https://github.com/org/repo/runs/3
`

	got, err := filterGhPrChecks(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(got, "FAILING") {
		t.Errorf("should not have FAILING section when all pass")
	}
	if !strings.Contains(got, "3 passed") {
		t.Errorf("expected 3 passed in summary, got:\n%s", got)
	}
}

func TestFilterGhPrListEmpty(t *testing.T) {
	got, err := filterGhPrList("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no pull requests" {
		t.Errorf("expected 'no pull requests', got %q", got)
	}
}
