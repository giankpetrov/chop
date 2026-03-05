package filters

import (
	"strings"
	"testing"
)

func TestFilterEslintProblems(t *testing.T) {
	raw := "src/components/App.tsx\n" +
		"   5:10  error    Unexpected var, use let or const instead  no-var\n" +
		"  12:3   error    'React' is defined but never used         no-unused-vars\n" +
		"  18:7   warning  Missing semicolon                         semi\n" +
		"  24:5   error    Unexpected var, use let or const instead  no-var\n" +
		"  30:8   error    Unexpected var, use let or const instead  no-var\n" +
		"\n" +
		"src/utils/helpers.ts\n" +
		"   3:1   error    Unexpected var, use let or const instead  no-var\n" +
		"   8:12  warning  Missing semicolon                         semi\n" +
		"  15:5   error    'lodash' is defined but never used        no-unused-vars\n" +
		"  22:8   warning  Unexpected console statement              no-console\n" +
		"  28:3   error    Unexpected var, use let or const instead  no-var\n" +
		"\n" +
		"src/services/api.ts\n" +
		"   4:10  error    'axios' is defined but never used         no-unused-vars\n" +
		"  11:3   warning  Missing semicolon                         semi\n" +
		"  19:7   error    Unexpected var, use let or const instead  no-var\n" +
		"  25:14  warning  Unexpected console statement              no-console\n" +
		"  33:5   error    Expected '===' and instead saw '=='       eqeqeq\n" +
		"  40:9   warning  Missing semicolon                         semi\n" +
		"\n" +
		"src/models/User.ts\n" +
		"   7:3   warning  Missing semicolon                         semi\n" +
		"  14:8   error    Unexpected var, use let or const instead  no-var\n" +
		"  21:5   warning  Unexpected console statement              no-console\n" +
		"\n" +
		"src/routes/index.tsx\n" +
		"   9:6   error    'useState' is defined but never used      no-unused-vars\n" +
		"  16:3   warning  Missing semicolon                         semi\n" +
		"\n" +
		"22 problems (12 errors, 10 warnings)\n" +
		"8 errors and 4 warnings potentially fixable with the `--fix` option.\n"

	got, err := filterEslint(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Grouped by rule
	if !strings.Contains(got, "no-var (7)") {
		t.Errorf("expected 'no-var (7)', got:\n%s", got)
	}
	if !strings.Contains(got, "no-unused-vars (4)") {
		t.Errorf("expected 'no-unused-vars (4)', got:\n%s", got)
	}
	if !strings.Contains(got, "semi (6)") {
		t.Errorf("expected 'semi (6)', got:\n%s", got)
	}
	if !strings.Contains(got, "no-console (3)") {
		t.Errorf("expected 'no-console (3)', got:\n%s", got)
	}
	if !strings.Contains(got, "eqeqeq (1)") {
		t.Errorf("expected 'eqeqeq (1)', got:\n%s", got)
	}

	// Summary
	if !strings.Contains(got, "21 problems (12 errors, 9 warnings)") {
		t.Errorf("expected summary line, got:\n%s", got)
	}

	// Fixable preserved
	if !strings.Contains(got, "fixable") {
		t.Errorf("expected fixable message preserved, got:\n%s", got)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 70.0 {
		t.Errorf("expected >=70%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestFilterEslintClean(t *testing.T) {
	got, err := filterEslint("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no problems" {
		t.Errorf("expected 'no problems', got %q", got)
	}
}
