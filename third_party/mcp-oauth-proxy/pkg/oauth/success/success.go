package success

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/obot-platform/mcp-oauth-proxy/pkg/handlerutils"
)

// Handler handles the MCP UI authentication success page
type Handler struct{}

// NewHandler creates a new success page handler
func NewHandler() http.Handler {
	return &Handler{}
}

// successPageTemplate is the HTML template for the success page
const successPageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authentication Success</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            margin: 0;
            padding: 0;
            height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            background: white;
            color: #333;
        }
        .content {
            text-align: center;
            line-height: 1.6;
        }
        .countdown {
            font-weight: bold;
        }
        a {
            color: #007bff;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
    </style>
    <script>
        let countdown = 5;
        let redirectUrl = "{{.RedirectURL}}";
        
        function updateCountdown() {
            document.getElementById('countdown').textContent = countdown;
            countdown--;
            
            if (countdown < 0) {
                window.location.href = redirectUrl;
            }
        }
        
        // Start countdown when page loads
        window.onload = function() {
            updateCountdown();
            setInterval(updateCountdown, 1000);
        };
    </script>
</head>
<body>
    <div class="content">
        <p>Authentication successful.</p>
        <p>Redirecting in <span id="countdown" class="countdown">5</span> seconds...</p>
        <p><a href="{{.RedirectURL}}">Continue manually</a></p>
    </div>
</body>
</html>`

// ServeHTTP handles the success page request
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the redirect path from query parameter
	rdParam := r.URL.Query().Get("rd")
	if rdParam == "" {
		rdParam = "/" // Default to root if no redirect path provided
	}

	// Parse and clean the redirect path
	redirectURL, err := url.Parse(rdParam)
	if err != nil || redirectURL.IsAbs() {
		// If invalid or absolute URL, default to root
		rdParam = "/"
	} else {
		// Remove mcp-ui-code parameter from query string
		query := redirectURL.Query()
		query.Del("mcp-ui-code")
		query.Del("mcp-ui-refresh-code") // Also remove refresh code if present

		// Rebuild the path properly
		if len(query) == 0 {
			// If no query parameters left, just use the path
			rdParam = redirectURL.Path
		} else {
			// If there are still query parameters, include them
			redirectURL.RawQuery = query.Encode()
			rdParam = redirectURL.String()
		}
	}

	// Ensure redirect path starts with /
	if !strings.HasPrefix(rdParam, "/") {
		rdParam = "/" + rdParam
	}

	// Build full redirect URL
	baseURL := handlerutils.GetBaseURL(r)
	fullRedirectURL := baseURL + rdParam

	// Prepare template data
	templateData := struct {
		RedirectURL string
		DisplayPath string
	}{
		RedirectURL: fullRedirectURL,
		DisplayPath: rdParam, // This is already cleaned above
	}

	// Parse and execute template
	tmpl, err := template.New("success").Parse(successPageTemplate)
	if err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}

	// Set content type and render the page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := tmpl.Execute(w, templateData); err != nil {
		_, _ = fmt.Fprintf(w, "Error rendering page: %v", err)
	}
}
