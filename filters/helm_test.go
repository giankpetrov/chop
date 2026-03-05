package filters

import (
	"strings"
	"testing"
)

func TestFilterHelmInstall(t *testing.T) {
	raw := `Release "my-release" has been installed. Happy Helming!
NAME: my-release
LAST DEPLOYED: Mon Jan 15 10:30:00 2024
NAMESPACE: default
STATUS: deployed
REVISION: 1
NOTES:
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods)
  echo "Visit http://127.0.0.1:8080"
  kubectl --namespace default port-forward $POD_NAME 8080:80`

	got, err := filterHelmInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "my-release") {
		t.Error("expected release name")
	}
	if !strings.Contains(got, "deployed") {
		t.Error("expected status")
	}
	if strings.Contains(got, "kubectl") {
		t.Error("NOTES should be stripped")
	}
}

func TestFilterHelmList(t *testing.T) {
	raw := "NAME            NAMESPACE       REVISION        UPDATED                                 STATUS          CHART           APP VERSION\n" +
		"my-release      default         3               2024-01-15 10:30:00.000 +0000 UTC       deployed        myapp-1.0.0     1.0.0\n" +
		"other-release   staging         1               2024-01-14 08:00:00.000 +0000 UTC       deployed        otherapp-2.0    2.0.0\n"

	got, err := filterHelmList(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "2 releases") {
		t.Errorf("expected release count, got: %s", got)
	}
}

func TestFilterHelmInstall_Empty(t *testing.T) {
	got, err := filterHelmInstall("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
