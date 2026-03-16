package hooks

import (
	"encoding/json"
	"io"
	"os"
)

// claudeHookInput represents the JSON payload received from Claude Code's PreToolUse hook.
type claudeHookInput struct {
	SessionID     string          `json:"session_id"`
	Cwd           string          `json:"cwd"`
	HookEventName string          `json:"hook_event_name"`
	ToolName      string          `json:"tool_name"`
	ToolInput     json.RawMessage `json:"tool_input"`
}

type toolInput struct {
	Command string `json:"command"`
}

type hookOutput struct {
	HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
}

type hookSpecificOutput struct {
	HookEventName      string    `json:"hookEventName"`
	PermissionDecision string    `json:"permissionDecision"`
	UpdatedInput       toolInput `json:"updatedInput"`
}

type claudeHookOutput = hookOutput
type claudeHookSpecificOutput = hookSpecificOutput

// geminiHookInput represents the JSON payload received from Gemini CLI's BeforeTool hook.
type geminiHookInput struct {
	SessionID string          `json:"session_id"`
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
}

type geminiToolInput struct {
	Command string `json:"command"`
}

type geminiHookOutput struct {
	Decision           string                   `json:"decision,omitempty"`
	HookSpecificOutput geminiHookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

type geminiHookSpecificOutput struct {
	ToolInput geminiToolInput `json:"tool_input"`
}

// RunHook attempts to detect the agent and run the appropriate hook handler.
func RunHook() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return
	}

	// Try Claude format
	var c claudeHookInput
	if err := json.Unmarshal(input, &c); err == nil && c.HookEventName != "" {
		processClaude(input)
		return
	}

	// Try Gemini format
	var g geminiHookInput
	if err := json.Unmarshal(input, &g); err == nil && g.ToolName != "" {
		processGemini(input)
		return
	}
}

// processHookInput is used by tests to verify Claude Code hook logic.
func processHookInput(input []byte) ([]byte, bool, string) {
	var h claudeHookInput
	if err := json.Unmarshal(input, &h); err != nil {
		return nil, false, ""
	}

	if h.ToolName != "Bash" {
		return nil, false, ""
	}

	if IsDisabledGlobally() {
		return nil, false, ""
	}

	var ti toolInput
	if err := json.Unmarshal(h.ToolInput, &ti); err != nil {
		return nil, false, ""
	}

	wrapped, modified := WrapCommand(ti.Command)
	if !modified {
		return nil, false, ti.Command
	}

	out := claudeHookOutput{
		HookSpecificOutput: claudeHookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
			UpdatedInput: toolInput{
				Command: wrapped,
			},
		},
	}
	data, err := json.Marshal(out)
	if err != nil {
		return nil, false, ti.Command
	}
	return data, true, ti.Command
}

func processClaude(input []byte) {
	output, modified, original := processHookInput(input)
	if modified {
		var result claudeHookOutput
		if err := json.Unmarshal(output, &result); err == nil {
			auditLog(original, result.HookSpecificOutput.UpdatedInput.Command)
		}
		os.Stdout.Write(output)
	}
}

func processGemini(input []byte) {
	var h geminiHookInput
	_ = json.Unmarshal(input, &h)

	if h.ToolName != "run_shell_command" {
		return
	}

	if IsDisabledGlobally() {
		return
	}

	var ti geminiToolInput
	if err := json.Unmarshal(h.ToolInput, &ti); err != nil {
		return
	}

	wrapped, modified := WrapCommand(ti.Command)
	if modified {
		out := geminiHookOutput{
			Decision: "allow",
			HookSpecificOutput: geminiHookSpecificOutput{
				ToolInput: geminiToolInput{
					Command: wrapped,
				},
			},
		}
		data, err := json.Marshal(out)
		if err == nil {
			auditLog(ti.Command, wrapped)
			os.Stdout.Write(data)
		}
	}
}
