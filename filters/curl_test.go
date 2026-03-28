package filters

import (
	"strings"
	"testing"
)

func TestFilterCurlJSONWithHeaders(t *testing.T) {
	raw := "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nX-Request-Id: abc-123\r\nDate: Thu, 01 Jan 2026 00:00:00 GMT\r\n\r\n" +
		`{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"},{"id":3,"name":"Carol"},{"id":4,"name":"Dave"},{"id":5,"name":"Eve"},{"id":6,"name":"Frank"}],"total":150,"page":1,"per_page":6,"has_more":true,"query":"active"}`

	got, err := filterCurl(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should keep status line
	if !strings.HasPrefix(got, "HTTP/1.1 200 OK") {
		t.Errorf("expected status line preserved, got:\n%s", got)
	}

	// Should NOT contain other headers
	if strings.Contains(got, "X-Request-Id") {
		t.Error("expected other headers stripped")
	}
	if strings.Contains(got, "Date:") {
		t.Error("expected Date header stripped")
	}

	// Should contain compressed JSON
	if !strings.Contains(got, "users") {
		t.Errorf("expected JSON key 'users' in output, got:\n%s", got)
	}
}

func TestFilterCurlHTMLResponse(t *testing.T) {
	raw := `<!DOCTYPE html>
<html>
<head><title>Not Found</title></head>
<body><h1>404 Not Found</h1><p>The page you requested was not found.</p></body>
</html>`

	got, err := filterCurl(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should extract text content from HTML, not raw tags
	if strings.Contains(got, "<html>") || strings.Contains(got, "<body>") {
		t.Errorf("expected HTML tags stripped, got:\n%s", got)
	}
	// Should contain meaningful text
	if !strings.Contains(got, "Not Found") {
		t.Errorf("expected text content preserved, got:\n%s", got)
	}
}

func TestFilterCurlError(t *testing.T) {
	raw := `curl: (7) Failed to connect to localhost port 8080: Connection refused`

	got, err := filterCurl(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Errors should be preserved in full
	if got != raw {
		t.Errorf("expected curl error preserved, got:\n%s", got)
	}
}

func TestFilterCurlTimeout(t *testing.T) {
	raw := `curl: (28) Operation timed out after 30000 milliseconds with 0 bytes received`

	got, err := filterCurl(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != raw {
		t.Errorf("expected timeout error preserved, got:\n%s", got)
	}
}

func TestFilterCurlPlainText(t *testing.T) {
	// Short text: pass through
	raw := "Hello, World!\nThis is plain text."
	got, err := filterCurl(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != raw {
		t.Errorf("expected plain text passthrough, got:\n%s", got)
	}
}

func TestFilterCurlPlainTextLong(t *testing.T) {
	// Build >100 lines of plain text
	var lines []string
	for i := 0; i < 150; i++ {
		lines = append(lines, "line of plain text output from server")
	}
	raw := strings.Join(lines, "\n")

	got, err := filterCurl(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "... (50 more lines)") {
		t.Errorf("expected truncation message, got last 100 chars:\n%s", got[len(got)-100:])
	}

	gotLines := strings.Split(got, "\n")
	// 100 lines + 1 truncation message
	if len(gotLines) != 101 {
		t.Errorf("expected 101 lines (100 + truncation), got %d", len(gotLines))
	}
}

func TestFilterCurlEmpty(t *testing.T) {
	got, err := filterCurl("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFilterCurlLargeJSONSavings(t *testing.T) {
	// Build a large JSON response
	var items []string
	for i := 0; i < 100; i++ {
		items = append(items, `{"id":`+strings.Repeat("1", 1)+`,"name":"User Name Here","email":"user@example.com","created":"2024-01-15T10:30:00Z","active":true}`)
	}
	raw := `[` + strings.Join(items, ",") + `]`

	got, err := filterCurl(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rawTokens := len(strings.Fields(raw))
	gotTokens := len(strings.Fields(got))
	savings := 100.0 - float64(gotTokens)/float64(rawTokens)*100.0
	if savings < 70.0 {
		t.Errorf("expected >=70%% token savings, got %.1f%% (raw=%d, filtered=%d)\noutput:\n%s", savings, rawTokens, gotTokens, got)
	}
}

func TestFilterCurlHeadersWithLFOnly(t *testing.T) {
	// Some servers send LF-only line endings
	raw := "HTTP/1.1 404 Not Found\nContent-Type: text/plain\n\nResource not found"

	got, err := filterCurl(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "HTTP/1.1 404 Not Found") {
		t.Errorf("expected status line, got:\n%s", got)
	}
	if !strings.Contains(got, "Resource not found") {
		t.Errorf("expected body preserved, got:\n%s", got)
	}
}

func TestFilterCurlRedactsSensitiveHeaders(t *testing.T) {
	raw := "> GET /api/v1/user HTTP/1.1\n" +
		"> Host: api.example.com\n" +
		"> Authorization: Bearer secret-token-123\n" +
		"> Cookie: session_id=abc123xyz789\n" +
		">\n" +
		"HTTP/1.1 200 OK\n" +
		"Content-Type: application/json\n" +
		"\n" +
		`{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"},{"id":3,"name":"Carol"},{"id":4,"name":"Dave"},{"id":5,"name":"Eve"},{"id":6,"name":"Frank"}],"total":150}`

	got, err := filterCurl(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Currently, this should FAIL once we implement redaction,
	// but for now, it should pass as "Authorization" is not redacted.
	// Actually, I'll write it as it SHOULD be, so it fails now.
	if strings.Contains(got, "secret-token-123") {
		t.Errorf("found sensitive token in output:\n%s", got)
	}
	if strings.Contains(got, "abc123xyz789") {
		t.Errorf("found sensitive cookie in output:\n%s", got)
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output:\n%s", got)
	}
}
