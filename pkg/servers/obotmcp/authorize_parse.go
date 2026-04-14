package obotmcp

import (
	"net/url"
	"strings"
)

// ResourceURLFromAuthorizeURL returns the RFC 8707-style resource URL from an OAuth
// authorize URL's query parameter (often the MCP HTTPS endpoint).
func ResourceURLFromAuthorizeURL(authorizeURL string) string {
	u, err := url.Parse(strings.TrimSpace(authorizeURL))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(u.Query().Get("resource"))
}
