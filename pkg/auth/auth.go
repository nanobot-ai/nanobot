package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/envvar"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/oauth/validate"
	"github.com/obot-platform/mcp-oauth-proxy/pkg/proxy"
	proxytypes "github.com/obot-platform/mcp-oauth-proxy/pkg/types"
)

func Wrap(env map[string]string, cfg types.Config, dsn string, next http.Handler) (http.Handler, error) {
	var (
		result = next
		err    error
	)
	auth := cfg.Auth
	if auth == nil {
		return result, nil
	}

	if err := envvar.ReplaceObject(env, auth); err != nil {
		return nil, fmt.Errorf("failed to replace variables in auth config: %w", err)
	}

	result = setupContext(auth, result)

	if auth.OAuthClientID != "" {
		if auth.OAuthClientSecret == "" {
			return nil, fmt.Errorf("oauthClientSecret is required")
		}
		if len(auth.OAuthScopes) == 0 {
			return nil, fmt.Errorf("oauthScopes is required")
		}
		if auth.OAuthAuthorizeURL == "" {
			return nil, fmt.Errorf("oauthAuthorizeURL is required")
		}

		result, err = mcpProxy(auth, dsn, result)
		if err != nil {
			return nil, fmt.Errorf("failed to create oauth proxy: %w", err)
		}
	}

	if len(auth.OAuthAuthorizationServerMetadata) > 0 {
		panic("not implemented")
	}

	return result, nil
}

func userFromHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var user types.User
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
		next.ServeHTTP(rw, req.WithContext(types.WithNanobotContext(req.Context(), nctx)))
	})
}

func setupContext(auth *types.Auth, next http.Handler) http.Handler {
	if auth.OAuthClientID == "" {
		return userFromHeaders(next)
	}
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		info := validate.GetTokenInfo(req)
		if info != nil {
			var user types.User
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
	})
}

func mcpProxy(auth *types.Auth, dsn string, next http.Handler) (_ http.Handler, err error) {
	hash := sha256.Sum256([]byte(strings.TrimSpace(auth.OAuthClientSecret)))

	if !strings.Contains(dsn, "postgres") {
		dsn = strings.TrimSuffix(dsn, ".db") + "_auth.db"
	}

	proxy, err := proxy.NewOAuthProxy(&proxytypes.Config{
		DatabaseDSN:       dsn,
		OAuthClientID:     auth.OAuthClientID,
		OAuthClientSecret: auth.OAuthClientSecret,
		OAuthAuthorizeURL: auth.OAuthAuthorizeURL,
		ScopesSupported:   strings.Join(auth.OAuthScopes, ","),
		EncryptionKey:     base64.StdEncoding.EncodeToString(hash[:]),
		Mode:              "middleware",
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
