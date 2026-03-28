package filters

import (
	"strings"
	"testing"
)

var dockerPullFixture = `Using default tag: latest
latest: Pulling from library/nginx
a2abf6c4d29d: Pull complete
a9edb18cadd1: Pull complete
589b7251471a: Pull complete
186b1aaa4aa6: Pull complete
b4df32aa5a72: Pull complete
a0bcbecc962e: Pull complete
Digest: sha256:0d17b565c37bcbd895e9d92315a3c2d87d2c8129c35c6f9ce0b3e12b8c7b4d1
Status: Downloaded newer image for nginx:latest
docker.io/library/nginx:latest
`

func TestFilterDockerPull(t *testing.T) {
	got, err := filterDockerPull(dockerPullFixture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should keep Status line
	if !strings.Contains(got, "Status: Downloaded") {
		t.Errorf("expected 'Status: Downloaded' preserved, got:\n%s", got)
	}

	// Should keep Digest line
	if !strings.Contains(got, "Digest: sha256:") {
		t.Errorf("expected 'Digest:' line preserved, got:\n%s", got)
	}

	// Should keep final image reference
	if !strings.Contains(got, "docker.io/library/nginx:latest") {
		t.Errorf("expected final image reference preserved, got:\n%s", got)
	}

	// Should drop layer lines
	if strings.Contains(got, "Pull complete") {
		t.Errorf("expected layer 'Pull complete' lines dropped, got:\n%s", got)
	}

	// Should drop "Using default tag:"
	if strings.Contains(got, "Using default tag:") {
		t.Errorf("expected 'Using default tag:' dropped, got:\n%s", got)
	}

	// Should drop "Pulling from"
	if strings.Contains(got, "Pulling from") {
		t.Errorf("expected 'Pulling from' dropped, got:\n%s", got)
	}
}

func TestFilterDockerPullUpToDate(t *testing.T) {
	raw := `Status: Image is up to date for nginx:latest
docker.io/library/nginx:latest
`
	got, err := filterDockerPull(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "Status: Image is up to date") {
		t.Errorf("expected status line preserved, got:\n%s", got)
	}
}

func TestFilterDockerPullEmpty(t *testing.T) {
	got, err := filterDockerPull("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestDockerPullRouted(t *testing.T) {
	if get("docker", []string{"pull"}) == nil {
		t.Error("expected non-nil filter for docker pull")
	}
}
