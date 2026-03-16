package hooks

import (
	"fmt"
)

// AiderInstructions prints manual integration instructions for Aider.
func AiderInstructions() {
	fmt.Println(`To use openchop with Aider, you can add it to your environment or use it directly.

Aider does not currently have a "hook" system like Claude Code or Gemini CLI,
but you can still benefit from openchop by following these steps:

1. Use openchop manually in Aider's chat:
   /run openchop git status
   /run openchop npm test

2. Add a hint to your .aider.conf.yml or .aider.instructions.md:
   "When running read-only CLI commands, prefer prefixing them with 'openchop'
   to save context tokens (e.g., 'openchop git status', 'openchop docker ps')."

3. For a more transparent experience, you can alias common commands
   within the shell where you run Aider (though Aider may bypass aliases
   depending on how it executes commands).`)
}
