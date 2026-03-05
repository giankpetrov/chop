package filters

import (
	"strings"
	"testing"
)

func TestFilterKubectlApply(t *testing.T) {
	raw := "namespace/my-namespace created\n" +
		"configmap/my-config-settings created\n" +
		"configmap/my-config-secrets created\n" +
		"deployment.apps/my-app-frontend configured\n" +
		"deployment.apps/my-app-backend configured\n" +
		"deployment.apps/my-app-worker configured\n" +
		"service/my-svc-frontend unchanged\n" +
		"service/my-svc-backend unchanged\n" +
		"secret/my-secret-tls created\n" +
		"ingress.networking.k8s.io/my-ingress created\n"

	got, err := filterKubectlApply(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression, got %d >= %d", len(got), len(raw))
	}
	if !strings.Contains(got, "created(") {
		t.Errorf("expected grouped output, got: %s", got)
	}
}

func TestFilterKubectlDelete(t *testing.T) {
	raw := "pod \"my-pod-abc123\" deleted\n" +
		"pod \"my-pod-def456\" deleted\n" +
		"pod \"my-pod-ghi789\" deleted\n" +
		"service \"my-svc\" deleted\n" +
		"deployment.apps \"my-app\" deleted\n"

	got, err := filterKubectlDelete(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "deleted 5 resources") {
		t.Errorf("expected delete count, got: %s", got)
	}
}

func TestFilterKubectlApply_Empty(t *testing.T) {
	got, err := filterKubectlApply("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
