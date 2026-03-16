package hooks

import "fmt"

// PrintAiderInstructions prints instructions on how to use chop with Aider.
func PrintAiderInstructions() {
	fmt.Println(`## Chop (Token Optimizer) for Aider

Aider does not yet support automatic hooks like Claude Code or Gemini CLI.
To use chop with Aider, you can instruct the agent to use it in your
conventions file (e.g. CONVENTIONS.md or .aider.conf.yml).

### Option 1: Add to your CONVENTIONS.md (Recommended)

Add the following to your project's conventions:

"""
## Chop (Token Optimizer)

'chop' is installed on this system. It compresses CLI output to reduce token consumption.

When running CLI commands via shell, prefix with 'chop' for read-only commands:
- 'chop git status', 'chop git log -10', 'chop git diff'
- 'chop docker ps', 'chop npm test', 'chop dotnet build'
- 'chop curl <url>' (auto-compresses JSON responses)

Do NOT use chop for: interactive commands, pipes, redirects, or write commands
(git commit, git push, npm init, docker run).
"""

### Option 2: Run with a prefix manually

You can also use chop manually when asking Aider to run commands:
> /run chop npm test`)
}
