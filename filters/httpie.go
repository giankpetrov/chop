package filters

import (
	"fmt"
	"strings"
)

// filterHttpie filters httpie (http command) output.
// httpie format: colored headers followed by blank line and body.
func filterHttpie(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	// Check for httpie errors
	if isHttpieError(raw) {
		return raw, nil
	}

	// httpie shows headers like:
	//   HTTP/1.1 200 OK
	//   Content-Type: application/json
	//   ...
	//   <blank line>
	//   {body}
	body, statusLine := extractHTTPBody(raw)

	trimmedBody := strings.TrimSpace(body)

	// JSON response
	if isJSON(trimmedBody) {
		compressed, err := compressJSON(trimmedBody)
		if err != nil {
			return raw, nil
		}
		if statusLine != "" {
			return statusLine + "\n" + compressed, nil
		}
		return compressed, nil
	}

	// HTML response
	if isHTML(trimmedBody) {
		size := len(trimmedBody)
		summary := fmt.Sprintf("HTML response (%d bytes)", size)
		if statusLine != "" {
			return statusLine + "\n" + summary, nil
		}
		return summary, nil
	}

	// Binary
	if isBinary(trimmedBody) {
		size := len(trimmedBody)
		summary := fmt.Sprintf("binary response (%d bytes)", size)
		if statusLine != "" {
			return statusLine + "\n" + summary, nil
		}
		return summary, nil
	}

	// Plain text
	lines := strings.Split(trimmedBody, "\n")
	if len(lines) > 100 {
		truncated := strings.Join(lines[:100], "\n")
		truncated += fmt.Sprintf("\n... (%d more lines)", len(lines)-100)
		if statusLine != "" {
			return statusLine + "\n" + truncated, nil
		}
		return truncated, nil
	}

	var result string
	if statusLine != "" {
		result = statusLine + "\n" + trimmedBody
	} else {
		result = trimmedBody
	}
	return outputSanityCheck(raw, redactHeaders(result)), nil
}

// isHttpieError detects httpie error messages.
func isHttpieError(raw string) bool {
	lower := strings.ToLower(raw)
	httpieErrors := []string{
		"http: error:",
		"connectionerror:",
		"connection refused",
		"connection timed out",
		"could not resolve",
	}
	for _, msg := range httpieErrors {
		if strings.Contains(lower, msg) {
			return true
		}
	}
	return false
}
