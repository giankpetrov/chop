# Contributing to chop

Thanks for your interest in contributing! chop has a modular design that makes adding new filters straightforward.

## Keeping your fork up to date

PRs submitted from a stale branch conflict with recent changes and end up applied manually, showing as "Closed" instead of "Merged". Follow these steps every time before starting new work.

### Option A - upstream remote (recommended)

**One-time setup:**

```bash
git remote add upstream https://github.com/AgusRdz/chop.git
```

**Before every PR:**

```bash
# 1. Fetch latest from upstream
git fetch upstream

# 2. Rebase your local main
git rebase upstream/main

# 3. Push to your fork
git push origin main

# 4. Create your feature branch from the updated main
git checkout -b feat/my-feature
```

### Option B - GitHub UI sync

If you use the **"Sync fork"** button on GitHub, you still need to pull locally before branching — otherwise your local copy is still on the old commit:

```bash
# 1. Click "Sync fork" on your fork's GitHub page

# 2. Pull the sync down locally
git pull origin main

# 3. Create your feature branch
git checkout -b feat/my-feature
```

## Development setup

chop builds inside Docker — no local Go installation required.

```bash
# Run tests
make test

# Build binary
make build

# Run a single test
docker compose run --rm dev go test ./filters/ -run TestFilterPing -v

# Coverage report
make coverage
```

## Adding a new filter

This is the most common contribution. Each filter is a self-contained file pair: implementation + test.

### 1. Create the filter

Create `filters/<command>.go` (or `filters/<command>_<subcommand>.go` for subcommand-specific filters):

```go
package filters

import "strings"

func filterMyCmd(raw string) (string, error) {
    trimmed := strings.TrimSpace(raw)
    if trimmed == "" {
        return "", nil
    }
    if !looksLikeMyCmd(trimmed) {
        return raw, nil
    }

    // Your compression logic here.
    // Strip noise, keep signal.
    var out []string
    // ...

    result := strings.Join(out, "\n")
    return outputSanityCheck(raw, result), nil
}
```

Key patterns:
- Return `("", nil)` for empty input
- Return `(raw, nil)` if the output doesn't look like expected format (sanity guard)
- Always call `outputSanityCheck(raw, result)` before returning — it prevents filters from accidentally expanding output
- Never return an error for bad input — just return `raw` unchanged

### 2. Add a sanity guard

Add a `looksLike*` function to `filters/sanity_guards.go`:

```go
func looksLikeMyCmd(s string) bool {
    return strings.Contains(s, "SOME_MARKER") ||
        strings.Contains(s, "ANOTHER_MARKER")
}
```

Use simple string checks (`Contains`, `HasPrefix`) — no regex. These run on every invocation and must be fast.

### 3. Wire it into the router

Add your filter to the `get()` switch in `filters/filters.go`:

```go
case "mycmd":
    return filterMyCmd
```

For commands with subcommands, add a routing function:

```go
case "mycmd":
    return getMyCmdFilter(args)
```

### 4. Write tests

Create `filters/<command>_test.go` with realistic fixture data:

```go
package filters

import (
    "strings"
    "testing"
)

var myCmdFixture = `... paste real command output here ...`

func TestFilterMyCmd(t *testing.T) {
    got, err := filterMyCmd(myCmdFixture)
    if err != nil {
        t.Fatal(err)
    }

    // Verify important content is preserved
    if !strings.Contains(got, "important thing") {
        t.Errorf("expected important thing in output: %s", got)
    }

    // Verify noise is stripped
    if strings.Contains(got, "noisy line") {
        t.Errorf("noisy line should be stripped: %s", got)
    }

    // Verify compression ratio
    rawTokens := countTokens(myCmdFixture)
    filteredTokens := countTokens(got)
    savings := 100.0 - (float64(filteredTokens)/float64(rawTokens))*100.0
    t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
    t.Logf("output:\n%s", got)
}

func TestFilterMyCmdEmpty(t *testing.T) {
    got, err := filterMyCmd("")
    if err != nil {
        t.Fatal(err)
    }
    if got != "" {
        t.Errorf("expected empty output, got: %q", got)
    }
}

func TestMyCmdRouted(t *testing.T) {
    f := get("mycmd", []string{})
    if f == nil {
        t.Fatal("expected filter for mycmd, got nil")
    }
}
```

Use `chop capture <command> [args...]` to grab real output for test fixtures.

### 5. Aliasing compatible commands

If a command produces identical output to an existing one (e.g., `podman` → `docker`, `tofu` → `terraform`), just add it to the same `case` in the router:

```go
case "docker", "podman":
    return getDockerFilter(args)
```

## Project structure

```
filters/          Filter implementations + tests (one file per command)
  filters.go      Router — maps commands to filter functions
  sanity_guards.go  Format recognition functions
  auto_detect.go  Heuristic fallback for unknown commands
config/           Configuration loading (global + local + custom filters)
hooks/            Claude Code PreToolUse hook integration
tracking/         SQLite-based token savings analytics
updater/          Self-update mechanism
cleanup/          Uninstall and reset logic
main.go           CLI entry point and subcommand handlers
```

## Commit conventions

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add kubectl rollout filter
fix: handle empty output in docker stats
perf: optimize string concatenation in log filter
test: add tests for custom filter loading
docs: update supported commands table
chore: update dependencies
```

These feed directly into the changelog via [git-cliff](https://git-cliff.org/).

## Guidelines

- **One filter per file** — keeps things modular and easy to review
- **Match existing patterns** — look at `ping.go` or `git_push.go` for clean examples
- **Test with real output** — use `chop capture` to grab fixtures from actual commands
- **Don't break on unexpected input** — return `raw` unchanged rather than erroring
- **Keep dependencies minimal** — stdlib only in filters (the only external dep is SQLite for tracking)