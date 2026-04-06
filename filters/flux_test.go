package filters

import (
	"strings"
	"testing"
)

func TestFilterFluxGetEmpty(t *testing.T) {
	got, err := filterFluxGet("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestFilterFluxGetAllReady(t *testing.T) {
	raw := `NAME            READY   MESSAGE                         REVISION        SUSPENDED
flux-system     True    Applied revision: main/abc1234  main/abc1234    False
myapp           True    Applied revision: main/def5678  main/def5678    False
other-app       True    Applied revision: main/xyz0001  main/xyz0001    False`

	got, err := filterFluxGet(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "NAME") {
		t.Errorf("expected NAME column in output, got:\n%s", got)
	}
	if !strings.Contains(got, "READY") {
		t.Errorf("expected READY column in output, got:\n%s", got)
	}
	if !strings.Contains(got, "MESSAGE") {
		t.Errorf("expected MESSAGE column in output, got:\n%s", got)
	}
	// Should drop REVISION and SUSPENDED columns
	if strings.Contains(got, "REVISION") {
		t.Errorf("should drop REVISION column, got:\n%s", got)
	}
}

func TestFilterFluxGetWithNotReady(t *testing.T) {
	raw := `NAME            READY   MESSAGE                         REVISION        SUSPENDED
flux-system     True    Applied revision: main/abc1234  main/abc1234    False
myapp           True    Applied revision: main/def5678  main/def5678    False
other-app       False   kustomize build failed          main/ghi9012    False`

	got, err := filterFluxGet(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The not-ready entry should be annotated
	if !strings.Contains(got, "NOT READY") {
		t.Errorf("expected [NOT READY] annotation for failed item, got:\n%s", got)
	}
	if !strings.Contains(got, "other-app") {
		t.Errorf("expected other-app in output, got:\n%s", got)
	}
}

func TestFilterFluxReconcileSuccess(t *testing.T) {
	raw := `► annotating GitRepository flux-system in flux-system namespace
◎ waiting for GitRepository reconciliation
✔ GitRepository reconciliation completed
► annotating Kustomization flux-system in flux-system namespace
◎ waiting for Kustomization reconciliation
✔ Kustomization reconciliation completed
✔ applied revision main/abc1234`

	got, err := filterFluxReconcile(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// On success, should show ✔ lines
	if !strings.Contains(got, "✔") {
		t.Errorf("expected ✔ success lines in output, got:\n%s", got)
	}
	// Should drop annotating/waiting lines
	if strings.Contains(got, "annotating") {
		t.Errorf("should drop annotating lines on success, got:\n%s", got)
	}
}

func TestFilterFluxReconcileFailure(t *testing.T) {
	raw := `► annotating GitRepository flux-system in flux-system namespace
◎ waiting for GitRepository reconciliation
✔ GitRepository reconciliation completed
► annotating Kustomization myapp in flux-system namespace
◎ waiting for Kustomization reconciliation
✗ Kustomization reconciliation failed: kustomize build failed: exit status 1`

	got, err := filterFluxReconcile(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// On failure, should show the failure line
	if !strings.Contains(got, "✗") {
		t.Errorf("expected ✗ failure line in output, got:\n%s", got)
	}
	if !strings.Contains(got, "failed") {
		t.Errorf("expected failure message in output, got:\n%s", got)
	}
}

func TestFilterFluxSanityCheck(t *testing.T) {
	raw := `NAME            READY   MESSAGE                         REVISION        SUSPENDED
flux-system     True    Applied revision: main/abc1234  main/abc1234    False
myapp           True    Applied revision: main/def5678  main/def5678    False`

	got, err := filterFluxGet(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	if filteredTokens > rawTokens {
		t.Errorf("filter expanded output: raw=%d tokens, filtered=%d tokens", rawTokens, filteredTokens)
	}
}
