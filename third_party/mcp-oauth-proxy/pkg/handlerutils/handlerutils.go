package handlerutils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func JSON(w http.ResponseWriter, statusCode int, obj any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if obj != nil {
		if err := json.NewEncoder(w).Encode(obj); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
			// Write error response if encoding fails
			errText, _ := json.Marshal(map[string]string{
				"error":             "internal_server_error",
				"error_description": "Failed to encode JSON response",
				"error_detail":      err.Error(),
			})
			_, _ = w.Write(errText)
		}
	}
}

// GetClientIP extracts the client IP from the request using the X-Forwarded-For,
// X-Real-IP and RemoteAddr headers.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Get the first IP in the comma-separated list
		ifs := strings.Split(xff, ",")
		return strings.TrimSpace(ifs[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}

// GetBaseURL returns the URL of the request without the path and
// infers the scheme (http or https)
func GetBaseURL(r *http.Request) string {
	if url := r.Header.Get("X-Mcp-Oauth-Proxy-URL"); url != "" {
		return url
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}
