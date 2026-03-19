package manifest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ExtractResult reads an HTTP response and extracts the result value according
// to the capability's response.from directive.
//
// response.from values:
//   - "body"           → return the full decoded JSON body
//   - "body.field"     → return body["field"] (supports dot-separated paths)
//   - "header.X-Name"  → return the named response header value
func ExtractResult(resp *http.Response, fromSpec string) (any, error) {
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Header extraction.
	if strings.HasPrefix(fromSpec, "header.") {
		headerName := strings.TrimPrefix(fromSpec, "header.")
		return resp.Header.Get(headerName), nil
	}

	// Body extraction — parse JSON.
	if len(raw) == 0 {
		return nil, nil
	}

	var parsed any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		// Return raw string if not valid JSON.
		return string(raw), nil
	}

	if fromSpec == "body" || fromSpec == "" {
		return parsed, nil
	}

	// Dot-path traversal: "body.receipts" → parsed["receipts"]
	path := strings.TrimPrefix(fromSpec, "body.")
	return traversePath(parsed, strings.Split(path, ".")), nil
}

// traversePath navigates nested maps/slices using a dot-split path.
func traversePath(v any, parts []string) any {
	if len(parts) == 0 || v == nil {
		return v
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	child, ok := m[parts[0]]
	if !ok {
		return nil
	}
	return traversePath(child, parts[1:])
}
