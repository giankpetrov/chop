package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGeminiInstallCreatesHook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	hookCmd := `"/usr/local/bin/openchop" hook --gemini`
	if err := geminiInstallWithCommand(path, hookCmd); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse settings: %v", err)
	}

	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("expected hooks map")
	}

	beforeTool, ok := hooks["BeforeTool"].([]interface{})
	if !ok {
		t.Fatal("expected BeforeTool array")
	}

	if len(beforeTool) != 1 {
		t.Fatalf("expected 1 BeforeTool entry, got %d", len(beforeTool))
	}

	entry := beforeTool[0].(map[string]interface{})
	if entry["matcher"] != "run_shell_command" {
		t.Errorf("expected matcher 'run_shell_command', got %v", entry["matcher"])
	}

	hooksArray := entry["hooks"].([]interface{})
	if len(hooksArray) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hooksArray))
	}

	hook := hooksArray[0].(map[string]interface{})
	if hook["command"] != hookCmd {
		t.Errorf("expected command %q, got %v", hookCmd, hook["command"])
	}
}

func TestGeminiInstallUpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Install twice — should update, not duplicate
	hookCmd := `"/usr/local/bin/openchop" hook --gemini`
	if err := geminiInstallWithCommand(path, hookCmd); err != nil {
		t.Fatalf("first install failed: %v", err)
	}

	newCmd := `"/new/path/openchop" hook --gemini`
	if err := geminiInstallWithCommand(path, newCmd); err != nil {
		t.Fatalf("second install failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	var settings map[string]interface{}
	json.Unmarshal(data, &settings)

	hooks := settings["hooks"].(map[string]interface{})
	beforeTool := hooks["BeforeTool"].([]interface{})
	entry := beforeTool[0].(map[string]interface{})
	hooksArray := entry["hooks"].([]interface{})

	if len(hooksArray) != 1 {
		t.Errorf("expected 1 hook after update, got %d", len(hooksArray))
	}

	hook := hooksArray[0].(map[string]interface{})
	if hook["command"] != newCmd {
		t.Errorf("expected updated command %q, got %v", newCmd, hook["command"])
	}
}

func TestGeminiUninstallRemovesHook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	hookCmd := `"/usr/local/bin/openchop" hook --gemini`
	geminiInstallWithCommand(path, hookCmd)

	if err := geminiUninstallFrom(path); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	var settings map[string]interface{}
	json.Unmarshal(data, &settings)

	// hooks key should be cleaned up
	if _, ok := settings["hooks"]; ok {
		t.Error("expected hooks to be removed after uninstall")
	}
}
