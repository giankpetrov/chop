package filters

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRedactHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"authorization header",
			"> Authorization: Bearer secret123",
			"> Authorization: [REDACTED]",
		},
		{
			"cookie header",
			"< Set-Cookie: session=abc; path=/",
			"< Set-Cookie: [REDACTED]",
		},
		{
			"password in text",
			"db_password: my-secret-pass",
			"db_password: [REDACTED]",
		},
		{
			"multiple headers",
			"> Host: example.com\n> api-key: 12345\n> User-Agent: curl",
			"> Host: example.com\n> api-key: [REDACTED]\n> User-Agent: curl",
		},
		{
			"case-insensitive headers",
			"> AUTHORIZATION: Bearer caps-token\n> Authorization: Bearer mixed-token",
			"> AUTHORIZATION: [REDACTED]\n> Authorization: [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactHeaders(tt.input)
			if got != tt.expected {
				t.Errorf("redactHeaders() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestRedactJSON(t *testing.T) {
	rawJSON := `{
		"id": 1,
		"username": "admin",
		"password": "super-secret-password",
		"meta": {
			"session_token": "token-123",
			"public": "visible"
		},
		"tags": ["private", "secret-tag"],
		"env": [
			"PATH=/usr/bin",
			"DB_PASSWORD=sqlpass",
			"DEBUG=true"
		]
	}`

	var parsed interface{}
	json.Unmarshal([]byte(rawJSON), &parsed)

	redacted := redactJSON(parsed)
	redactedJSON, _ := json.Marshal(redacted)
	s := string(redactedJSON)

	if strings.Contains(s, "super-secret-password") {
		t.Error("JSON password not redacted")
	}
	if strings.Contains(s, "token-123") {
		t.Error("JSON session_token not redacted")
	}
	if strings.Contains(s, "sqlpass") {
		t.Error("Env list DB_PASSWORD not redacted")
	}
	if !strings.Contains(s, "visible") {
		t.Error("Non-sensitive data lost during redaction")
	}
	if !strings.Contains(s, "[REDACTED]") {
		t.Error("Redaction marker [REDACTED] missing")
	}
}

func TestRedactJSONRecursiveArray(t *testing.T) {
	rawJSON := `[
		{"id": 1, "token": "t1"},
		{"id": 2, "token": "t2"}
	]`

	var parsed interface{}
	json.Unmarshal([]byte(rawJSON), &parsed)

	redacted := redactJSON(parsed)
	redactedJSON, _ := json.Marshal(redacted)
	s := string(redactedJSON)

	if strings.Contains(s, "t1") || strings.Contains(s, "t2") {
		t.Error("token not redacted in array of objects")
	}
	if !strings.Contains(s, "[REDACTED]") {
		t.Error("Redaction marker missing in array redaction")
	}
}

func TestSmallJSONRedaction(t *testing.T) {
	// Small JSON that would be preserved by isSmallJSON
	raw := `{"status":"ok","api_key":"sk-123"}`

	// Using compressJSON which uses redactJSON internally
	got, err := compressJSON(raw)
	if err != nil {
		t.Fatalf("compressJSON failed: %v", err)
	}

	if strings.Contains(got, "sk-123") {
		t.Errorf("Small JSON secret leaked: %s", got)
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Errorf("Small JSON secret not redacted: %s", got)
	}
}

func TestDockerInspectRedaction(t *testing.T) {
	// Small inspect output that would have been passed through raw
	raw := `[{
		"Id": "123",
		"Config": {
			"Env": ["PASSWORD=secret123", "SAFE=true"]
		}
	}]`

	got, err := filterDockerInspect(raw)
	if err != nil {
		t.Fatalf("filterDockerInspect failed: %v", err)
	}

	if strings.Contains(got, "secret123") {
		t.Error("Docker inspect environment secret leaked in small output")
	}
	if !strings.Contains(got, "PASSWORD=[REDACTED]") {
		t.Error("Docker inspect environment secret not correctly redacted")
	}
}
