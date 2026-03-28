package filters

import (
	"strings"
	"testing"
)

var sshBannerFixture = `Warning: Permanently added '192.168.1.100' (ECDSA) to the list of known hosts.
Welcome to Ubuntu 22.04.3 LTS (GNU/Linux 5.15.0-91-generic x86_64)
Last login: Mon Mar 15 09:00:00 2024 from 192.168.1.1
System information as of Mon Mar 15 09:30:00 UTC 2024
---------------------------------------------------------------------------
ls /var/log
auth.log
syslog
`

func TestFilterSsh(t *testing.T) {
	got, err := filterSsh(sshBannerFixture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should drop "Warning: Permanently added"
	if strings.Contains(got, "Warning: Permanently added") {
		t.Errorf("expected 'Warning: Permanently added' dropped, got:\n%s", got)
	}

	// Should drop "Welcome to"
	if strings.Contains(got, "Welcome to") {
		t.Errorf("expected 'Welcome to' dropped, got:\n%s", got)
	}

	// Should drop "Last login:"
	if strings.Contains(got, "Last login:") {
		t.Errorf("expected 'Last login:' dropped, got:\n%s", got)
	}

	// Should keep actual command output
	if !strings.Contains(got, "auth.log") {
		t.Errorf("expected command output 'auth.log' preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "syslog") {
		t.Errorf("expected command output 'syslog' preserved, got:\n%s", got)
	}
}

func TestFilterSshEmpty(t *testing.T) {
	got, err := filterSsh("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFilterSshPureCommandOutput(t *testing.T) {
	// Output with no banner noise - should pass through
	raw := "total 0\ndrwxr-xr-x 1 root root 4096 Mar 15 10:00 .\ndrwxr-xr-x 1 root root 4096 Mar 15 10:00 ..\n"
	got, err := filterSsh(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No banner noise - should return raw (looksLikeSshOutput will be false)
	if got != raw {
		t.Errorf("expected pure command output to pass through, got:\n%s", got)
	}
}

func TestSshRouted(t *testing.T) {
	if get("ssh", []string{}) == nil {
		t.Error("expected non-nil filter for ssh")
	}
}
