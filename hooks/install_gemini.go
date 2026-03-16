package hooks

import (
	"fmt"
	"os"
	"path/filepath"
)

// InstallGemini registers the chop hook in ~/.gemini/settings.json.
func InstallGemini() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "chop: failed to get home directory: %v\n", err)
		os.Exit(1)
	}
	settingsPath := filepath.Join(home, ".gemini", "settings.json")
	if err := installGeminiTo(settingsPath); err != nil {
		fmt.Fprintf(os.Stderr, "chop: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("chop hook installed in %s\n", settingsPath)
}

// UninstallGemini removes the chop hook from ~/.gemini/settings.json.
func UninstallGemini() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "chop: failed to get home directory: %v\n", err)
		os.Exit(1)
	}
	settingsPath := filepath.Join(home, ".gemini", "settings.json")
	if err := uninstallGeminiFrom(settingsPath); err != nil {
		fmt.Fprintf(os.Stderr, "chop: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("chop hook removed from %s\n", settingsPath)
}

func installGeminiTo(settingsPath string) error {
	hookCmd, err := buildHookCommand()
	if err != nil {
		return err
	}

	settings, err := readSettings(settingsPath)
	if err != nil {
		return err
	}

	hooksRaw, ok := settings["hooks"]
	if !ok {
		hooksRaw = make(map[string]interface{})
		settings["hooks"] = hooksRaw
	}
	hooksMap, ok := hooksRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("hooks field is not an object in %s", settingsPath)
	}

	beforeToolRaw, ok := hooksMap["BeforeTool"]
	if !ok {
		beforeToolRaw = []interface{}{}
	}
	beforeTool, ok := beforeToolRaw.([]interface{})
	if !ok {
		return fmt.Errorf("hooks.BeforeTool is not an array in %s", settingsPath)
	}

	chopHookEntry := map[string]interface{}{
		"type":    "command",
		"command": hookCmd,
		"name":    "chop",
	}

	// Find existing run_shell_command matcher
	shellIdx := -1
	for i, entry := range beforeTool {
		m, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		if matcher, _ := m["matcher"].(string); matcher == "run_shell_command" {
			shellIdx = i
			break
		}
	}

	if shellIdx >= 0 {
		shellMatcher := beforeTool[shellIdx].(map[string]interface{})
		hooksArrayRaw, ok := shellMatcher["hooks"]
		if !ok {
			hooksArrayRaw = []interface{}{}
		}
		hooksArray, ok := hooksArrayRaw.([]interface{})
		if !ok {
			return fmt.Errorf("run_shell_command matcher hooks is not an array")
		}

		chopIdx := -1
		for i, h := range hooksArray {
			hMap, ok := h.(map[string]interface{})
			if !ok {
				continue
			}
			if isChopHook(hMap) {
				chopIdx = i
				break
			}
		}

		if chopIdx >= 0 {
			hooksArray[chopIdx] = chopHookEntry
		} else {
			hooksArray = append(hooksArray, chopHookEntry)
		}
		shellMatcher["hooks"] = hooksArray
	} else {
		shellMatcher := map[string]interface{}{
			"matcher": "run_shell_command",
			"hooks": []interface{}{
				chopHookEntry,
			},
		}
		beforeTool = append(beforeTool, shellMatcher)
	}

	hooksMap["BeforeTool"] = beforeTool
	settings["hooks"] = hooksMap

	return writeSettings(settingsPath, settings)
}

func uninstallGeminiFrom(settingsPath string) error {
	settings, err := readSettings(settingsPath)
	if err != nil {
		return err
	}

	hooksRaw, ok := settings["hooks"]
	if !ok {
		return nil
	}
	hooksMap, ok := hooksRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	beforeToolRaw, ok := hooksMap["BeforeTool"]
	if !ok {
		return nil
	}
	beforeTool, ok := beforeToolRaw.([]interface{})
	if !ok {
		return nil
	}

	newBeforeTool := make([]interface{}, 0, len(beforeTool))
	for _, entry := range beforeTool {
		m, ok := entry.(map[string]interface{})
		if !ok {
			newBeforeTool = append(newBeforeTool, entry)
			continue
		}
		matcher, _ := m["matcher"].(string)
		if matcher != "run_shell_command" {
			newBeforeTool = append(newBeforeTool, entry)
			continue
		}

		hooksArrayRaw, ok := m["hooks"]
		if !ok {
			newBeforeTool = append(newBeforeTool, entry)
			continue
		}
		hooksArray, ok := hooksArrayRaw.([]interface{})
		if !ok {
			newBeforeTool = append(newBeforeTool, entry)
			continue
		}

		newHooks := make([]interface{}, 0, len(hooksArray))
		for _, h := range hooksArray {
			hMap, ok := h.(map[string]interface{})
			if !ok {
				newHooks = append(newHooks, h)
				continue
			}
			if !isChopHook(hMap) {
				newHooks = append(newHooks, h)
			}
		}

		if len(newHooks) > 0 {
			m["hooks"] = newHooks
			newBeforeTool = append(newBeforeTool, m)
		}
	}

	if len(newBeforeTool) > 0 {
		hooksMap["BeforeTool"] = newBeforeTool
	} else {
		delete(hooksMap, "BeforeTool")
	}

	if len(hooksMap) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooksMap
	}

	return writeSettings(settingsPath, settings)
}

// IsGeminiInstalled checks whether the chop hook is registered in ~/.gemini/settings.json.
func IsGeminiInstalled() (bool, string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, ""
	}
	settingsPath := filepath.Join(home, ".gemini", "settings.json")
	settings, err := readSettings(settingsPath)
	if err != nil {
		return false, settingsPath
	}

	hooksRaw, ok := settings["hooks"]
	if !ok {
		return false, settingsPath
	}
	hooksMap, ok := hooksRaw.(map[string]interface{})
	if !ok {
		return false, settingsPath
	}

	beforeToolRaw, ok := hooksMap["BeforeTool"]
	if !ok {
		return false, settingsPath
	}
	beforeTool, ok := beforeToolRaw.([]interface{})
	if !ok {
		return false, settingsPath
	}

	for _, entry := range beforeTool {
		m, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		if matcher, _ := m["matcher"].(string); matcher != "run_shell_command" {
			continue
		}
		hooksArrayRaw, ok := m["hooks"]
		if !ok {
			continue
		}
		hooksArray, ok := hooksArrayRaw.([]interface{})
		if !ok {
			continue
		}
		for _, h := range hooksArray {
			hMap, ok := h.(map[string]interface{})
			if !ok {
				continue
			}
			if isChopHook(hMap) {
				return true, settingsPath
			}
		}
	}

	return false, settingsPath
}
