package filters

import (
	"strings"
	"testing"
)

func TestNpmInstallBasic(t *testing.T) {
	raw := `npm warn deprecated inflight@1.0.6: This module is not supported, and leaks memory.
npm warn deprecated glob@7.2.3: Glob versions prior to v9 are no longer supported

added 150 packages, and audited 151 packages in 12s

23 packages are looking for funding
  run ` + "`npm fund`" + ` for details

found 0 vulnerabilities
`
	got, err := filterNpmInstall(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "added 150 packages") {
		t.Errorf("expected added count, got: %s", got)
	}
	if !strings.Contains(got, "warnings") {
		t.Errorf("expected warning count, got: %s", got)
	}
	if !strings.Contains(got, "deprecated") {
		t.Errorf("expected 'deprecated' in warning summary, got: %s", got)
	}
	// Funding notice should be stripped
	if strings.Contains(got, "funding") {
		t.Errorf("expected funding stripped, got: %s", got)
	}

	rawTokens := countTokens(raw)
	filteredTokens := countTokens(got)
	savings := 100.0 - (float64(filteredTokens) / float64(rawTokens) * 100.0)
	if savings < 70.0 {
		t.Errorf("expected >=70%% savings, got %.1f%%", savings)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output: %s", got)
}

func TestNpmInstallWithVulnerabilities(t *testing.T) {
	raw := `npm warn deprecated stable@0.1.8: Modern JS already guarantees Array#sort() is a stable sort
npm warn deprecated svgo@1.3.2: This SVGO version is no longer supported. Upgrade to v2.x.x.

added 1245 packages, and audited 1246 packages in 45s

198 packages are looking for funding
  run ` + "`npm fund`" + ` for details

8 vulnerabilities (2 moderate, 6 high)

To address all issues, run:
  npm audit fix
`
	got, err := filterNpmInstall(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "added 1245 packages") {
		t.Errorf("expected added count, got: %s", got)
	}
	if !strings.Contains(got, "vulnerabilities") {
		t.Errorf("expected vulnerabilities preserved, got: %s", got)
	}
	if strings.Contains(got, "npm audit fix") {
		t.Errorf("expected audit fix suggestion stripped, got: %s", got)
	}

	rawTokens := countTokens(raw)
	filteredTokens := countTokens(got)
	savings := 100.0 - (float64(filteredTokens) / float64(rawTokens) * 100.0)
	if savings < 70.0 {
		t.Errorf("expected >=70%% savings, got %.1f%%", savings)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output: %s", got)
}

func TestNpmInstallErrors(t *testing.T) {
	raw := `npm error code ERESOLVE
npm error ERESOLVE unable to resolve dependency tree
npm error
npm error While resolving: my-app@1.0.0
npm error Found: react@18.2.0
npm error   node_modules/react
npm error     react@"^18.2.0" from the root project
npm error
npm error Could not resolve dependency:
npm error peer react@"^17.0.0" from some-old-lib@2.1.0
`
	got, err := filterNpmInstall(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "npm error") {
		t.Errorf("expected errors preserved, got: %s", got)
	}
	t.Logf("output: %s", got)
}
