package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CursorInstall registers the chop hook in ~/.cursor/hooks.json.
func CursorInstall() {
	settingsPath := cursorSettingsPath()
	if err := cursorInstallTo(settingsPath); err != nil {
		fmt.Fprintf(os.Stderr, "chop: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("chop hook installed in %s\n", settingsPath)
}

// CursorUninstall removes the chop hook from ~/.cursor/hooks.json.
func CursorUninstall() {
	settingsPath := cursorSettingsPath()
	if err := cursorUninstallFrom(settingsPath); err != nil {
		fmt.Fprintf(os.Stderr, "chop: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("chop hook removed from %s\n", settingsPath)
}

// CursorIsInstalled checks whether the chop hook is registered in ~/.cursor/hooks.json.
func CursorIsInstalled() (bool, string) {
	settingsPath := cursorSettingsPath()
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

	preToolUseRaw, ok := hooksMap["preToolUse"]
	if !ok {
		return false, settingsPath
	}
	preToolUse, ok := preToolUseRaw.([]interface{})
	if !ok {
		return false, settingsPath
	}

	for _, entry := range preToolUse {
		m, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		matcher, _ := m["matcher"].(string)
		if matcher != "bash" && matcher != "Bash" {
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
			if isChopCursorHook(hMap) {
				return true, settingsPath
			}
		}
	}

	return false, settingsPath
}

func cursorSettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cursor/hooks.json"
	}
	return filepath.Join(home, ".cursor", "hooks.json")
}

func buildCursorHookCommand() (string, error) {
	binPath, err := chopBinaryPath()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`"%s" hook --cursor`, binPath), nil
}

func isChopCursorHook(hookObj map[string]interface{}) bool {
	cmd, ok := hookObj["command"].(string)
	if !ok {
		return false
	}
	return strings.Contains(cmd, chopBinaryName) && strings.Contains(cmd, "hook") && strings.Contains(cmd, "--cursor")
}

func cursorInstallTo(settingsPath string) error {
	hookCmd, err := buildCursorHookCommand()
	if err != nil {
		return err
	}
	return cursorInstallWithCommand(settingsPath, hookCmd)
}

func cursorInstallWithCommand(settingsPath string, hookCmd string) error {
	settings, err := readSettings(settingsPath)
	if err != nil {
		return err
	}

	// Ensure version is set (required for Cursor hooks.json)
	if _, ok := settings["version"]; !ok {
		settings["version"] = 1
	}

	// Ensure hooks map exists
	hooksRaw, ok := settings["hooks"]
	if !ok {
		hooksRaw = make(map[string]interface{})
		settings["hooks"] = hooksRaw
	}
	hooksMap, ok := hooksRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("hooks field is not an object in %s", settingsPath)
	}

	// Ensure preToolUse array exists (note: Cursor docs show preToolUse with lowercase p)
	preToolUseRaw, ok := hooksMap["preToolUse"]
	if !ok {
		preToolUseRaw = []interface{}{}
	}
	preToolUse, ok := preToolUseRaw.([]interface{})
	if !ok {
		return fmt.Errorf("hooks.preToolUse is not an array in %s", settingsPath)
	}

	chopHookEntry := map[string]interface{}{
		"type":    "command",
		"command": hookCmd,
	}

	// Find existing bash matcher
	bashIdx := -1
	for i, entry := range preToolUse {
		m, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		matcher, _ := m["matcher"].(string)
		if matcher == "bash" || matcher == "Bash" {
			bashIdx = i
			break
		}
	}

	if bashIdx >= 0 {
		bashMatcher := preToolUse[bashIdx].(map[string]interface{})
		hooksArrayRaw, ok := bashMatcher["hooks"]
		if !ok {
			hooksArrayRaw = []interface{}{}
		}
		hooksArray, ok := hooksArrayRaw.([]interface{})
		if !ok {
			return fmt.Errorf("bash matcher hooks is not an array")
		}

		// Check if chop hook already exists — update it
		chopIdx := -1
		for i, h := range hooksArray {
			hMap, ok := h.(map[string]interface{})
			if !ok {
				continue
			}
			if isChopCursorHook(hMap) {
				chopIdx = i
				break
			}
		}

		if chopIdx >= 0 {
			hooksArray[chopIdx] = chopHookEntry
		} else {
			hooksArray = append(hooksArray, chopHookEntry)
		}
		bashMatcher["hooks"] = hooksArray
	} else {
		// Create new bash matcher
		newMatcher := map[string]interface{}{
			"matcher": "bash",
			"hooks": []interface{}{
				chopHookEntry,
			},
		}
		preToolUse = append(preToolUse, newMatcher)
	}

	hooksMap["preToolUse"] = preToolUse
	settings["hooks"] = hooksMap

	return writeSettings(settingsPath, settings)
}

func cursorUninstallFrom(settingsPath string) error {
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

	preToolUseRaw, ok := hooksMap["preToolUse"]
	if !ok {
		return nil
	}
	preToolUse, ok := preToolUseRaw.([]interface{})
	if !ok {
		return nil
	}

	newPreToolUse := make([]interface{}, 0, len(preToolUse))
	for _, entry := range preToolUse {
		m, ok := entry.(map[string]interface{})
		if !ok {
			newPreToolUse = append(newPreToolUse, entry)
			continue
		}
		matcher, _ := m["matcher"].(string)
		if matcher != "bash" && matcher != "Bash" {
			newPreToolUse = append(newPreToolUse, entry)
			continue
		}

		hooksArrayRaw, ok := m["hooks"]
		if !ok {
			newPreToolUse = append(newPreToolUse, entry)
			continue
		}
		hooksArray, ok := hooksArrayRaw.([]interface{})
		if !ok {
			newPreToolUse = append(newPreToolUse, entry)
			continue
		}

		newHooks := make([]interface{}, 0, len(hooksArray))
		for _, h := range hooksArray {
			hMap, ok := h.(map[string]interface{})
			if !ok {
				newHooks = append(newHooks, h)
				continue
			}
			if !isChopCursorHook(hMap) {
				newHooks = append(newHooks, h)
			}
		}

		if len(newHooks) > 0 {
			m["hooks"] = newHooks
			newPreToolUse = append(newPreToolUse, m)
		}
	}

	if len(newPreToolUse) > 0 {
		hooksMap["preToolUse"] = newPreToolUse
	} else {
		delete(hooksMap, "preToolUse")
	}

	if len(hooksMap) == 0 {
		delete(settings, "hooks")
		delete(settings, "version")
	} else {
		settings["hooks"] = hooksMap
	}

	return writeSettings(settingsPath, settings)
}
