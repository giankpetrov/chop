package filters

import (
	"strings"
	"testing"
)

func TestFilterArgoCDAppListEmpty(t *testing.T) {
	got, err := filterArgoCDAppList("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestFilterArgoCDAppListCompressed(t *testing.T) {
	raw := `NAME          CLUSTER                         NAMESPACE   PROJECT  STATUS     HEALTH   SYNCPOLICY  CONDITIONS  REPO                              PATH         TARGET
myapp-prod    https://kubernetes.default.svc  production  default  Synced     Healthy  Auto-Prune  <none>      https://github.com/org/gitops.git  apps/myapp   main
myapp-staging https://kubernetes.default.svc  staging     default  OutOfSync  Healthy  Auto        <none>      https://github.com/org/gitops.git  apps/myapp   develop
myapp-dev     https://kubernetes.default.svc  dev         default  Synced     Healthy  Auto        <none>      https://github.com/org/gitops.git  apps/myapp   dev
myapp-qa      https://kubernetes.default.svc  qa          default  Synced     Degraded Auto        <none>      https://github.com/org/gitops.git  apps/myapp   qa`

	got, err := filterArgoCDAppList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must keep NAME, STATUS, HEALTH columns
	if !strings.Contains(got, "NAME") {
		t.Errorf("expected NAME column in output, got:\n%s", got)
	}
	if !strings.Contains(got, "STATUS") {
		t.Errorf("expected STATUS column in output, got:\n%s", got)
	}
	if !strings.Contains(got, "HEALTH") {
		t.Errorf("expected HEALTH column in output, got:\n%s", got)
	}

	// Must drop verbose columns
	if strings.Contains(got, "https://kubernetes.default.svc") {
		t.Errorf("should drop CLUSTER column, got:\n%s", got)
	}
	if strings.Contains(got, "https://github.com") {
		t.Errorf("should drop REPO column, got:\n%s", got)
	}

	// Token savings >= 40%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 40.0 {
		t.Errorf("expected >=40%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterArgoCDAppSyncSuccess(t *testing.T) {
	raw := `TIMESTAMP                  GROUP        KIND         NAMESPACE    NAME         STATUS     HEALTH   HOOK  MESSAGE
2024-01-15T10:23:45+00:00             Service      production   myapp        Synced     Healthy        service unchanged
2024-01-15T10:23:46+00:00  apps        Deployment   production   myapp        Synced     Healthy        deployment updated

Name:               myapp-prod
Sync Status:        Synced
Health Status:      Healthy`

	got, err := filterArgoCDAppSync(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Sync:") && !strings.Contains(got, "Synced") {
		t.Errorf("expected sync status in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Healthy") {
		t.Errorf("expected health status in output, got:\n%s", got)
	}
	// Should include resource list
	if !strings.Contains(got, "Deployment") && !strings.Contains(got, "Service") {
		t.Errorf("expected resource list in output, got:\n%s", got)
	}
}

func TestFilterArgoCDAppGetHealthy(t *testing.T) {
	raw := `Name:               myapp-prod
Project:            default
Server:             https://kubernetes.default.svc
Namespace:          production
URL:                https://argocd.example.com/applications/myapp-prod
Repo:               https://github.com/org/gitops.git
Target:             main
Path:               apps/myapp
SyncWindow:         Sync Allowed
Sync Policy:        Automated (Prune)
Sync Status:        Synced to main (abc1234)
Health Status:      Healthy`

	got, err := filterArgoCDAppGet(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Name:") {
		t.Errorf("expected Name in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Sync Status:") {
		t.Errorf("expected Sync Status in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Health Status:") {
		t.Errorf("expected Health Status in output, got:\n%s", got)
	}
	// Should drop verbose fields
	if strings.Contains(got, "SyncWindow") {
		t.Errorf("should drop SyncWindow line, got:\n%s", got)
	}
}

func TestFilterArgoCDAppGetDegraded(t *testing.T) {
	raw := `Name:               myapp-prod
Project:            default
Server:             https://kubernetes.default.svc
Namespace:          production
Sync Status:        OutOfSync
Health Status:      Degraded
Conditions:
  - Type: DegradedApp
    Message: Deployment myapp has 0/3 replicas available`

	got, err := filterArgoCDAppGet(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Degraded") {
		t.Errorf("expected Degraded in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Sync Status:") {
		t.Errorf("expected Sync Status in output, got:\n%s", got)
	}
}

func TestFilterArgoCDSanityCheck(t *testing.T) {
	raw := `NAME          CLUSTER                         NAMESPACE   STATUS   HEALTH
myapp-prod    https://kubernetes.default.svc  production  Synced   Healthy
myapp-staging https://kubernetes.default.svc  staging     Synced   Healthy`

	got, err := filterArgoCDAppList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	if filteredTokens > rawTokens {
		t.Errorf("filter expanded output: raw=%d tokens, filtered=%d tokens", rawTokens, filteredTokens)
	}
}
