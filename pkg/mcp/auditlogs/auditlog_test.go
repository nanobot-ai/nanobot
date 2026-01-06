package auditlogs

import "testing"

func TestRedactAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "standard API key",
			apiKey:   "ok1-123-456-secretABC",
			expected: "ok1-123-45",
		},
		{
			name:     "longer API key",
			apiKey:   "ok1-12345-67890-secret-with-dashes-and-more",
			expected: "ok1-12345-67890-secre",
		},
		{
			name:     "short key",
			apiKey:   "ab",
			expected: "a",
		},
		{
			name:     "single character",
			apiKey:   "a",
			expected: "a",
		},
		{
			name:     "empty string",
			apiKey:   "",
			expected: "",
		},
		{
			name:     "four characters",
			apiKey:   "ok1-",
			expected: "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactAPIKey(tt.apiKey)
			if result != tt.expected {
				t.Errorf("RedactAPIKey(%q) = %q, want %q", tt.apiKey, result, tt.expected)
			}
		})
	}
}
