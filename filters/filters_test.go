package filters

import (
	"reflect"
	"testing"
)

func TestSkipGitGlobalFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "no flags",
			args:     []string{"status"},
			expected: []string{"status"},
		},
		{
			name:     "single dash C flag",
			args:     []string{"-C", "/tmp", "status"},
			expected: []string{"status"},
		},
		{
			name:     "multiple flags",
			args:     []string{"--no-pager", "-C", "/tmp", "log", "-n", "1"},
			expected: []string{"log", "-n", "1"},
		},
		{
			name:     "flags with equals",
			args:     []string{"--git-dir=/tmp/.git", "status"},
			expected: []string{"status"},
		},
		{
			name:     "mixed flags",
			args:     []string{"--bare", "-c", "user.name=test", "rev-parse", "HEAD"},
			expected: []string{"rev-parse", "HEAD"},
		},
		{
			name:     "no subcommand",
			args:     []string{"--version"},
			expected: []string{"--version"}, // doesn't match any known flag to skip, stops there
		},
		{
			name:     "unknown flag",
			args:     []string{"--unknown", "status"},
			expected: []string{"--unknown", "status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := skipGitGlobalFlags(tt.args)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("skipGitGlobalFlags(%v) = %v, want %v", tt.args, got, tt.expected)
			}
		})
	}
}

func TestGetRouters(t *testing.T) {
	tests := []struct {
		command string
		args    []string
	}{
		{"docker", []string{}},
		{"docker", []string{"ps"}},
		{"docker", []string{"build"}},
		{"docker", []string{"images"}},
		{"docker", []string{"logs"}},
		{"docker", []string{"rmi"}},
		{"docker", []string{"inspect"}},
		{"docker", []string{"stats"}},
		{"docker", []string{"top"}},
		{"docker", []string{"diff"}},
		{"docker", []string{"history"}},
		{"docker", []string{"network"}},
		{"docker", []string{"network", "ls"}},
		{"docker", []string{"volume"}},
		{"docker", []string{"volume", "ls"}},
		{"docker", []string{"system"}},
		{"docker", []string{"system", "df"}},
		{"docker", []string{"compose"}},
		{"docker", []string{"compose", "ps"}},
		{"docker", []string{"compose", "build"}},
		{"docker", []string{"compose", "logs"}},
		{"docker", []string{"compose", "images"}},
		{"git", []string{"status"}},
		{"git", []string{"log"}},
		{"git", []string{"diff"}},
		{"git", []string{"show"}},
		{"git", []string{"branch"}},
		{"git", []string{"push"}},
		{"git", []string{"pull"}},
		{"git", []string{"fetch"}},
		{"git", []string{"remote"}},
		{"git", []string{"tag"}},
		{"git", []string{"checkout"}},
		{"git", []string{"reset"}},
		{"git", []string{"stash"}},
		{"git", []string{"stash", "list"}},
		{"npm", []string{"install"}},
		{"npm", []string{"update"}},
		{"npm", []string{"list"}},
		{"npm", []string{"view"}},
		{"npm", []string{"i"}},
		{"npm", []string{"up"}},
		{"npm", []string{"upgrade"}},
		{"npm", []string{"info"}},
		{"npm", []string{"show"}},
		{"npm", []string{"test"}},
		{"npm", []string{"t"}},
		{"npm", []string{"run"}},
		{"npm", []string{"run", "test"}},
		{"npm", []string{"run", "build"}},
		{"npm", []string{"run", "lint"}},
		{"dotnet", []string{}},
		{"dotnet", []string{"build"}},
		{"dotnet", []string{"test"}},
		{"kubectl", []string{}},
		{"kubectl", []string{"get"}},
		{"kubectl", []string{"describe"}},
		{"kubectl", []string{"logs"}},
		{"kubectl", []string{"log"}},
		{"kubectl", []string{"top"}},
		{"kubectl", []string{"apply"}},
		{"kubectl", []string{"delete"}},
		{"helm", []string{}},
		{"helm", []string{"install"}},
		{"helm", []string{"upgrade"}},
		{"helm", []string{"list"}},
		{"helm", []string{"ls"}},
		{"helm", []string{"status"}},
		{"npx", []string{"playwright", "test"}},
		{"npx", []string{"tsc"}},
		{"npx", []string{}},
		{"npx", []string{"jest"}},
		{"npx", []string{"nx", "build"}},
		{"npx", []string{"ng", "build"}},
		{"npx", []string{"playwright"}},
		{"npx", []string{"playwright", "test"}},
		{"cargo", []string{}},
		{"cargo", []string{"test"}},
		{"cargo", []string{"build"}},
		{"cargo", []string{"check"}},
		{"cargo", []string{"clippy"}},
		{"go", []string{}},
		{"go", []string{"test"}},
		{"go", []string{"build"}},
		{"go", []string{"vet"}},
		{"gh", []string{}},
		{"gh", []string{"pr", "list"}},
		{"gh", []string{"issue", "list"}},
		{"gh", []string{"run", "list"}},
		{"terraform", []string{}},
		{"terraform", []string{"plan"}},
		{"terraform", []string{"apply"}},
		{"terraform", []string{"init"}},
		{"pnpm", []string{}},
		{"pnpm", []string{"install"}},
		{"pnpm", []string{"i"}},
		{"pnpm", []string{"add"}},
		{"pnpm", []string{"list"}},
		{"pnpm", []string{"ls"}},
		{"pnpm", []string{"test"}},
		{"pnpm", []string{"t"}},
		{"yarn", []string{}},
		{"yarn", []string{"install"}},
		{"yarn", []string{"add"}},
		{"yarn", []string{"list"}},
		{"yarn", []string{"test"}},
		{"bun", []string{}},
		{"bun", []string{"install"}},
		{"bun", []string{"i"}},
		{"bun", []string{"add"}},
		{"bun", []string{"test"}},
		{"bun", []string{"t"}},
		{"ng", []string{}},
		{"ng", []string{"build"}},
		{"ng", []string{"b"}},
		{"ng", []string{"test"}},
		{"ng", []string{"t"}},
		{"ng", []string{"serve"}},
		{"ng", []string{"s"}},
		{"ng", []string{"lint"}},
		{"nx", []string{}},
		{"nx", []string{"build"}},
		{"nx", []string{"run"}},
		{"nx", []string{"test"}},
		{"nx", []string{"lint"}},
		{"pip", []string{}},
		{"pip", []string{"install"}},
		{"pip", []string{"list"}},
		{"uv", []string{"pip"}},
		{"uv", []string{"pip", "install"}},
		{"uv", []string{"install"}},
		{"uv", []string{"add"}},
		{"uv", []string{"pip", "list"}},
		{"bundle", []string{}},
		{"bundle", []string{"install"}},
		{"composer", []string{}},
		{"composer", []string{"install"}},
		{"composer", []string{"update"}},
		{"composer", []string{"require"}},
	}

	for _, tt := range tests {
		name := tt.command
		if len(tt.args) > 0 {
			name += "_" + tt.args[0]
		}
		t.Run(name, func(t *testing.T) {
			// Some of these are expected to return nil, but they should at least not panic
			// and we want to exercise the code.
			_ = Get(tt.command, tt.args)
		})
	}
}

func TestGetBuiltin(t *testing.T) {
	tests := []struct {
		command string
		args    []string
		wantNil bool
	}{
		{"git", []string{"status"}, false},
		{"git", []string{"-C", "dir", "log"}, false},
		{"docker", []string{"ps"}, false},
		{"docker", []string{"compose", "ps"}, false},
		{"npm", []string{"install"}, false},
		{"kubectl", []string{"get", "pods"}, false},
		{"unknown", nil, true},
		{"git", []string{"unknown-subcommand"}, true},
		{"docker", []string{"unknown-subcommand"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			fn := Get(tt.command, tt.args)
			if (fn == nil) != tt.wantNil {
				t.Errorf("Get(%q, %v) nil = %v, want %v", tt.command, tt.args, fn == nil, tt.wantNil)
			}
		})
	}
}

func TestHasFilterBuiltin(t *testing.T) {
	tests := []struct {
		command string
		args    []string
		want    bool
	}{
		{"git", []string{"status"}, true},
		{"docker", []string{"build"}, true},
		{"npm", []string{"list"}, true},
		{"go", []string{"test"}, true},
		{"unknown", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := HasFilter(tt.command, tt.args)
			if got != tt.want {
				t.Errorf("HasFilter(%q, %v) = %v, want %v", tt.command, tt.args, got, tt.want)
			}
		})
	}
}

func TestGetDockerComposeFilter_EdgeCases(t *testing.T) {
	// Test the case where docker compose is called without subcommands
	fn := getDockerComposeFilter([]string{})
	if fn == nil {
		t.Error("expected non-nil filter for empty docker compose args")
	}

	fn = getDockerComposeFilter([]string{"unknown"})
	if fn != nil {
		t.Error("expected nil filter for unknown docker compose subcommand")
	}
}

func TestGetGitFilter_EdgeCases(t *testing.T) {
	if getGitFilter([]string{}) != nil {
		t.Error("expected nil for empty git args")
	}
	if getGitFilter([]string{"--no-pager"}) != nil {
		t.Error("expected nil for git with only global flags")
	}
}
