package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// cursorHookInput represents the JSON payload from Cursor IDE's PreToolUse hook.
type cursorHookInput struct {
	SessionID     string          `json:"session_id"`
	Cwd           string          `json:"cwd"`
	HookEventName string          `json:"hook_event_name"`
	ToolName      string          `json:"tool_name"`
	ToolInput     json.RawMessage `json:"tool_input"`
}

// cursorToolInput matches Cursor IDE's bash tool input.
type cursorToolInput struct {
	Command string `json:"command"`
}

type cursorHookOutput struct {
	HookSpecificOutput cursorHookSpecificOutput `json:"hookSpecificOutput"`
}

type cursorHookSpecificOutput struct {
	HookEventName      string          `json:"hookEventName"`
	PermissionDecision string          `json:"permissionDecision"`
	UpdatedInput       cursorToolInput `json:"updatedInput"`
}

// RunCursorHook reads a Cursor IDE PreToolUse hook payload from stdin,
// checks if the command should be wrapped with chop, and outputs
// modified JSON on stdout. Always exits 0.
func RunCursorHook() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(0)
	}

	output, shouldModify, original := processCursorHookInput(input)
	if shouldModify {
		var result cursorHookOutput
		if err := json.Unmarshal(output, &result); err == nil {
			auditLog(original, result.HookSpecificOutput.UpdatedInput.Command)
		}
		fmt.Print(string(output))
	}
	// If not modifying, output nothing (passthrough)
}

// processCursorHookInput parses the Cursor IDE hook JSON and determines whether
// to wrap the command.
func processCursorHookInput(input []byte) ([]byte, bool, string) {
	var h cursorHookInput
	if err := json.Unmarshal(input, &h); err != nil {
		return nil, false, ""
	}

	// Cursor uses "bash" or "Bash"
	if h.ToolName != "bash" && h.ToolName != "Bash" {
		return nil, false, ""
	}

	// Global kill switch
	if IsDisabledGlobally() {
		return nil, false, ""
	}

	var ti cursorToolInput
	if err := json.Unmarshal(h.ToolInput, &ti); err != nil {
		return nil, false, ""
	}

	wrapped, shouldModify, original := rewriteCommand(ti.Command)
	if !shouldModify {
		return nil, false, original
	}

	return buildCursorOutput(original, wrapped)
}

// buildCursorOutput constructs the Cursor IDE hook JSON response.
func buildCursorOutput(original, wrapped string) ([]byte, bool, string) {
	out := cursorHookOutput{
		HookSpecificOutput: cursorHookSpecificOutput{
			HookEventName:      "preToolUse",
			PermissionDecision: "allow",
			UpdatedInput: cursorToolInput{
				Command: wrapped,
			},
		},
	}
	data, err := json.Marshal(out)
	if err != nil {
		return nil, false, original
	}
	return data, true, original
}
