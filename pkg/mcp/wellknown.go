package mcp

import (
	"encoding/json"
	"net/http"
)

func (h *HTTPServer) protectedMetadata(w http.ResponseWriter, r *http.Request) {
	if h.protectedResourceMetadata == nil {
		// Not protected, return not found
		http.NotFound(w, r)
		return
	}

	protectedResourceMetadata := *h.protectedResourceMetadata

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(protectedResourceMetadata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
