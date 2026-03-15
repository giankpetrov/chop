package filters

import (
	"testing"
)

func TestRegistryCoversAllExpectedCommands(t *testing.T) {
	// Spot-check that key commands resolve through the registry
	cases := []struct {
		command string
		args    []string
	}{
		{"git", []string{"status"}},
		{"git", []string{"pull"}},
		{"git", []string{"fetch"}},
		{"docker", []string{"ps"}},
		{"kubectl", []string{"get"}},
		{"terraform", []string{"plan"}},
		{"npm", []string{"install"}},
		{"go", []string{"test"}},
		{"cargo", []string{"build"}},
		{"ping", nil},
		{"curl", nil},
		{"eslint", nil},
		{"pytest", nil},
		{"make", nil},
	}

	for _, tc := range cases {
		f := get(tc.command, tc.args)
		if f == nil {
			t.Errorf("expected filter for %s %v, got nil", tc.command, tc.args)
		}
	}
}

func TestRegistryAliasesResolve(t *testing.T) {
	cases := []struct {
		alias  string
		args   []string
	}{
		{"gradlew", []string{"build"}},
		{"pip3", []string{"install"}},
		{"bundler", []string{"install"}},
		{"g++", nil},
		{"clang", nil},
		{"clang++", nil},
		{"cc", nil},
		{"c++", nil},
		{"du", nil},
	}

	for _, tc := range cases {
		f := get(tc.alias, tc.args)
		if f == nil {
			t.Errorf("expected alias %s %v to resolve, got nil", tc.alias, tc.args)
		}
	}
}

func TestRegistryHiddenNotInListBuiltins(t *testing.T) {
	builtins := ListBuiltins()
	hidden := map[string]bool{
		"cat": true, "tail": true, "less": true, "more": true,
		"ls": true, "find": true,
		"node": true, "node16": true, "node18": true, "node20": true, "node22": true,
	}

	for _, b := range builtins {
		if hidden[b.Command] {
			t.Errorf("hidden command %q should not appear in ListBuiltins", b.Command)
		}
	}
}

func TestRegistryAliasesNotInListBuiltins(t *testing.T) {
	builtins := ListBuiltins()
	aliased := map[string]bool{
		"gradlew": true, "pip3": true, "bundler": true,
		"g++": true, "cc": true, "c++": true, "clang": true, "clang++": true,
		"du": true,
	}

	for _, b := range builtins {
		if aliased[b.Command] {
			t.Errorf("alias %q should not appear in ListBuiltins", b.Command)
		}
	}
}

func TestListBuiltinsNotEmpty(t *testing.T) {
	builtins := ListBuiltins()
	if len(builtins) == 0 {
		t.Fatal("ListBuiltins returned empty")
	}
	// Should have at least the major commands
	if len(builtins) < 40 {
		t.Errorf("expected at least 40 builtin entries, got %d", len(builtins))
	}
	t.Logf("ListBuiltins returned %d entries", len(builtins))
}