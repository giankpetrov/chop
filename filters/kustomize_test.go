package filters

import (
	"strings"
	"testing"
)

func TestFilterKustomizeBuildEmpty(t *testing.T) {
	got, err := filterKustomizeBuild("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestFilterKustomizeBuildFewResources(t *testing.T) {
	raw := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: production
spec:
  replicas: 3
---
apiVersion: v1
kind: Service
metadata:
  name: myapp
spec:
  ports:
  - port: 80`

	got, err := filterKustomizeBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2 resources — below threshold of 3, should passthrough
	if got != raw {
		t.Errorf("expected passthrough for ≤3 resources, got:\n%s", got)
	}
}

func TestFilterKustomizeBuildManyResources(t *testing.T) {
	raw := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-worker
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-scheduler
---
apiVersion: v1
kind: Service
metadata:
  name: myapp
---
apiVersion: v1
kind: Service
metadata:
  name: myapp-headless
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
data:
  key: value
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp-ingress
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: myapp-hpa`

	got, err := filterKustomizeBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "resources:") {
		t.Errorf("expected summary line with 'resources:', got:\n%s", got)
	}
	if !strings.Contains(got, "Deployment") {
		t.Errorf("expected Deployment in summary, got:\n%s", got)
	}

	// Token savings >= 60%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 60.0 {
		t.Errorf("expected >=60%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterKustomizeBuildResourceCounting(t *testing.T) {
	raw := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-3
---
apiVersion: v1
kind: Service
metadata:
  name: svc-1
---
apiVersion: v1
kind: Service
metadata:
  name: svc-2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cfg-1`

	got, err := filterKustomizeBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "6 resources") {
		t.Errorf("expected '6 resources' in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Deployment(3)") {
		t.Errorf("expected 'Deployment(3)' in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Service(2)") {
		t.Errorf("expected 'Service(2)' in output, got:\n%s", got)
	}
	if !strings.Contains(got, "ConfigMap(1)") {
		t.Errorf("expected 'ConfigMap(1)' in output, got:\n%s", got)
	}
}

func TestFilterKustomizeSanityCheck(t *testing.T) {
	raw := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
---
apiVersion: v1
kind: Service
metadata:
  name: myapp`

	got, err := filterKustomizeBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	if filteredTokens > rawTokens {
		t.Errorf("filter expanded output: raw=%d tokens, filtered=%d tokens", rawTokens, filteredTokens)
	}
}
