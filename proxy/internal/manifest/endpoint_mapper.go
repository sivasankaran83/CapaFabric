package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sivasankaran83/CapaFabric/shared/models"
)

// BuildRequest constructs an *http.Request that forwards a capability invocation
// to the backing application, applying path / query / body / header argument mapping.
//
// appBaseURL is the application's base URL, e.g. "http://localhost:8081/api".
func BuildRequest(appBaseURL string, cap models.CapabilityMetadata, args map[string]any) (*http.Request, error) {
	rawPath := cap.Endpoint.Path
	query := url.Values{}
	bodyFields := map[string]any{}

	for argName, spec := range cap.Endpoint.Arguments {
		val, ok := args[argName]
		if !ok {
			// Use default if provided
			if spec.Default != nil {
				val = spec.Default
			} else {
				continue
			}
		}

		strVal := fmt.Sprintf("%v", val)

		switch spec.In {
		case models.ArgumentLocationPath:
			rawPath = strings.ReplaceAll(rawPath, "{"+argName+"}", url.PathEscape(strVal))
		case models.ArgumentLocationQuery:
			query.Set(argName, strVal)
		case models.ArgumentLocationBody:
			bodyFields[argName] = val
		case models.ArgumentLocationHeader:
			// headers are applied after the request is built — stored in bodyFields
			// with a sentinel prefix so BuildRequest callers can apply them.
			// Handled separately below via the returned request.
		}
	}

	// Resolve final URL.
	base := strings.TrimRight(appBaseURL, "/")
	if !strings.HasPrefix(rawPath, "/") {
		rawPath = "/" + rawPath
	}
	fullURL := base + rawPath
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	// Build body.
	var bodyReader io.Reader
	method := strings.ToUpper(cap.Endpoint.Method)
	if method != http.MethodGet && method != http.MethodDelete && len(bodyFields) > 0 {
		b, err := json.Marshal(bodyFields)
		if err != nil {
			return nil, fmt.Errorf("marshalling request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("building request to %s: %w", fullURL, err)
	}

	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// Apply header arguments.
	for argName, spec := range cap.Endpoint.Arguments {
		if spec.In != models.ArgumentLocationHeader {
			continue
		}
		val, ok := args[argName]
		if !ok && spec.Default != nil {
			val = spec.Default
		}
		if val != nil {
			req.Header.Set(argName, fmt.Sprintf("%v", val))
		}
	}

	return req, nil
}
