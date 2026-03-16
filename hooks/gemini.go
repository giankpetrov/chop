package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// geminiHookInput represents the JSON payload from Gemini CLI's BeforeTool hook.
type geminiHookInput struct {
	SessionID     string          `json:"session_id"`
	Cwd           string          `json:"cwd"`
	HookEventName string          `json:"hook_event_name"`
	ToolName      string          `json:"tool_name"`
	ToolInput     json.RawMessage `json:"tool_input"`
}

// geminiToolInput matches Gemini CLI's run_shell_command tool input.
type geminiToolInput struct {
	Command string `json:"command"`
}

type geminiHookOutput struct {
	Decision           string                    `json:"decision"`
	HookSpecificOutput geminiHookSpecificOutput  `json:"hookSpecificOutput,omitempty"`
}

type geminiHookSpecificOutput struct {
	ToolInput geminiToolInput `json:"tool_input"`
}

// RunGeminiHook reads a Gemini CLI BeforeTool hook payload from stdin,
// checks if the command should be wrapped with chop, and outputs
// modified JSON on stdout. Always exits 0.
func RunGeminiHook() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(0)
	}

	output, shouldModify, original := processGeminiHookInput(input)
	if shouldModify {
		var result geminiHookOutput
		if err := json.Unmarshal(output, &result); err == nil {
			auditLog(original, result.HookSpecificOutput.ToolInput.Command)
		}
		fmt.Print(string(output))
	}
	// If not modifying, output nothing (passthrough = allow with original args)
}

// processGeminiHookInput parses the Gemini CLI hook JSON and determines whether
// to wrap the command. Reuses the same shouldWrap/wrapCompound logic as Claude Code.
func processGeminiHookInput(input []byte) ([]byte, bool, string) {
	var h geminiHookInput
	if err := json.Unmarshal(input, &h); err != nil {
		return nil, false, ""
	}

	if h.ToolName != "run_shell_command" {
		return nil, false, ""
	}

	// Global kill switch
	if IsDisabledGlobally() {
		return nil, false, ""
	}

	var ti geminiToolInput
	if err := json.Unmarshal(h.ToolInput, &ti); err != nil {
		return nil, false, ""
	}

	wrapped, shouldModify, original := rewriteCommand(ti.Command)
	if !shouldModify {
		return nil, false, original
	}

	return buildGeminiOutput(wrapped)
}

// buildGeminiOutput constructs the Gemini CLI hook JSON response.
func buildGeminiOutput(wrapped string) ([]byte, bool, string) {
	out := geminiHookOutput{
		Decision: "allow",
		HookSpecificOutput: geminiHookSpecificOutput{
			ToolInput: geminiToolInput{
				Command: wrapped,
			},
		},
	}
	data, err := json.Marshal(out)
	if err != nil {
		return nil, false, ""
	}
	return data, true, wrapped
}
