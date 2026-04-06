package filters

import (
	"strings"
	"testing"
)

func TestFilterSnykTestEmpty(t *testing.T) {
	got, err := filterSnykTest("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no issues found" {
		t.Errorf("expected 'no issues found', got %q", got)
	}
}

func TestFilterSnykTestNoVulnerabilities(t *testing.T) {
	raw := "Tested 423 dependencies for known issues, found 0 issues."

	got, err := filterSnykTest(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no issues found" {
		t.Errorf("expected 'no issues found', got %q", got)
	}
}

func TestFilterSnykTestWithVulnerabilities(t *testing.T) {
	raw := `Testing /home/user/myapp...

Tested 423 dependencies for known issues, found 12 issues, 8 vulnerable paths.

Issues to fix by upgrading:

  Upgrade express@4.17.1 to express@4.18.2 to fix
  ✗ Cross-site Scripting (XSS) [High Severity][https://snyk.io/vuln/npm:express:20221208] in express@4.17.1
    introduced by express@4.17.1
    introduced by body-parser@1.19.0 > express@4.17.1

  Upgrade lodash@4.17.20 to lodash@4.17.21 to fix
  ✗ Prototype Pollution [High Severity][https://snyk.io/vuln/SNYK-JS-LODASH-567746] in lodash@4.17.20
    introduced by lodash@4.17.20
    introduced by async@3.2.0 > lodash@4.17.20

  Upgrade angular@14.0.0 to angular@14.2.0 to fix
  ✗ Denial of Service (DoS) [Critical Severity][https://snyk.io/vuln/SNYK-JS-ANGULAR-12345] in angular@14.0.0
    introduced by angular@14.0.0
    introduced by @angular/core@14.0.0 > angular@14.0.0

Issues with no direct upgrade or patch:
  ✗ Regular Expression Denial of Service (ReDoS) [Medium Severity][https://snyk.io/vuln/npm:moment:20170905] in moment@2.29.1
    introduced by moment@2.29.1
    introduced by date-utils@1.2.0 > moment@2.29.1
  ✗ Improper Input Validation [Low Severity][https://snyk.io/vuln/npm:qs:20170213] in qs@6.7.3
    introduced by qs@6.7.3
    introduced by express@4.17.1 > qs@6.7.3
  ✗ Information Exposure [Medium Severity][https://snyk.io/vuln/SNYK-JS-FOO-99999] in foo@1.0.0
    introduced by foo@1.0.0
    introduced by bar@2.0.0 > foo@1.0.0

Tested 423 dependencies for known issues, found 12 issues, 8 vulnerable paths.`

	got, err := filterSnykTest(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should group by severity
	if !strings.Contains(got, "Critical") {
		t.Errorf("expected Critical severity group in output, got:\n%s", got)
	}
	if !strings.Contains(got, "High") {
		t.Errorf("expected High severity group in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Medium") {
		t.Errorf("expected Medium severity group in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Low") {
		t.Errorf("expected Low severity group in output, got:\n%s", got)
	}

	// Should include summary line
	if !strings.Contains(got, "issues found") {
		t.Errorf("expected summary line in output, got:\n%s", got)
	}

	// Token savings >= 60%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 60.0 {
		t.Errorf("expected >=60%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterSnykTestURLsStripped(t *testing.T) {
	raw := `Testing /home/user/myapp...

Tested 10 dependencies for known issues, found 2 issues, 2 vulnerable paths.

  ✗ Cross-site Scripting (XSS) [High Severity][https://snyk.io/vuln/npm:express:20221208] in express@4.17.1
    introduced by express@4.17.1

  ✗ Prototype Pollution [Medium Severity][https://snyk.io/vuln/SNYK-JS-LODASH-567746] in lodash@4.17.20
    introduced by lodash@4.17.20`

	got, err := filterSnykTest(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(got, "https://snyk.io") {
		t.Errorf("expected URLs to be stripped from output, got:\n%s", got)
	}
}

func TestFilterSnykCodeTestWithIssues(t *testing.T) {
	raw := `Testing /home/user/myapp ...

 ✗ [High] SQL Injection
     Path: src/db.js, line 45
     Info: Unsanitized input flows into query

 ✗ [Medium] Cross-site Scripting
     Path: src/routes/user.js, line 12
     Info: Unsanitized input from HTTP request flows into response

 ✗ [Critical] Remote Code Execution
     Path: src/exec.js, line 88
     Info: User-controlled data used in exec call

✔ Test completed.

Test Summary
 Organization:      my-org
 ✗ Files with issues: 3
 Total issues: 3 [ 1 critical, 1 high, 1 medium, 0 low ]`

	got, err := filterSnykCodeTest(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show severity groups
	if !strings.Contains(got, "Critical") {
		t.Errorf("expected Critical group in output, got:\n%s", got)
	}
	if !strings.Contains(got, "High") {
		t.Errorf("expected High group in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Medium") {
		t.Errorf("expected Medium group in output, got:\n%s", got)
	}

	// Should include file paths
	if !strings.Contains(got, "src/db.js") {
		t.Errorf("expected file path src/db.js in output, got:\n%s", got)
	}
	if !strings.Contains(got, "src/routes/user.js") {
		t.Errorf("expected file path src/routes/user.js in output, got:\n%s", got)
	}
}

func TestFilterSnykSanityCheck(t *testing.T) {
	raw := `Testing /home/user/myapp...

Tested 423 dependencies for known issues, found 12 issues, 8 vulnerable paths.

  ✗ Cross-site Scripting (XSS) [High Severity][https://snyk.io/vuln/npm:express:20221208] in express@4.17.1
    introduced by express@4.17.1

  ✗ Prototype Pollution [High Severity][https://snyk.io/vuln/SNYK-JS-LODASH-567746] in lodash@4.17.20
    introduced by lodash@4.17.20

  ✗ ReDoS [Medium Severity][https://snyk.io/vuln/npm:moment:20170905] in moment@2.29.1
    introduced by moment@2.29.1

Tested 423 dependencies for known issues, found 12 issues, 8 vulnerable paths.`

	got, err := filterSnykTest(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) > len(raw) {
		t.Errorf("filter expanded output: raw=%d bytes, filtered=%d bytes", len(raw), len(got))
	}
}
