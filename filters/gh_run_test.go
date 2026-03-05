package filters

import (
	"strings"
	"testing"
)

func TestFilterGhRunList(t *testing.T) {
	raw := "completed\tsuccess\tCI Build\tCI\tmain\tpush\t12345678\t3m20s\t2 hours ago\n" +
		"completed\tfailure\tCI Build\tCI\tfeat/auth\tpush\t12345679\t5m10s\t1 hour ago\n" +
		"completed\tsuccess\tDeploy\tDeploy\tmain\tpush\t12345680\t2m00s\t30 minutes ago\n" +
		"in_progress\t\tCI Build\tCI\tfix/bug\tpush\t12345681\t1m30s\t5 minutes ago\n"

	got, err := filterGhRunList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d:\n%s", len(lines), got)
	}

	if !strings.Contains(got, "12345678") {
		t.Errorf("expected run ID in output")
	}
	if !strings.Contains(got, "success") {
		t.Errorf("expected conclusion in output")
	}
	if !strings.Contains(got, "(main)") {
		t.Errorf("expected branch in output")
	}
}

func TestFilterGhRunViewWithFailures(t *testing.T) {
	raw := `workflow: CI Pipeline
status: completed
conclusion: failure
event: push
branch: feat/new-feature
sha: abc123def456
url: https://github.com/org/repo/actions/runs/123456
created: 2026-03-04T10:00:00Z
updated: 2026-03-04T10:15:00Z

jobs:
  build (ubuntu-latest)
    status: completed
    conclusion: success
    steps:
      - Checkout        success  5s
      - Setup Node      success  12s
      - Install deps    success  45s
      - Build           success  1m30s

  test (ubuntu-latest)
    status: completed
    conclusion: failure
    steps:
      - Checkout        success  5s
      - Setup Node      success  12s
      - Install deps    success  45s
      - Run tests       fail     2m10s
      - Upload coverage success  8s

  lint
    status: completed
    conclusion: success
    steps:
      - Checkout        success  5s
      - Setup Node      success  12s
      - Run lint        success  30s

  deploy-preview
    status: completed
    conclusion: success
    steps:
      - Checkout        success  5s
      - Build           success  1m00s
      - Deploy          success  45s
`

	got, err := filterGhRunView(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "CI Pipeline") {
		t.Errorf("expected workflow name in output")
	}
	if !strings.Contains(got, "failure") {
		t.Errorf("expected failure status in output")
	}
	if !strings.Contains(got, "FAILED") {
		t.Errorf("expected FAILED STEPS section, got:\n%s", got)
	}
	if !strings.Contains(got, "fail") {
		t.Errorf("expected failed step details in output")
	}

	// Token savings >= 50%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterGhRunViewSuccess(t *testing.T) {
	raw := `workflow: CI Pipeline
status: completed
conclusion: success
`

	got, err := filterGhRunView(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "CI Pipeline") {
		t.Errorf("expected workflow name")
	}
	if !strings.Contains(got, "success") {
		t.Errorf("expected success status")
	}
	if strings.Contains(got, "FAILED") {
		t.Errorf("should not have FAILED section for successful run")
	}
}

func TestFilterGhRunListEmpty(t *testing.T) {
	got, err := filterGhRunList("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no workflow runs" {
		t.Errorf("expected 'no workflow runs', got %q", got)
	}
}
