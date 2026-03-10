package types

// APIKeyAuthRequest is the request body sent to the API key authentication webhook.
type APIKeyAuthRequest struct {
	MCPID string `json:"mcpId,omitempty"`
}

// APIKeyAuthResponse is the response from the API key authentication webhook.
type APIKeyAuthResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`

	Subject           string `json:"sub,omitempty"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Email             string `json:"email,omitempty"`
}
