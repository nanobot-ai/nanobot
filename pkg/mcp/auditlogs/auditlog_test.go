package auditlogs

import "testing"

func TestRedactAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "valid API key",
			apiKey:   "ok1-123-456-secretABC",
			expected: "ok1-123-456-*****",
		},
		{
			name:     "valid API key with secret containing dashes",
			apiKey:   "ok1-12345-67890-secret-with-dashes-and-more",
			expected: "ok1-12345-67890-*****",
		},
		{
			name:     "invalid prefix",
			apiKey:   "xyz-user123-key456-secret",
			expected: "",
		},
		{
			name:     "too few parts",
			apiKey:   "ok1-user123-key456",
			expected: "",
		},
		{
			name:     "empty string",
			apiKey:   "",
			expected: "",
		},
		{
			name:     "only prefix",
			apiKey:   "ok1-",
			expected: "",
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
