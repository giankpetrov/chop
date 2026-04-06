package filters

import (
	"strings"
	"testing"
)

func TestFilterTrivyEmpty(t *testing.T) {
	got, err := filterTrivy("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "No vulnerabilities found" {
		t.Errorf("expected 'No vulnerabilities found', got %q", got)
	}
}

func TestFilterTrivyNoVulnerabilities(t *testing.T) {
	raw := `2024-01-15T10:23:45.678Z	INFO	Vulnerability scanning is enabled
2024-01-15T10:23:45.678Z	INFO	Detected OS: alpine 3.18

myapp:latest (alpine 3.18)
==========================
Total: 0 (UNKNOWN: 0, LOW: 0, MEDIUM: 0, HIGH: 0, CRITICAL: 0)`

	got, err := filterTrivy(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "No vulnerabilities found" {
		t.Errorf("expected 'No vulnerabilities found', got %q", got)
	}
}

func TestFilterTrivyStripsInfoLines(t *testing.T) {
	raw := `2024-01-15T10:23:45.678Z	INFO	Vulnerability scanning is enabled
2024-01-15T10:23:45.678Z	INFO	Detected OS: debian 11.6
2024-01-15T10:23:46.456Z	INFO	Number of language-specific files: 3

myapp:latest (debian 11.6)
===========================
Total: 2 (UNKNOWN: 0, LOW: 0, MEDIUM: 0, HIGH: 1, CRITICAL: 1)

┌──────────────────────┬───────────────┬──────────┬──────────────────┬──────────────┬─────────────────┐
│       Library        │ Vulnerability │ Severity │ Installed Version│ Fixed Version│      Title      │
├──────────────────────┼───────────────┼──────────┼──────────────────┼──────────────┼─────────────────┤
│ curl                 │ CVE-2023-38545│ CRITICAL │ 7.74.0-1.3       │ 7.74.0-1.4   │ curl: overflow  │
│ libc6                │ CVE-2023-4911 │ HIGH     │ 2.31-13          │              │ glibc: overflow │
└──────────────────────┴───────────────┴──────────┴──────────────────┴──────────────┴─────────────────┘`

	got, err := filterTrivy(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Timestamped INFO lines should not appear in output
	if strings.Contains(got, "INFO") {
		t.Errorf("expected INFO log lines to be stripped, got:\n%s", got)
	}
	if strings.Contains(got, "2024-01-15") {
		t.Errorf("expected timestamps to be stripped, got:\n%s", got)
	}
}

func TestFilterTrivyWithVulnerabilities(t *testing.T) {
	raw := `2024-01-15T10:23:45.678Z	INFO	Vulnerability scanning is enabled
2024-01-15T10:23:45.678Z	INFO	Detected OS: debian 11.6
2024-01-15T10:23:46.456Z	INFO	Number of language-specific files: 3

myapp:latest (debian 11.6)
===========================
Total: 47 (UNKNOWN: 0, LOW: 18, MEDIUM: 22, HIGH: 5, CRITICAL: 2)

┌──────────────────────┬───────────────┬──────────┬──────────────────────────────┬──────────────────────────────┬─────────────────────────┐
│       Library        │ Vulnerability │ Severity │      Installed Version       │        Fixed Version         │          Title          │
├──────────────────────┼───────────────┼──────────┼──────────────────────────────┼──────────────────────────────┼─────────────────────────┤
│ curl                 │ CVE-2023-38545│ CRITICAL │ 7.74.0-1.3+deb11u7           │ 7.74.0-1.3+deb11u11          │ curl: SOCKS5 overflow   │
│ libc6                │ CVE-2023-4911 │ HIGH     │ 2.31-13+deb11u7              │                              │ glibc: buffer overflow  │
│ openssl              │ CVE-2023-0465 │ MEDIUM   │ 1.1.1n-0+deb11u4             │ 1.1.1n-0+deb11u5             │ openssl: Invalid cert   │
│ zlib1g               │ CVE-2022-37434│ HIGH     │ 1:1.2.11.dfsg-2+deb11u2      │ 1:1.2.11.dfsg-2+deb11u3      │ zlib: heap buffer       │
│ libssl1.1            │ CVE-2023-0466 │ LOW      │ 1.1.1n-0+deb11u4             │                              │ openssl: cert verify    │
└──────────────────────┴───────────────┴──────────┴──────────────────────────────┴──────────────────────────────┴─────────────────────────┘`

	got, err := filterTrivy(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should include a summary line for the target
	if !strings.Contains(got, "myapp:latest") {
		t.Errorf("expected target name in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Total 47") {
		t.Errorf("expected total count in output, got:\n%s", got)
	}

	// Should list CRITICAL and HIGH individually
	if !strings.Contains(got, "CVE-2023-38545") {
		t.Errorf("expected CRITICAL CVE in output, got:\n%s", got)
	}
	if !strings.Contains(got, "CVE-2023-4911") {
		t.Errorf("expected HIGH CVE in output, got:\n%s", got)
	}

	// MEDIUM and LOW should not be listed individually
	if strings.Contains(got, "CVE-2023-0465") {
		t.Errorf("MEDIUM CVE should not be listed individually, got:\n%s", got)
	}

	// Token savings >= 60%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 60.0 {
		t.Errorf("expected >=60%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterTrivyMultipleTargets(t *testing.T) {
	raw := `2024-01-15T10:23:45.678Z	INFO	Vulnerability scanning is enabled

myapp:latest (debian 11.6)
===========================
Total: 5 (UNKNOWN: 0, LOW: 2, MEDIUM: 1, HIGH: 1, CRITICAL: 1)

┌──────────────────────┬───────────────┬──────────┬──────────────────┬──────────────┬─────────────────┐
│       Library        │ Vulnerability │ Severity │ Installed Version│ Fixed Version│      Title      │
├──────────────────────┼───────────────┼──────────┼──────────────────┼──────────────┼─────────────────┤
│ curl                 │ CVE-2023-38545│ CRITICAL │ 7.74.0-1.3       │ 7.74.0-1.4   │ curl: overflow  │
│ libc6                │ CVE-2023-4911 │ HIGH     │ 2.31-13          │              │ glibc: overflow │
└──────────────────────┴───────────────┴──────────┴──────────────────┴──────────────┴─────────────────┘

Node.js (node-pkg)
==================
Total: 3 (UNKNOWN: 0, LOW: 0, MEDIUM: 2, HIGH: 1, CRITICAL: 0)

┌──────────────────────┬───────────────┬──────────┬──────────────────┬──────────────┬──────────────────┐
│       Library        │ Vulnerability │ Severity │ Installed Version│ Fixed Version│      Title       │
├──────────────────────┼───────────────┼──────────┼──────────────────┼──────────────┼──────────────────┤
│ lodash               │ CVE-2021-23337│ HIGH     │ 4.17.20          │ 4.17.21      │ lodash: injection│
└──────────────────────┴───────────────┴──────────┴──────────────────┴──────────────┴──────────────────┘`

	got, err := filterTrivy(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have summary for both targets
	if !strings.Contains(got, "myapp:latest") {
		t.Errorf("expected first target in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Node.js") {
		t.Errorf("expected second target in output, got:\n%s", got)
	}
}

func TestFilterTrivySanityCheck(t *testing.T) {
	raw := `2024-01-15T10:23:45.678Z	INFO	Vulnerability scanning is enabled
2024-01-15T10:23:45.678Z	INFO	Detected OS: debian 11.6

myapp:latest (debian 11.6)
===========================
Total: 47 (UNKNOWN: 0, LOW: 18, MEDIUM: 22, HIGH: 5, CRITICAL: 2)

┌──────────────────────┬───────────────┬──────────┬──────────────────────────────┬──────────────────────────────┬─────────────────────────┐
│       Library        │ Vulnerability │ Severity │      Installed Version       │        Fixed Version         │          Title          │
├──────────────────────┼───────────────┼──────────┼──────────────────────────────┼──────────────────────────────┼─────────────────────────┤
│ curl                 │ CVE-2023-38545│ CRITICAL │ 7.74.0-1.3+deb11u7           │ 7.74.0-1.3+deb11u11          │ curl: SOCKS5 overflow   │
│ libc6                │ CVE-2023-4911 │ HIGH     │ 2.31-13+deb11u7              │                              │ glibc: buffer overflow  │
└──────────────────────┴───────────────┴──────────┴──────────────────────────────┴──────────────────────────────┴─────────────────────────┘`

	got, err := filterTrivy(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) > len(raw) {
		t.Errorf("filter expanded output: raw=%d bytes, filtered=%d bytes", len(raw), len(got))
	}
}
