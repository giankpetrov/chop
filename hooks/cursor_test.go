package hooks

import (
	"encoding/json"
	"testing"
)

func TestCursorHookWrapsBash(t *testing.T) {
	input := `{
		"session_id": "test",
		"cwd": "/tmp",
		"hook_event_name": "preToolUse",
		"tool_name": "bash",
		"tool_input": {
			"command": "git status"
		}
	}`

	output, shouldModify, original := processCursorHookInput([]byte(input))
	if !shouldModify {
		t.Fatal("expected shouldModify to be true")
	}
	if original != "git status" {
		t.Errorf("expected original 'git status', got %q", original)
	}

	var out cursorHookOutput
	if err := json.Unmarshal(output, &out); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if out.HookSpecificOutput.UpdatedInput.Command != "chop git status" {
		t.Errorf("expected 'chop git status', got %q", out.HookSpecificOutput.UpdatedInput.Command)
	}
}

func TestCursorHookIgnoresNonBashTool(t *testing.T) {
	input := `{
		"session_id": "test",
		"cwd": "/tmp",
		"hook_event_name": "preToolUse",
		"tool_name": "FileRead",
		"tool_input": {
			"path": "test.txt"
		}
	}`

	_, shouldModify, _ := processCursorHookInput([]byte(input))
	if shouldModify {
		t.Error("expected shouldModify to be false for non-bash tool")
	}
}
