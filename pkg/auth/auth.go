package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/oauth/validate"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/proxy"
	proxytypes "github.com/obot-platform/mcp-oauth-proxy/pkg/types"
	"golang.org/x/oauth2"
)

type Auth struct {
	OAuthClientID     string   `usage:"OAuth client ID"`
	OAuthClientSecret string   `usage:"OAuth client secret"`
	OAuthAuthorizeURL string   `usage:"OAuth authorize URL for third-party OAuth provider"`
	OAuthTokenURL     string   `usage:"OAuth token URL for third-party OAuth provider"`
	OAuthScopes       []string `usage:"OAuth scopes to request during OAuth flow"`
	TrustedIssuer     string   `usage:"Trusted issuer for JWT tokens"`
	JWKS              string   `usage:"JWKS public key for JWT tokens"`
	TrustedAudiences  []string `usage:"Trusted audiences for JWT tokens"`
	EncryptionKey     string   `usage:"Encryption key for storing sensitive data"`
}

func Wrap(ctx context.Context, env map[string]string, auth Auth, dsn, address, healthzPath string, runningUI bool, oauthCallbackHandler mcp.CallbackServer, next http.Handler) (http.Handler, error) {
	if auth.OAuthClientID == "" {
		return next, nil
	}

	result, err := setupContext(ctx, auth, address, runningUI, oauthCallbackHandler, next)
	if err != nil {
		return nil, err
	}

	if auth.OAuthClientID != "" {
		if auth.OAuthClientSecret == "" {
			return nil, fmt.Errorf("oauthClientSecret is required")
		}
		if auth.OAuthAuthorizeURL == "" {
			return nil, fmt.Errorf("oauthAuthorizeURL is required")
		}

		proxy, err := mcpProxy(auth, dsn, result)
		if err != nil {
			return nil, fmt.Errorf("failed to create oauth proxy: %w", err)
		}

		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if shouldIgnoreAuth(req, healthzPath) {
				next.ServeHTTP(rw, req)
				return
			} else if strings.HasPrefix(req.URL.Path, "/.well-known/") {
				proxy.ServeHTTP(rw, req)
			} else {
				result.ServeHTTP(rw, req)
			}
		}), nil
	}

	return result, nil
}

func userFromHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var user mcp.User
		keys := map[string]any{}
		_ = mcp.JSONCoerce(user, &keys)
		for key := range keys {
			v := req.Header.Get("X-Forwarded-" + strings.ReplaceAll(key, "_", "_"))
			if key == "email_verified" {
				keys[key] = v == "true"
			} else {
				keys[key] = v
			}
		}
		_ = mcp.JSONCoerce(keys, &user)

		if user.ID == "" {
			user.ID = user.Sub
		}
		if user.ID == "" {
			user.ID = user.Login
		}

		nctx := types.NanobotContext(req.Context())
		nctx.User = user
		next.ServeHTTP(rw, req.WithContext(types.WithNanobotContext(mcp.WithUser(req.Context(), user), nctx)))
	})
}

func userFromJWT(ctx context.Context, auth Auth, address string, runningUI bool, oauthCallbackHandler mcp.CallbackServer, next http.Handler) (http.Handler, error) {
	var (
		k   keyfunc.Keyfunc
		err error
	)
	if auth.TrustedIssuer != "" {
		if auth.JWKS != "" {
			var b []byte
			b, err = base64.StdEncoding.DecodeString(auth.JWKS)
			if err != nil {
				return nil, fmt.Errorf("failed to decode JWKS: %w", err)
			}

			k, err = keyfunc.NewJWKSetJSON(b)
		} else {
			// Find the jwks URL for this issuer
			jwksURL, err := findJWKSURL(ctx, auth.TrustedIssuer)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch JWKS URL: %w", err)
			}

			k, err = keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create client JWK set: %w", err)
		}
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		originalToken := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
		if originalToken == "" {
			// Don't care about the error here, it will only be a not-found error.
			cookie, _ := req.Cookie("nanobot-token")
			if cookie != nil {
				originalToken = cookie.Value
			}
		}
		token, err := jwt.Parse(
			originalToken,
			k.Keyfunc,
			jwt.WithAudience(auth.TrustedAudiences...),
			jwt.WithIssuer(auth.TrustedIssuer),
		)
		if err != nil {
			log.Infof(ctx, "Failed to parse JWT token for %s: %v", req.URL, err)
			if runningUI && strings.Contains(strings.ToLower(req.UserAgent()), "mozilla") {
				conf := &oauth2.Config{
					ClientID:     auth.OAuthClientID,
					ClientSecret: auth.OAuthClientSecret,
					RedirectURL:  fmt.Sprintf("http://%s/oauth/callback", address),
					Endpoint: oauth2.Endpoint{
						AuthURL:   auth.OAuthAuthorizeURL,
						TokenURL:  auth.OAuthTokenURL,
						AuthStyle: oauth2.AuthStyleInHeader,
					},
					Scopes: auth.OAuthScopes,
				}

				verifier := oauth2.GenerateVerifier()

				state, err := oauthCallbackHandler.NewStateWithRedirect(ctx, conf, verifier, req.URL.Path)
				if err != nil {
					http.Error(rw, fmt.Sprintf("error: %s", err.Error()), http.StatusInternalServerError)
					return
				}

				// Redirect user to consent page to ask for permission
				// for the scopes specified above.
				http.Redirect(rw, req, conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("resource", auth.TrustedAudiences[0])), http.StatusFound)
				return
			}

			host := req.Header.Get("X-Forwarded-Host")
			if host == "" {
				host = req.Host
			}
			scheme := req.Header.Get("X-Forwarded-Proto")
			if scheme == "" {
				if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
					scheme = "http"
				} else {
					scheme = "https"
				}
			}
			resourceMetadata := strings.TrimSuffix(fmt.Sprintf("%s://%s/.well-known/oauth-protected-resource/%s", scheme, host, strings.TrimPrefix(req.URL.Path, "/")), "/")

			rw.Header().Set("WWW-Authenticate",
				strings.TrimSuffix(
					fmt.Sprintf(`Bearer error="invalid_request", error_description="Invalid access token", resource_metadata="%s"`,
						resourceMetadata,
					),
					"/"),
			)
			http.Error(rw, `{"http_error": "unauthorized"}`, http.StatusUnauthorized)
			return
		}

		delete(req.Header, "Authorization")

		claims := token.Claims.(jwt.MapClaims)
		nctx := types.NanobotContext(req.Context())
		nctx.User = mcp.User{
			ID:            fmt.Sprintf("%v", claims["user_id"]),
			Sub:           fmt.Sprintf("%v", claims["sub"]),
			Login:         fmt.Sprintf("%v", claims["login"]),
			Email:         fmt.Sprintf("%v", claims["email"]),
			EmailVerified: claims["email_verified"] == "true",
			Name:          fmt.Sprintf("%v", claims["name"]),
			GivenName:     fmt.Sprintf("%v", claims["given_name"]),
			FamilyName:    fmt.Sprintf("%v", claims["family_name"]),
			Picture:       fmt.Sprintf("%v", claims["picture"]),
			Locale:        fmt.Sprintf("%v", claims["locale"]),
		}
		next.ServeHTTP(rw, req.WithContext(types.WithNanobotContext(mcp.WithUser(mcp.WithToken(req.Context(), token.Raw), nctx.User), nctx)))
	}), nil
}

func setupContext(ctx context.Context, auth Auth, address string, runningUI bool, oauthCallbackHandler mcp.CallbackServer, next http.Handler) (http.Handler, error) {
	if auth.JWKS != "" || auth.TrustedIssuer != "" {
		return userFromJWT(ctx, auth, address, runningUI, oauthCallbackHandler, next)
	} else if auth.OAuthClientID == "" {
		return userFromHeaders(next), nil
	}
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		info := validate.GetTokenInfo(req)
		if info != nil {
			var user mcp.User
			infoString, ok := info.Props["info"].(string)
			if ok {
				_ = json.Unmarshal([]byte(infoString), &user)
			}
			nctx := types.NanobotContext(req.Context())
			nctx.User = user
			nctx.User.ID = info.UserID
			req = req.WithContext(types.WithNanobotContext(req.Context(), nctx))
		}
		next.ServeHTTP(rw, req)
	}), nil
}

func mcpProxy(auth Auth, dsn string, next http.Handler) (_ http.Handler, err error) {
	hash := sha256.Sum256([]byte(strings.TrimSpace(auth.EncryptionKey)))

	if !strings.Contains(dsn, "postgres") {
		dsn = strings.TrimSuffix(dsn, ".db") + "_auth.db"
	}

	u, _ := url.Parse(auth.TrustedAudiences[0])
	proxy, err := proxy.NewOAuthProxy(&proxytypes.Config{
		DatabaseDSN:       dsn,
		OAuthClientID:     auth.OAuthClientID,
		OAuthClientSecret: auth.OAuthClientSecret,
		OAuthAuthorizeURL: auth.OAuthAuthorizeURL,
		ScopesSupported:   strings.Join(auth.OAuthScopes, ","),
		EncryptionKey:     base64.StdEncoding.EncodeToString(hash[:]),
		Mode:              "middleware",
		// TODO(thedadams): This needs to be configurable.
		RoutePrefix: u.Path,
		//CookieName:        "nanobot_auth_code",
		//RequiredAuthPaths: []string{"/mcp", "/api"},
	})
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	proxy.SetupRoutes(mux, next)
	return mux, nil
}

func findJWKSURL(ctx context.Context, issuer string) (string, error) {
	u, err := url.Parse(issuer)
	if err != nil {
		return "", fmt.Errorf("invalid issuer URL: %w", err)
	}

	u.Path = strings.TrimSuffix(fmt.Sprintf("/.well-known/oauth-authorization-server%s", u.Path), "/")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch JWKS URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Infof(ctx, "Unable to fetch OAuth Authorization Server metadata (status %d), using default JWKS URL", resp.StatusCode)
		return fmt.Sprintf("%s/.well-known/jwks.json", issuer), nil
	}

	type Metadata struct {
		JWKSURI string `json:"jwks_uri"`
	}

	var metadata Metadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return "", fmt.Errorf("failed to decode OAuth Authorization Server metadata: %w", err)
	}

	if metadata.JWKSURI == "" {
		log.Infof(ctx, "OAuth Authorization Server metadata does not contain a JWKS URI, using default JWKS URL")
		return fmt.Sprintf("%s/.well-known/jwks.json", issuer), nil
	}

	return metadata.JWKSURI, nil
}

func shouldIgnoreAuth(req *http.Request, healthzPath string) bool {
	if healthzPath != "" && req.URL.Path == healthzPath {
		return true
	}
	return strings.HasPrefix(req.URL.Path, "/oauth")
}
