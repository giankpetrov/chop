package filters

import (
	"strings"
	"testing"
)

func TestFilterDockerPs(t *testing.T) {
	raw := `CONTAINER ID   IMAGE                    COMMAND                  CREATED          STATUS          PORTS                                       NAMES
a1b2c3d4e5f6   nginx:latest             "/docker-entrypoint.…"   2 hours ago      Up 2 hours      0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp   web-proxy
b2c3d4e5f6a7   postgres:15              "docker-entrypoint.s…"   3 hours ago      Up 3 hours      0.0.0.0:5432->5432/tcp                      app-database
c3d4e5f6a7b8   redis:7-alpine           "docker-entrypoint.s…"   3 hours ago      Up 3 hours      0.0.0.0:6379->6379/tcp                      cache-redis
d4e5f6a7b8c9   node:20-slim             "docker-entrypoint.s…"   45 minutes ago   Up 45 minutes   0.0.0.0:3000->3000/tcp                      api-server
e5f6a7b8c9d0   grafana/grafana:latest   "/run.sh"                5 hours ago      Up 5 hours      0.0.0.0:3001->3000/tcp                      monitoring
f6a7b8c9d0e1   prom/prometheus:latest    "/bin/prometheus --c…"   5 hours ago      Up 5 hours      0.0.0.0:9090->9090/tcp                      prometheus`

	got, err := filterDockerPs(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 6 lines, one per container
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 6 {
		t.Errorf("expected 6 lines, got %d:\n%s", len(lines), got)
	}

	// Each line should have the container name
	for _, name := range []string{"web-proxy", "app-database", "cache-redis", "api-server", "monitoring", "prometheus"} {
		if !strings.Contains(got, name) {
			t.Errorf("expected output to contain %q", name)
		}
	}

	// Each line should have the image in parens
	if !strings.Contains(got, "(nginx:latest)") {
		t.Errorf("expected output to contain image in parens, got:\n%s", got)
	}

	// Token savings >= 60%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 60.0 {
		t.Errorf("expected >=60%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterDockerPsEmpty(t *testing.T) {
	got, err := filterDockerPs("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no running containers" {
		t.Errorf("expected 'no running containers', got %q", got)
	}
}

func TestFilterDockerPsHeaderOnly(t *testing.T) {
	raw := `CONTAINER ID   IMAGE   COMMAND   CREATED   STATUS   PORTS   NAMES`
	got, err := filterDockerPs(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no running containers" {
		t.Errorf("expected 'no running containers', got %q", got)
	}
}

func TestPodmanRoutesToDockerFilter(t *testing.T) {
	for _, sub := range []string{"ps", "build", "images", "logs"} {
		f := get("podman", []string{sub})
		if f == nil {
			t.Errorf("expected filter for podman %s, got nil", sub)
		}
	}
}
