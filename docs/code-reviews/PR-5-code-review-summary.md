# Code Review: PR-5

**Verdict:** ⚠️ CHANGES REQUESTED

| | |
| - | - |
| **Branch** | `feat/auto-update` |
| **Title** | feat: add auto-update - check and apply updates automatically |
| **Author** | @aeanez (Andrés) |
| **Files Changed** | 3 |
| **Lines Changed** | +355 / -0 |
| **Date** | 2026-03-12 |

---

## Summary

Adds background auto-update: after printing command output, chop checks GitHub for a new release every 24h, downloads it silently, and applies it on the next invocation. The design and test coverage are solid. One HIGH issue: `BackgroundCheck` is described as "non-blocking" but actually blocks the process from exiting while the binary downloads (~5–30s on slow connections), delaying shell prompt return. Two MEDIUMs: the `ApplyPendingUpdate` return value is ignored, and tests use a deprecated env-override pattern instead of `t.Setenv`.

---

## Findings Overview

| Severity | In Scope | Out of Scope |
| -------- | -------- | ------------ |
| 🔴 CRITICAL | 0 | 0 |
| 🟠 HIGH | 1 | 0 |
| 🟡 MEDIUM | 3 | 0 |
| 🟢 LOW | 1 | 0 |
| ℹ️ INFO | 1 | 0 |

---

## In Scope Findings

### 🟠 HIGH-001: `BackgroundCheck` blocks process exit during download

**Domains:** [Architecture, Code Quality]
**Location:** `updater/auto_update.go:148`

The PR description and function comment say "non-blocking" and "never delays command output". The output (`fmt.Print`) is indeed printed before the check, but `BackgroundCheck` calls `latestVersion()` (HTTP GET to GitHub API) and then `download()` (downloads a 5–10MB binary) entirely synchronously. The shell prompt does not return until `os.Exit` is called — which happens only after `BackgroundCheck` returns. On a slow connection, this means the user sees their output but waits 10–30 seconds for the prompt. That's surprising and not what the PR claims.

```go
// main.go:166 — download blocks os.Exit
fmt.Print(finalOutput)
trackSilent(...)
updater.BackgroundCheck(version)  // ← HTTP + download, blocks here
os.Exit(exitCode)                 // ← never reached until download finishes
```

**Recommendation:**
Run the download in a goroutine and exit immediately — the goroutine will be killed when the process exits, which is fine for a best-effort background download. Instead, separate the check from the download:

```go
// Option A: goroutine for the whole thing — fire-and-forget, process exits immediately
go updater.BackgroundCheck(version)
os.Exit(exitCode)
```

However, since `os.Exit` kills goroutines, the download would never complete. A cleaner model: do the fast HTTP version check synchronously (< 200ms), and if a newer version is found, spawn a detached subprocess to do the download:

```go
// Option B: version check only (fast), detached subprocess for download
updater.BackgroundCheck(version)  // only does HTTP HEAD, ~200ms
os.Exit(exitCode)
```

Or the simplest fix: only run the HTTP version check inline (tolerable ~200ms delay), write the URL to a marker file, and let a future invocation do the actual download while the user is reading output.

---

### 🟡 MEDIUM-001: `ApplyPendingUpdate` return value ignored — `--post-update-check` never called

**Domains:** [Code Quality]
**Location:** `main.go:26`

```go
// main.go — return value discarded
updater.ApplyPendingUpdate(version)
```

The function comment says _"Returns true if an update was applied (caller should exit)"_ but main never acts on it. More importantly, `chop update` re-execs the new binary with `--post-update-check` after updating, which calls `checkInstallDir()` for post-update validation. The auto-update path skips this entirely — the update is silently applied with no post-update hook.

**Recommendation:**
Either drop the return value from `ApplyPendingUpdate` (it's clearly not used and never will be in the current design), or honor it:

```go
if updater.ApplyPendingUpdate(version) {
    // re-exec so the new binary handles this invocation
    cmd := exec.Command(exe, os.Args[1:]...)
    cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
    _ = cmd.Run()
    os.Exit(0)
}
```

At minimum, remove the misleading comment.

---

### 🟡 MEDIUM-002: Tests use `os.Setenv`/defer instead of `t.Setenv`

**Domains:** [Code Quality]
**Location:** `updater/auto_update_test.go:14`

```go
// Current pattern (7 occurrences)
origHome := os.Getenv("HOME")
tmpDir := t.TempDir()
os.Setenv("HOME", tmpDir)
defer os.Setenv("HOME", origHome)
```

The rest of the codebase uses `t.Setenv` (e.g. `tracking/tracker_test.go`), which automatically restores the env var after the test — including on panic or failure. The manual pattern can leave env vars dirty if the test panics.

**Recommendation:**
```go
t.Setenv("HOME", t.TempDir())
```

---

### 🟡 MEDIUM-003: `touchLastCheck()` called before verifying network is reachable

**Domains:** [Code Quality]
**Location:** `updater/auto_update.go:146`

```go
touchLastCheck()  // ← timestamp updated here

latest, err := latestVersion()
if err != nil {   // ← network error — but check already marked as done
    return
}
```

If GitHub is unreachable (network down, API rate-limited), the check is still marked as completed, suppressing retries for 24h. This means a user on a flaky connection could go days without receiving an update notification.

**Recommendation:**
Move `touchLastCheck()` after a successful version fetch, or only touch it when `latest == currentVersion` (already up to date):

```go
latest, err := latestVersion()
if err != nil {
    return  // don't touch — retry next time
}
touchLastCheck()  // only mark done on successful check
```

---

### 🟢 LOW-001: `dataDir()` duplicated in `updater` and `cleanup` packages

**Domains:** [Code Quality]
**Location:** `updater/auto_update.go:15`, `cleanup/cleanup.go:59`

Both packages independently define `dataDir()` returning `~/.local/share/chop`. If the path ever changes, it needs updating in two places.

**Recommendation:**
Extract to a shared `internal/paths` package, or expose it from one package and import it in the other. Low priority — it's a one-liner, but the duplication is unnecessary.

---

### ℹ️ INFO-001: Good test coverage for edge cases

**Location:** `updater/auto_update_test.go`

10 tests covering: never-checked, recently-checked, stale check, dev version guard, missing marker, invalid marker format, missing binary, `touchLastCheck`, and `replaceBinary`. The stale-check test correctly backdates the file using `os.Chtimes` rather than sleeping. Well done.

---

## Action Items

### Must Fix (blocks merge)

- [ ] **HIGH-001** — Make the download truly non-blocking so `os.Exit` is not delayed

### Should Fix

- [ ] **MEDIUM-001** — Either honor the `ApplyPendingUpdate` return value or remove it and the misleading comment
- [ ] **MEDIUM-002** — Replace `os.Setenv`/defer with `t.Setenv` across all 7 test occurrences
- [ ] **MEDIUM-003** — Move `touchLastCheck()` after a successful version check

### Consider

- [ ] **LOW-001** — Extract `dataDir()` to a shared location to avoid duplication

---

## Files Reviewed

| File | Findings |
| ---- | -------- |
| `updater/auto_update.go` | HIGH-001, MEDIUM-003, LOW-001 |
| `main.go` | HIGH-001, MEDIUM-001 |
| `updater/auto_update_test.go` | MEDIUM-002, INFO-001 |

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
