package shell

import (
	"strings"
	"testing"
)

func TestGenerateInitBash(t *testing.T) {
	out := GenerateInit("bash")

	mustContain := []string{
		"__chop_wrap()",
		"command chop \"$@\"",
		"git() { __chop_wrap git \"$@\"; }",
		"docker() { __chop_wrap docker \"$@\"; }",
		"kubectl() { __chop_wrap kubectl \"$@\"; }",
		"npm() { __chop_wrap npm \"$@\"; }",
		"cargo() { __chop_wrap cargo \"$@\"; }",
		"go() { __chop_wrap go \"$@\"; }",
		"aws() { __chop_wrap aws \"$@\"; }",
		"curl() { __chop_wrap curl \"$@\"; }",
		"pytest() { __chop_wrap pytest \"$@\"; }",
		"make() { __chop_wrap make \"$@\"; }",
		"ping() { __chop_wrap ping \"$@\"; }",
		"unchop() { command \"$@\"; }",
	}

	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("bash output missing %q", s)
		}
	}
}

func TestGenerateInitZsh(t *testing.T) {
	// zsh should produce the same output as bash
	bash := GenerateInit("bash")
	zsh := GenerateInit("zsh")
	if bash != zsh {
		t.Error("bash and zsh output should be identical")
	}
}

func TestGenerateInitFish(t *testing.T) {
	out := GenerateInit("fish")

	mustContain := []string{
		"function __chop_wrap",
		"command chop $argv",
		"function git; __chop_wrap git $argv; end",
		"function docker; __chop_wrap docker $argv; end",
		"function kubectl; __chop_wrap kubectl $argv; end",
		"function npm; __chop_wrap npm $argv; end",
		"function unchop; command $argv; end",
	}

	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("fish output missing %q", s)
		}
	}

	// Fish should not contain bash syntax
	if strings.Contains(out, "\"$@\"") {
		t.Error("fish output should not contain bash-style \"$@\"")
	}
}

func TestGenerateInitUnsupported(t *testing.T) {
	out := GenerateInit("tcsh")
	if !strings.Contains(out, "unsupported shell") {
		t.Error("unsupported shell should produce error message")
	}
}

func TestGenerateInitPowerShell(t *testing.T) {
	out := GenerateInit("powershell")

	mustContain := []string{
		"function git { chop git @args }",
		"function docker { chop docker @args }",
		"function kubectl { chop kubectl @args }",
		"function npm { chop npm @args }",
		"function unchop {",
	}

	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("powershell output missing %q", s)
		}
	}

	// Should not contain bash syntax
	if strings.Contains(out, "\"$@\"") {
		t.Error("powershell output should not contain bash-style \"$@\"")
	}
}

func TestGenerateInitPwsh(t *testing.T) {
	// pwsh should produce the same output as powershell
	ps := GenerateInit("powershell")
	pwsh := GenerateInit("pwsh")
	if ps != pwsh {
		t.Error("powershell and pwsh output should be identical")
	}
}

func TestAllCommandsPresent(t *testing.T) {
	out := GenerateInit("bash")

	// Commands with + in name (g++, c++, clang++) cannot be bash functions
	skipBash := map[string]bool{
		"g++": true, "c++": true, "clang++": true,
	}

	for _, cmd := range commands {
		if skipBash[cmd] {
			continue
		}
		expected := cmd + "() { __chop_wrap " + cmd
		if !strings.Contains(out, expected) {
			t.Errorf("bash output missing wrapper for %q", cmd)
		}
	}
}

func TestPlusPlusCommandsSkipped(t *testing.T) {
	out := GenerateInit("bash")

	// g++, c++, clang++ can't be bash function names
	for _, cmd := range []string{"g++", "c++", "clang++"} {
		fn := cmd + "() {"
		if strings.Contains(out, fn) {
			t.Errorf("bash output should NOT contain function %q (invalid bash identifier)", cmd)
		}
	}
}
