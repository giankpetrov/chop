package filters

import "testing"

func TestFilterDockerInspect_JSON(t *testing.T) {
	raw := `[{"Id":"sha256:abc123def456789","RepoTags":["myapp:latest","myapp:v1.0.0"],"Created":"2024-01-15T10:30:00.123456789Z","Size":150000000,"Config":{"Hostname":"container-hostname-abc","Domainname":"example.com","User":"appuser","ExposedPorts":{"8080/tcp":{},"8443/tcp":{},"9090/tcp":{}},"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","NODE_VERSION=20.11.0","NODE_ENV=production","APP_PORT=8080","LOG_LEVEL=info"],"Cmd":["node","--max-old-space-size=4096","server.js"],"Image":"node:20-alpine","Volumes":{"/data":{},"/logs":{}},"WorkingDir":"/app","Entrypoint":["docker-entrypoint.sh"],"Labels":{"maintainer":"dev@example.com","version":"1.0.0","description":"My application server","org.opencontainers.image.source":"https://github.com/example/myapp"}},"Architecture":"amd64","Os":"linux","RootFS":{"Type":"layers","Layers":["sha256:aaaaaaaaaaaaaaaaaa","sha256:bbbbbbbbbbbbbbbbbb","sha256:cccccccccccccccccc","sha256:dddddddddddddddddd","sha256:eeeeeeeeeeeeeeeeee","sha256:ffffffffffffffffff","sha256:1111111111111111111","sha256:2222222222222222222"]}}]`

	got, err := filterDockerInspect(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression, got %d >= %d", len(got), len(raw))
	}
}

func TestFilterDockerInspect_Short(t *testing.T) {
	raw := `[{"Id":"abc"}]`
	got, err := filterDockerInspect(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != raw {
		t.Errorf("short input should pass through")
	}
}

func TestFilterDockerInspect_Empty(t *testing.T) {
	got, err := filterDockerInspect("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
