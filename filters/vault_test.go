package filters

import (
	"fmt"
	"strings"
	"testing"
)

func TestFilterVaultReadEmpty(t *testing.T) {
	got, err := filterVaultRead("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty passthrough, got %q", got)
	}
}

func TestFilterVaultReadRedactsSecrets(t *testing.T) {
	raw := `Key                 Value
---                 -----
lease_id            database/creds/my-role/abc123
lease_duration      1h
lease_renewable     true
password            s3cr3t-passw0rd
token               hvs.CAESABCDEF123456
secret_key          my-super-secret-key
username            v-root-my-role-xyz789`

	got, err := filterVaultRead(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Secret keys must be redacted
	if strings.Contains(got, "s3cr3t-passw0rd") {
		t.Errorf("password value should be redacted")
	}
	if strings.Contains(got, "hvs.CAESABCDEF123456") {
		t.Errorf("token value should be redacted")
	}
	if strings.Contains(got, "my-super-secret-key") {
		t.Errorf("secret_key value should be redacted")
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Errorf("expected [REDACTED] marker in output, got:\n%s", got)
	}

	// Non-secret keys keep their values
	if !strings.Contains(got, "database/creds/my-role/abc123") {
		t.Errorf("lease_id value should not be redacted")
	}
	if !strings.Contains(got, "v-root-my-role-xyz789") {
		t.Errorf("username value should not be redacted")
	}
}

func TestFilterVaultReadKeepsNonSecretKeys(t *testing.T) {
	raw := `Key                 Value
---                 -----
lease_id            database/creds/my-role/abc123
lease_duration      1h
lease_renewable     true
ttl                 768h
username            v-root-my-role-xyz789`

	got, err := filterVaultRead(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// None of these should be redacted
	if strings.Contains(got, "[REDACTED]") {
		t.Errorf("expected no redaction for non-secret keys, got:\n%s", got)
	}

	if !strings.Contains(got, "database/creds/my-role/abc123") {
		t.Errorf("lease_id value should be present")
	}
	if !strings.Contains(got, "v-root-my-role-xyz789") {
		t.Errorf("username value should be present")
	}
	if !strings.Contains(got, "768h") {
		t.Errorf("ttl value should be present")
	}
}

func TestFilterVaultKvGet(t *testing.T) {
	raw := `======= Secret Path =======
secret/data/myapp/config

======= Metadata =======
Key              Value
---              -----
created_time     2024-01-15T10:23:45.678Z
version          3

====== Data ======
Key          Value
---          -----
api_key      xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
db_password  s3cr3t-passw0rd
debug        false`

	got, err := filterVaultRead(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Section label should be present in compressed form
	if !strings.Contains(got, "[Data]") {
		t.Errorf("expected [Data] section label in output, got:\n%s", got)
	}

	// Secret values must be redacted
	if strings.Contains(got, "s3cr3t-passw0rd") {
		t.Errorf("db_password value should be redacted")
	}
	if strings.Contains(got, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx") {
		t.Errorf("api_key value should be redacted")
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output, got:\n%s", got)
	}

	// Non-secret values should stay
	if !strings.Contains(got, "false") {
		t.Errorf("debug value 'false' should be present")
	}
}

func TestFilterVaultListShort(t *testing.T) {
	raw := `Keys
----
config
credentials
database
tokens
users`

	got, err := filterVaultList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// <=10 items — passthrough: all items shown
	if !strings.Contains(got, "config") {
		t.Errorf("expected 'config' in output")
	}
	if !strings.Contains(got, "credentials") {
		t.Errorf("expected 'credentials' in output")
	}
	if !strings.Contains(got, "database") {
		t.Errorf("expected 'database' in output")
	}
}

func TestFilterVaultListLong(t *testing.T) {
	var items []string
	for i := 1; i <= 25; i++ {
		items = append(items, fmt.Sprintf("secret-%02d", i))
	}
	raw := "Keys\n----\n" + strings.Join(items, "\n")

	got, err := filterVaultList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should summarize: count, first few items, and "and N more"
	if !strings.Contains(got, "25") {
		t.Errorf("expected item count 25 in output, got:\n%s", got)
	}
	if !strings.Contains(got, "more") {
		t.Errorf("expected 'more' in truncated output, got:\n%s", got)
	}
	if !strings.Contains(got, "secret-01") {
		t.Errorf("expected first item 'secret-01' in output, got:\n%s", got)
	}
}

func TestFilterVaultMountList(t *testing.T) {
	raw := `Path          Type         Accessor              Description
----          ----         --------              -----------
cubbyhole/    cubbyhole    cubbyhole_abc123      per-token private secret storage
identity/     identity     identity_def456       identity store
secret/       kv           kv_ghi789             key/value secret storage (v2)
sys/          system       system_jkl012         system endpoints used for control`

	got, err := filterVaultMountList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should keep Path and Type columns
	if !strings.Contains(got, "cubbyhole/") {
		t.Errorf("expected path 'cubbyhole/' in output")
	}
	if !strings.Contains(got, "kv") {
		t.Errorf("expected type 'kv' in output")
	}

	// Should drop Accessor and Description
	if strings.Contains(got, "cubbyhole_abc123") {
		t.Errorf("accessor should be dropped from output")
	}
	if strings.Contains(got, "per-token private secret storage") {
		t.Errorf("description should be dropped from output")
	}
}

func TestFilterVaultSanityCheck(t *testing.T) {
	// Generate a large vault read output
	var lines []string
	lines = append(lines, "Key                 Value")
	lines = append(lines, "---                 -----")
	for i := 1; i <= 30; i++ {
		lines = append(lines, fmt.Sprintf("key_%02d              value_%02d_data_here", i, i))
	}
	raw := strings.Join(lines, "\n")

	got, err := filterVaultRead(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) > len(raw) {
		t.Errorf("output longer than input: raw=%d bytes, filtered=%d bytes", len(raw), len(got))
	}
}

func TestLooksLikeVaultOutputNegative(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{
			name:  "just equals signs",
			input: "====",
		},
		{
			name:  "keys and dashes without vault header",
			input: "Keys\n----\nconfig\ncredentials",
		},
		{
			name:  "markdown underline header",
			input: "some markdown\n====\nheader",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if looksLikeVaultOutput(tc.input) {
				t.Errorf("looksLikeVaultOutput(%q) = true, want false", tc.input)
			}
		})
	}
}
