package filters

import (
	"strings"
	"testing"
)

var kubectlRolloutFixture = `Waiting for deployment "myapp" rollout to finish: 0 of 3 updated replicas are available...
Waiting for deployment "myapp" rollout to finish: 1 of 3 updated replicas are available...
Waiting for deployment "myapp" rollout to finish: 2 of 3 updated replicas are available...
deployment "myapp" successfully rolled out
`

func TestFilterKubectlRolloutStatus(t *testing.T) {
	got, err := filterKubectlRolloutStatus(kubectlRolloutFixture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should keep "successfully rolled out"
	if !strings.Contains(got, "successfully rolled out") {
		t.Errorf("expected 'successfully rolled out' preserved, got:\n%s", got)
	}

	// Should keep "Waiting for" lines
	if !strings.Contains(got, "Waiting for") {
		t.Errorf("expected 'Waiting for' lines preserved, got:\n%s", got)
	}
}

func TestFilterKubectlRolloutStatusEmpty(t *testing.T) {
	got, err := filterKubectlRolloutStatus("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestKubectlRolloutRouted(t *testing.T) {
	if get("kubectl", []string{"rollout", "status"}) == nil {
		t.Error("expected non-nil filter for kubectl rollout status")
	}
	// "rollout history" should return filterAutoDetect (non-nil)
	if get("kubectl", []string{"rollout", "history"}) == nil {
		t.Error("expected non-nil filter for kubectl rollout history")
	}
}
