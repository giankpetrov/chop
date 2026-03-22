package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testHookCmd = `"/usr/local/bin/chop" hook`

func tempSettingsPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, ".claude", "settings.json")
}

func readJSON(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	return m
}

func writeJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
}

func TestInstallCreatesCorrectStructure(t *testing.T) {
	path := tempSettingsPath(t)

	if err := installWithCommand(path, testHookCmd); err != nil {
		t.Fatalf("installWithCommand failed: %v", err)
	}

	settings := readJSON(t, path)

	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("hooks not found or not an object")
	}

	preToolUse, ok := hooks["PreToolUse"].([]interface{})
	if !ok {
		t.Fatal("PreToolUse not found or not an array")
	}

	if len(preToolUse) != 1 {
		t.Fatalf("expected 1 PreToolUse entry, got %d", len(preToolUse))
	}

	matcher := preToolUse[0].(map[string]interface{})
	if matcher["matcher"] != "Bash" {
		t.Errorf("expected matcher 'Bash', got %v", matcher["matcher"])
	}

	hooksArr := matcher["hooks"].([]interface{})
	if len(hooksArr) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hooksArr))
	}

	hookEntry := hooksArr[0].(map[string]interface{})
	if hookEntry["type"] != "command" {
		t.Errorf("expected type 'command', got %v", hookEntry["type"])
	}

	cmd, _ := hookEntry["command"].(string)
	if cmd != testHookCmd {
		t.Errorf("expected command %q, got %q", testHookCmd, cmd)
	}
}

func TestInstallPreservesExistingSettings(t *testing.T) {
	path := tempSettingsPath(t)

	existing := map[string]interface{}{
		"apiKey":      "sk-test-123",
		"permissions": []interface{}{"allow_all"},
		"model":       "claude-opus-4-20250514",
	}
	writeJSON(t, path, existing)

	if err := installWithCommand(path, testHookCmd); err != nil {
		t.Fatalf("installWithCommand failed: %v", err)
	}

	settings := readJSON(t, path)

	if settings["apiKey"] != "sk-test-123" {
		t.Errorf("apiKey was lost, got %v", settings["apiKey"])
	}
	if settings["model"] != "claude-opus-4-20250514" {
		t.Errorf("model was lost, got %v", settings["model"])
	}
	perms, ok := settings["permissions"].([]interface{})
	if !ok || len(perms) != 1 || perms[0] != "allow_all" {
		t.Errorf("permissions were lost, got %v", settings["permissions"])
	}

	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("hooks not added")
	}
	preToolUse, ok := hooks["PreToolUse"].([]interface{})
	if !ok || len(preToolUse) == 0 {
		t.Fatal("PreToolUse not added")
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	path := tempSettingsPath(t)

	if err := installWithCommand(path, testHookCmd); err != nil {
		t.Fatalf("first install failed: %v", err)
	}
	if err := installWithCommand(path, testHookCmd); err != nil {
		t.Fatalf("second install failed: %v", err)
	}

	settings := readJSON(t, path)
	hooks := settings["hooks"].(map[string]interface{})
	preToolUse := hooks["PreToolUse"].([]interface{})

	if len(preToolUse) != 1 {
		t.Fatalf("expected 1 PreToolUse entry after double install, got %d", len(preToolUse))
	}

	bashMatcher := preToolUse[0].(map[string]interface{})
	hooksArr := bashMatcher["hooks"].([]interface{})
	if len(hooksArr) != 1 {
		t.Fatalf("expected 1 hook after double install, got %d", len(hooksArr))
	}
}

func TestInstallPreservesOtherMatchers(t *testing.T) {
	path := tempSettingsPath(t)

	existing := map[string]interface{}{
		"hooks": map[string]interface{}{
			"PreToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Read",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "some-other-tool",
						},
					},
				},
			},
		},
	}
	writeJSON(t, path, existing)

	if err := installWithCommand(path, testHookCmd); err != nil {
		t.Fatalf("installWithCommand failed: %v", err)
	}

	settings := readJSON(t, path)
	hooks := settings["hooks"].(map[string]interface{})
	preToolUse := hooks["PreToolUse"].([]interface{})

	if len(preToolUse) != 2 {
		t.Fatalf("expected 2 PreToolUse entries (Read + Bash), got %d", len(preToolUse))
	}

	found := false
	for _, entry := range preToolUse {
		m := entry.(map[string]interface{})
		if m["matcher"] == "Read" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Read matcher was lost")
	}
}

func TestInstallAddsToBashMatcherWithOtherHooks(t *testing.T) {
	path := tempSettingsPath(t)

	existing := map[string]interface{}{
		"hooks": map[string]interface{}{
			"PreToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Bash",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "some-linter hook",
						},
					},
				},
			},
		},
	}
	writeJSON(t, path, existing)

	if err := installWithCommand(path, testHookCmd); err != nil {
		t.Fatalf("installWithCommand failed: %v", err)
	}

	settings := readJSON(t, path)
	hooks := settings["hooks"].(map[string]interface{})
	preToolUse := hooks["PreToolUse"].([]interface{})
	bashMatcher := preToolUse[0].(map[string]interface{})
	hooksArr := bashMatcher["hooks"].([]interface{})

	if len(hooksArr) != 2 {
		t.Fatalf("expected 2 hooks in Bash matcher, got %d", len(hooksArr))
	}
}

func TestUninstallRemovesChopHook(t *testing.T) {
	path := tempSettingsPath(t)

	if err := installWithCommand(path, testHookCmd); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	if err := uninstallFrom(path); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	settings := readJSON(t, path)

	if _, ok := settings["hooks"]; ok {
		t.Error("hooks key should have been removed when empty")
	}
}

func TestUninstallPreservesOtherHooks(t *testing.T) {
	path := tempSettingsPath(t)

	existing := map[string]interface{}{
		"hooks": map[string]interface{}{
			"PreToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Bash",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "some-linter hook",
						},
						map[string]interface{}{
							"type":    "command",
							"command": `"/usr/local/bin/chop" hook`,
						},
					},
				},
			},
		},
	}
	writeJSON(t, path, existing)

	if err := uninstallFrom(path); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	settings := readJSON(t, path)
	hooks := settings["hooks"].(map[string]interface{})
	preToolUse := hooks["PreToolUse"].([]interface{})

	if len(preToolUse) != 1 {
		t.Fatalf("expected 1 PreToolUse entry, got %d", len(preToolUse))
	}

	bashMatcher := preToolUse[0].(map[string]interface{})
	hooksArr := bashMatcher["hooks"].([]interface{})
	if len(hooksArr) != 1 {
		t.Fatalf("expected 1 hook after uninstall, got %d", len(hooksArr))
	}

	remaining := hooksArr[0].(map[string]interface{})
	if remaining["command"] != "some-linter hook" {
		t.Errorf("wrong hook preserved: %v", remaining["command"])
	}
}

func TestUninstallPreservesOtherMatchers(t *testing.T) {
	path := tempSettingsPath(t)

	existing := map[string]interface{}{
		"hooks": map[string]interface{}{
			"PreToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Bash",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": `"/usr/local/bin/chop" hook`,
						},
					},
				},
				map[string]interface{}{
					"matcher": "Read",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "some-read-hook",
						},
					},
				},
			},
		},
	}
	writeJSON(t, path, existing)

	if err := uninstallFrom(path); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	settings := readJSON(t, path)
	hooks := settings["hooks"].(map[string]interface{})
	preToolUse := hooks["PreToolUse"].([]interface{})

	if len(preToolUse) != 1 {
		t.Fatalf("expected 1 PreToolUse entry (Read), got %d", len(preToolUse))
	}

	matcher := preToolUse[0].(map[string]interface{})
	if matcher["matcher"] != "Read" {
		t.Errorf("expected Read matcher to be preserved, got %v", matcher["matcher"])
	}
}

func TestUninstallNoSettingsFile(t *testing.T) {
	path := tempSettingsPath(t)

	if err := uninstallFrom(path); err != nil {
		t.Fatalf("uninstall on missing file failed: %v", err)
	}
}

func TestIsChopHook(t *testing.T) {
	tests := []struct {
		command string
		want    bool
	}{
		{`"/usr/local/bin/chop" hook`, true},
		{`"C:/Users/me/bin/chop" hook`, true},
		{`chop hook`, true},
		{`"/path/to/chop.exe" hook`, true},
		{"some-linter hook", false},
		{"other-command", false},
	}
	for _, tt := range tests {
		hookObj := map[string]interface{}{"command": tt.command}
		got := isChopHook(hookObj)
		if got != tt.want {
			t.Errorf("isChopHook(%q) = %v, want %v", tt.command, got, tt.want)
		}
	}
}

func TestBuildHookCommandFormat(t *testing.T) {
	cmd, err := buildHookCommand()
	if err != nil {
		t.Fatalf("buildHookCommand failed: %v", err)
	}
	if !strings.HasSuffix(cmd, " hook") {
		t.Errorf("hook command should end with ' hook', got %q", cmd)
	}
	if !strings.HasPrefix(cmd, `"`) {
		t.Errorf("hook command should start with quote, got %q", cmd)
	}
	if strings.Contains(cmd, `\`) {
		t.Errorf("hook command should not contain backslashes, got %q", cmd)
	}
}
