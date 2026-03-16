package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCodexInstallCreatesHook(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, ".codex", "settings.json")

	err := codexInstallWithCommand(settingsPath, `"chop" hook --codex`)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to unmarshal settings: %v", err)
	}

	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("hooks map missing")
	}

	preToolUse, ok := hooks["PreToolUse"].([]interface{})
	if !ok {
		t.Fatal("PreToolUse array missing")
	}

	found := false
	for _, entry := range preToolUse {
		m := entry.(map[string]interface{})
		if m["matcher"] == "bash" {
			hArray := m["hooks"].([]interface{})
			for _, h := range hArray {
				hMap := h.(map[string]interface{})
				if hMap["command"] == `"chop" hook --codex` {
					found = true
				}
			}
		}
	}

	if !found {
		t.Error("chop hook not found in settings.json")
	}
}

func TestCodexUninstallRemovesHook(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, ".codex", "settings.json")

	// Install first
	codexInstallWithCommand(settingsPath, `"chop" hook --codex`)

	// Now uninstall
	err := codexUninstallFrom(settingsPath)
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		// If file doesn't exist, it's also fine as it was empty before
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("failed to read settings: %v", err)
	}

	var settings map[string]interface{}
	json.Unmarshal(data, &settings)

	if h, ok := settings["hooks"]; ok {
		hooks := h.(map[string]interface{})
		if p, ok := hooks["PreToolUse"]; ok {
			preToolUse := p.([]interface{})
			for _, entry := range preToolUse {
				m := entry.(map[string]interface{})
				if m["matcher"] == "bash" {
					hArray := m["hooks"].([]interface{})
					for _, h := range hArray {
						hMap := h.(map[string]interface{})
						if hMap["command"] == `"chop" hook --codex` {
							t.Error("chop hook still present after uninstall")
						}
					}
				}
			}
		}
	}
}
