package session

import (
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"github.com/nanobot-ai/nanobot/ui"
	"gorm.io/gorm"
)

func getCookieID(req *http.Request) string {
	cookie, err := req.Cookie("nanobot-session-id")
	if err == nil {
		return cookie.Value
	}
	return ""
}

func UISession(next http.Handler, sessionStore *Manager) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !strings.Contains(strings.ToLower(req.UserAgent()), "mozilla") || req.Header.Get("Mcp-Session-Id") != "" {
			next.ServeHTTP(rw, req)
			return
		}

		nctx := types.NanobotContext(req.Context())
		user := nctx.User
		id := getCookieID(req)

		if id != "" {
			session, err := sessionStore.DB.GetByIDByAccountID(req.Context(), id, complete.First(user.ID, id))
			if errors.Is(err, gorm.ErrRecordNotFound) {
				id = ""
			} else if err != nil {
				http.Error(rw, "Failed to load session: "+err.Error(), http.StatusInternalServerError)
				return
			}
			id = session.SessionID
		}

		if id == "" {
			id = uuid.String()
			err := sessionStore.DB.Create(req.Context(), &Session{
				Type:      "ui",
				SessionID: id,
				AccountID: complete.First(user.ID, id),
				State: State{
					InitializeResult: mcp.InitializeResult{},
					InitializeRequest: mcp.InitializeRequest{
						Capabilities: mcp.ClientCapabilities{
							Elicitation: &struct{}{},
						},
					},
				},
			})
			if err != nil {
				http.Error(rw, "Failed to create session: "+err.Error(), http.StatusInternalServerError)
				return
			}

			cookie := http.Cookie{
				Name:     "nanobot-session-id",
				Value:    id,
				Secure:   false,
				Path:     "/",
				HttpOnly: true,
			}
			http.SetCookie(rw, &cookie)
		}

		if user.ID == "" {
			user.ID = id
			nctx.User = user
			req = req.WithContext(types.WithNanobotContext(req.Context(), nctx))
		}

		req.Header.Set("Mcp-Session-Id", id)

		if strings.HasPrefix(req.URL.Path, "/mcp") {
			next.ServeHTTP(rw, req)
			return
		}

		uiFS, _ := fs.Sub(ui.FS, "dist")
		_, err := fs.Stat(uiFS, "fallback.html")
		if err == nil {
			if _, err := fs.Stat(uiFS, strings.TrimPrefix(req.URL.Path, "/")); err == nil {
				if strings.Contains(req.URL.Path, "immutable") {
					serveGzipAndCached(req, rw, uiFS)
				} else {
					http.FileServer(http.FS(uiFS)).ServeHTTP(rw, req)
				}
			} else {
				http.ServeFileFS(rw, req, uiFS, "fallback.html")
			}
		} else {
			url, _ := url.ParseRequestURI("http://localhost:5173")
			httputil.NewSingleHostReverseProxy(url).ServeHTTP(rw, req)
		}
	})
}

func serveGzipAndCached(req *http.Request, rw http.ResponseWriter, fs fs.FS) {
	path := req.URL.Path
	file, err := fs.Open(strings.TrimPrefix(path, "/"))
	if err != nil {
		http.Error(rw, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(rw, "File stat error", http.StatusInternalServerError)
		return
	}

	// Set cache headers
	rw.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	rw.Header().Set("Last-Modified", info.ModTime().UTC().Format(http.TimeFormat))

	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		ctype = http.DetectContentType(nil)
	}
	rw.Header().Set("Content-Type", ctype)

	// Check if client accepts gzip
	if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		rw.Header().Set("Content-Encoding", "gzip")
		rw.Header().Del("Content-Length")
		gz := gzip.NewWriter(rw)
		defer gz.Close()
		io.Copy(gz, file)
		return
	}

	// Serve uncompressed
	io.Copy(rw, file)
}
