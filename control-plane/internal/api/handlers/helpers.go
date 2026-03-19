package handlers

import (
	"encoding/json"
	"net/http"
)

// writeJSON serialises body as JSON with the given status.
// Used only for successful responses — errors are returned and handled by Wrap.
func writeJSON(w http.ResponseWriter, status int, body any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(body)
}
