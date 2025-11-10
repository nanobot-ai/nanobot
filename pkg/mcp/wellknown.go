package mcp

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (h *HTTPServer) protectedMetadata(w http.ResponseWriter, r *http.Request) {
	if h.protectedResourceMetadata == nil {
		// Not protected, return not found
		http.NotFound(w, r)
		return
	}

	protectedResourceMetadata := *h.protectedResourceMetadata
	protectedResourceMetadata.Resource = strings.Replace(r.URL.String(), "/.well-known/oauth-protected-resource", "", 1)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(protectedResourceMetadata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
