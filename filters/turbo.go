package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// "• Packages in scope: api, web, docs"
	reTurboScope = regexp.MustCompile(`^[•·]\s*Packages in scope:\s*(.+)$`)
	// "api:build: cache miss, executing abc123"  or  "api:build: cache hit, replaying output abc456"
	reTurboTaskCache = regexp.MustCompile(`^(\S+:\S+): cache (hit|miss),`)
	// " Tasks:    2 successful, 1 failed"
	reTurboSummary = regexp.MustCompile(`Tasks:\s+(.+)`)
	// " Time:     12.345s"  or  ">>> FULL TURBO" timing
	reTurboTime = regexp.MustCompile(`Time:\s+(\S+)`)
	// " Cached:   1 cached, 2 not cached"
	reTurboCached = regexp.MustCompile(`Cached:\s+(.+)`)
	// Task output lines like "api:build: ..."
	reTurboTaskLine = regexp.MustCompile(`^(\S+:\S+): (.+)$`)
)

func filterTurbo(raw string) (string, error) {
	trimmed := strings.TrimSpace(stripAnsi(raw))
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikeTurboOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")

	type taskInfo struct {
		name      string
		cacheHit  bool
		cacheMiss bool
		success   bool
		failed    bool
		errors    []string
	}

	tasks := make(map[string]*taskInfo)
	var taskOrder []string

	var scopeLine string
	var summaryLine string
	var timeLine string
	var cachedLine string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Packages in scope
		if m := reTurboScope.FindStringSubmatch(line); m != nil {
			scopeLine = "Packages in scope: " + m[1]
			continue
		}

		// Summary
		if m := reTurboSummary.FindStringSubmatch(line); m != nil {
			summaryLine = "Tasks: " + m[1]
			continue
		}

		// Time
		if m := reTurboTime.FindStringSubmatch(line); m != nil {
			timeLine = "Time: " + m[1]
			continue
		}

		// Cached line
		if m := reTurboCached.FindStringSubmatch(line); m != nil {
			cachedLine = "Cached: " + m[1]
			_ = cachedLine // used in compact path
			continue
		}

		// Cache hit/miss marker
		if m := reTurboTaskCache.FindStringSubmatch(line); m != nil {
			name := m[1]
			if _, exists := tasks[name]; !exists {
				tasks[name] = &taskInfo{name: name}
				taskOrder = append(taskOrder, name)
			}
			if m[2] == "hit" {
				tasks[name].cacheHit = true
			} else {
				tasks[name].cacheMiss = true
			}
			continue
		}

		// Task output line
		if m := reTurboTaskLine.FindStringSubmatch(line); m != nil {
			name, content := m[1], m[2]
			if _, exists := tasks[name]; !exists {
				tasks[name] = &taskInfo{name: name}
				taskOrder = append(taskOrder, name)
			}
			t := tasks[name]

			// Detect completion signals
			lc := strings.ToLower(content)
			if strings.Contains(lc, "compiled successfully") ||
				strings.Contains(content, "✓") ||
				strings.Contains(lc, "build success") {
				t.success = true
			}
			if strings.Contains(lc, "error") || strings.Contains(lc, "failed") {
				t.failed = true
			}

			// Capture error lines
			if strings.Contains(content, "ERROR") || strings.HasPrefix(content, "Error:") {
				t.errors = append(t.errors, content)
			}
			continue
		}
	}

	// Infer success for tasks without an explicit success marker but no failure
	for _, t := range tasks {
		if !t.failed && !t.success {
			t.success = true
		}
	}

	// Compact path: all successful and all cached
	if summaryLine != "" && !strings.Contains(summaryLine, "failed") && cachedLine != "" && !strings.Contains(cachedLine, "0 cached") {
		// Check if every known task was a cache hit
		allCached := true
		for _, t := range tasks {
			if !t.cacheHit {
				allCached = false
				break
			}
		}
		if allCached && len(tasks) > 0 && timeLine != "" {
			result := fmt.Sprintf("%d tasks, all cached, %s", len(tasks), strings.TrimPrefix(timeLine, "Time: "))
			return outputSanityCheck(trimmed, result), nil
		}
	}

	var out []string

	if scopeLine != "" {
		out = append(out, scopeLine)
	}

	for _, name := range taskOrder {
		t := tasks[name]

		cacheLabel := ""
		if t.cacheHit {
			cacheLabel = "cache hit"
		} else if t.cacheMiss {
			cacheLabel = "cache miss"
		}

		statusLabel := ""
		if t.failed {
			statusLabel = "FAILED"
		} else {
			statusLabel = "success"
		}

		var taskLine string
		if cacheLabel != "" {
			taskLine = fmt.Sprintf("%s (%s) → %s", name, cacheLabel, statusLabel)
		} else {
			taskLine = fmt.Sprintf("%s → %s", name, statusLabel)
		}
		out = append(out, taskLine)

		for _, errLine := range t.errors {
			out = append(out, "  "+errLine)
		}
	}

	if summaryLine != "" {
		out = append(out, summaryLine)
	}
	if timeLine != "" {
		out = append(out, timeLine)
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(trimmed, result), nil
}
