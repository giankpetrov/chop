package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCursorInstallCreatesHook(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, ".cursor", "hooks.json")

	err := cursorInstallWithCommand(settingsPath, `"chop" hook --cursor`)
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

	if v, ok := settings["version"].(float64); !ok || v != 1 {
		t.Errorf("expected version 1, got %v", settings["version"])
	}

	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("hooks map missing")
	}

	preToolUse, ok := hooks["preToolUse"].([]interface{})
	if !ok {
		t.Fatal("preToolUse array missing")
	}

	found := false
	for _, entry := range preToolUse {
		m := entry.(map[string]interface{})
		if m["matcher"] == "bash" {
			hArray := m["hooks"].([]interface{})
			for _, h := range hArray {
				hMap := h.(map[string]interface{})
				if hMap["command"] == `"chop" hook --cursor` {
					found = true
				}
			}
		}
	}

	if !found {
		t.Error("chop hook not found in hooks.json")
	}
}

func TestCursorUninstallRemovesHook(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, ".cursor", "hooks.json")

	// Install first
	cursorInstallWithCommand(settingsPath, `"chop" hook --cursor`)

	// Now uninstall
	err := cursorUninstallFrom(settingsPath)
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("failed to read settings: %v", err)
	}

	var settings map[string]interface{}
	json.Unmarshal(data, &settings)

	if h, ok := settings["hooks"]; ok {
		hooks := h.(map[string]interface{})
		if p, ok := hooks["preToolUse"]; ok {
			preToolUse := p.([]interface{})
			for _, entry := range preToolUse {
				m := entry.(map[string]interface{})
				if m["matcher"] == "bash" {
					hArray := m["hooks"].([]interface{})
					for _, h := range hArray {
						hMap := h.(map[string]interface{})
						if hMap["command"] == `"chop" hook --cursor` {
							t.Error("chop hook still present after uninstall")
						}
					}
				}
			}
		}
	}
}
