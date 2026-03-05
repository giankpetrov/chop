package filters

import (
	"strings"
	"testing"
)

func TestFilterGoBuildClean(t *testing.T) {
	got, err := filterGoBuild("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "build ok" {
		t.Errorf("expected 'build ok', got %q", got)
	}
}

func TestFilterGoBuildErrors(t *testing.T) {
	// Real go build output includes package headers and sometimes duplicate
	// error references. The filter strips headers and deduplicates.
	raw := "# github.com/example/myapp/cmd/server\n" +
		"cmd/server/main.go:12:5: undefined: Config\n" +
		"cmd/server/main.go:12:5: undefined: Config\n" +
		"cmd/server/main.go:12:5: undefined: Config\n" +
		"cmd/server/main.go:15:10: cannot use x (variable of type string) as type int in argument to fmt.Println\n" +
		"cmd/server/main.go:15:10: cannot use x (variable of type string) as type int in argument to fmt.Println\n" +
		"# github.com/example/myapp/internal/handlers\n" +
		"internal/handlers/user.go:23:8: imported and not used: \"fmt\"\n" +
		"internal/handlers/user.go:23:8: imported and not used: \"fmt\"\n" +
		"internal/handlers/user.go:23:8: imported and not used: \"fmt\"\n" +
		"internal/handlers/user.go:45:3: undefined: handleRequest\n" +
		"internal/handlers/user.go:45:3: undefined: handleRequest\n" +
		"# github.com/example/myapp/internal/services\n" +
		"internal/services/auth.go:67:12: not enough arguments in call to db.Query\n" +
		"internal/services/auth.go:67:12: not enough arguments in call to db.Query\n" +
		"internal/services/auth.go:67:12: not enough arguments in call to db.Query\n" +
		"internal/services/auth.go:89:7: cannot convert result (variable of type []byte) to type string\n" +
		"internal/services/auth.go:89:7: cannot convert result (variable of type []byte) to type string\n" +
		"# github.com/example/myapp/internal/models\n" +
		"internal/models/user.go:34:5: undefined: UserRole\n" +
		"internal/models/user.go:34:5: undefined: UserRole\n" +
		"internal/models/user.go:56:9: too many arguments in call to NewUser\n" +
		"internal/models/user.go:56:9: too many arguments in call to NewUser\n" +
		"# github.com/example/myapp/internal/middleware\n" +
		"internal/middleware/auth.go:11:2: imported and not used: \"context\"\n" +
		"internal/middleware/auth.go:11:2: imported and not used: \"context\"\n" +
		"internal/middleware/auth.go:11:2: imported and not used: \"context\"\n"

	got, err := filterGoBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Errors preserved
	if !strings.Contains(got, "undefined: Config") {
		t.Errorf("expected error message preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "main.go:12") {
		t.Errorf("expected file:line preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "9 error(s)") {
		t.Errorf("expected '9 error(s)' summary, got:\n%s", got)
	}

	// Package headers stripped
	if strings.Contains(got, "# github.com") {
		t.Errorf("expected package header stripped, got:\n%s", got)
	}

	// Duplicates stripped
	count := strings.Count(got, "undefined: Config")
	if count != 1 {
		t.Errorf("expected duplicate stripped (got %d occurrences):\n%s", count, got)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestFilterGoBuildVetErrors(t *testing.T) {
	raw := "# github.com/example/myapp\n" +
		"./main.go:25:2: printf: fmt.Sprintf format %d has arg of wrong type string\n" +
		"./handler.go:42:3: unreachable code\n" +
		"./service.go:18:5: structtag: struct field Tag has malformed tag\n"

	got, err := filterGoBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "main.go:25") {
		t.Errorf("expected vet error preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "3 error(s)") {
		t.Errorf("expected '3 error(s)' summary, got:\n%s", got)
	}
	t.Logf("output:\n%s", got)
}
