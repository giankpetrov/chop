package filters

import (
	"strings"
	"testing"
)

func TestFilterHttpieJSONResponse(t *testing.T) {
	// httpie typically shows headers + body
	raw := "HTTP/1.1 200 OK\nContent-Type: application/json\nServer: nginx\n\n" +
		`{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"},{"id":3,"name":"Carol"},{"id":4,"name":"Dave"}],"total":100,"page":1}`

	got, err := filterHttpie(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should keep status line
	if !strings.HasPrefix(got, "HTTP/1.1 200 OK") {
		t.Errorf("expected status line, got:\n%s", got)
	}

	// Should strip other headers
	if strings.Contains(got, "Server: nginx") {
		t.Error("expected headers stripped")
	}

	// Should contain JSON structure
	if !strings.Contains(got, "users") {
		t.Errorf("expected 'users' key in output, got:\n%s", got)
	}
}

func TestFilterHttpieError(t *testing.T) {
	raw := `http: error: ConnectionError: HTTPConnectionPool(host='localhost', port=8080): Max retries exceeded with url: /api/users (Caused by NewConnectionError('<urllib3.connection.HTTPConnection object>: Failed to establish a new connection: [Errno 111] Connection refused'))`

	got, err := filterHttpie(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Errors preserved
	if got != raw {
		t.Errorf("expected error preserved, got:\n%s", got)
	}
}

func TestFilterHttpieRedactsSensitiveHeaders(t *testing.T) {
	raw := "HTTP/1.1 200 OK\n" +
		"Authorization: Basic dXNlcjpwYXNzd29yZA==\n" +
		"X-Api-Key: top-secret-key\n" +
		"\n" +
		`{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"},{"id":3,"name":"Carol"},{"id":4,"name":"Dave"},{"id":5,"name":"Eve"},{"id":6,"name":"Frank"}],"total":150}`

	got, err := filterHttpie(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(got, "dXNlcjpwYXNzd29yZA==") {
		t.Errorf("found sensitive token in output:\n%s", got)
	}
	if strings.Contains(got, "top-secret-key") {
		t.Errorf("found sensitive api key in output:\n%s", got)
	}
}

func TestFilterHttpieHTMLResponse(t *testing.T) {
	raw := "HTTP/1.1 404 Not Found\nContent-Type: text/html\n\n<!DOCTYPE html><html><body><h1>Not Found</h1></body></html>"

	got, err := filterHttpie(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "HTTP/1.1 404 Not Found") {
		t.Errorf("expected status line, got:\n%s", got)
	}
	if !strings.Contains(got, "HTML response (") {
		t.Errorf("expected HTML summary, got:\n%s", got)
	}
}

func TestFilterHttpieEmpty(t *testing.T) {
	got, err := filterHttpie("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestFilterHttpieBodyOnly(t *testing.T) {
	// httpie without headers (--body flag)
	raw := `{"status": "ok", "message": "created"}`

	got, err := filterHttpie(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Small JSON should preserve values
	if !strings.Contains(got, "ok") {
		t.Errorf("expected value preserved, got:\n%s", got)
	}
}
