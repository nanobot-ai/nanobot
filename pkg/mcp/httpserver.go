package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

type HTTPServer struct {
	env            map[string]string
	MessageHandler MessageHandler
	sessions       SessionStore
	ctx            context.Context
}

type HTTPServerOptions struct {
	SessionStore SessionStore
	BaseContext  context.Context
}

func (h HTTPServerOptions) Complete() HTTPServerOptions {
	if h.SessionStore == nil {
		h.SessionStore = NewInMemorySessionStore()
	}
	if h.BaseContext == nil {
		h.BaseContext = context.Background()
	}
	return h
}

func (h HTTPServerOptions) Merge(other HTTPServerOptions) (result HTTPServerOptions) {
	h.SessionStore = complete.Last(h.SessionStore, other.SessionStore)
	h.BaseContext = complete.Last(h.BaseContext, other.BaseContext)
	return h
}

func NewHTTPServer(env map[string]string, handler MessageHandler, opts ...HTTPServerOptions) *HTTPServer {
	o := complete.Complete(opts...)
	return &HTTPServer{
		MessageHandler: handler,
		env:            env,
		sessions:       o.SessionStore,
		ctx:            o.BaseContext,
	}
}

func (h *HTTPServer) streamEvents(rw http.ResponseWriter, req *http.Request) {
	id := req.Header.Get("Mcp-Session-Id")
	if id == "" {
		id = req.URL.Query().Get("id")
	}

	if id == "" {
		http.Error(rw, "Session ID is required", http.StatusBadRequest)
		return
	}

	session, ok, err := h.sessions.Load(req, id)
	if err != nil {
		http.Error(rw, "Failed to load session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(rw, "Session not found", http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.WriteHeader(http.StatusOK)
	if flusher, ok := rw.(http.Flusher); ok {
		flusher.Flush()
	}
	for {
		msg, ok := session.Read(req.Context())
		if !ok {
			return
		}

		data, _ := json.Marshal(msg)
		_, err := rw.Write([]byte("data: " + string(data) + "\n\n"))
		if err != nil {
			http.Error(rw, "Failed to write message: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if f, ok := rw.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func (h *HTTPServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		h.streamEvents(rw, req)
		return
	}

	streamingID := req.Header.Get("Mcp-Session-Id")
	sseID := req.URL.Query().Get("id")

	if streamingID != "" && req.Method == http.MethodDelete {
		sseSession, ok, err := h.sessions.LoadAndDelete(req, streamingID)
		if err != nil {
			http.Error(rw, "Failed to delete session: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(rw, "Session not found", http.StatusNotFound)
			return
		}

		sseSession.Close()
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	if req.Method != http.MethodPost {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg Message
	if err := json.NewDecoder(req.Body).Decode(&msg); err != nil {
		http.Error(rw, "Failed to decode message: "+err.Error(), http.StatusBadRequest)
		return
	}

	if streamingID != "" {
		streamingSession, ok, err := h.sessions.Load(req, streamingID)
		if err != nil {
			http.Error(rw, "Failed to load session: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(rw, "Session not found", http.StatusNotFound)
			return
		}

		var setID bool
		if msg.ID == nil {
			msg.ID = uuid.String()
			setID = true
		}

		response, err := streamingSession.Exchange(req.Context(), msg)
		if setID {
			response.ID = nil
		}
		if errors.Is(err, ErrNoResponse) {
			rw.WriteHeader(http.StatusAccepted)
			return
		} else if err != nil {
			response = Message{
				JSONRPC: msg.JSONRPC,
				ID:      msg.ID,
				Error: &RPCError{
					Code:    -32603,
					Message: err.Error(),
				},
			}
		}

		rw.Header().Set("Content-Type", "application/json")

		if (len(response.Result) <= 2 && response.Error == nil) && msg.Method != "ping" {
			// Response has no data, write status accepted.
			rw.WriteHeader(http.StatusAccepted)
		}

		if err := json.NewEncoder(rw).Encode(response); err != nil {
			http.Error(rw, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}

		_ = h.sessions.Store(req, streamingSession.ID(), streamingSession)
		return
	} else if sseID != "" {
		sseSession, ok, err := h.sessions.Load(req, sseID)
		if err != nil {
			http.Error(rw, "Failed to load session: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(rw, "Session not found", http.StatusNotFound)
			return
		}

		if err := sseSession.Send(req.Context(), msg); err != nil {
			http.Error(rw, "Failed to handle message: "+err.Error(), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusAccepted)
		return
	}

	if msg.Method != "initialize" {
		http.Error(rw, fmt.Sprintf("Method %s not allowed", msg.Method), http.StatusMethodNotAllowed)
		return
	}

	session, err := NewServerSession(h.ctx, h.MessageHandler)
	if err != nil {
		http.Error(rw, "Failed to create session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	maps.Copy(session.session.EnvMap(), h.getEnv(req))
	if err := h.sessions.Store(req, session.ID(), session); err != nil {
		http.Error(rw, "Failed to store session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := session.Exchange(req.Context(), msg)
	if err != nil {
		http.Error(rw, "Failed to handle message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Mcp-Session-Id", session.ID())
	rw.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(rw).Encode(resp); err != nil {
		http.Error(rw, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *HTTPServer) getEnv(req *http.Request) map[string]string {
	env := make(map[string]string)
	maps.Copy(env, h.env)
	token, ok := strings.CutPrefix(req.Header.Get("Authorization"), "Bearer ")
	if ok {
		env["http:bearer-token"] = token
	}
	for k, v := range req.Header {
		if key, ok := strings.CutPrefix(k, "X-Nanobot-Env-"); ok {
			env[key] = strings.Join(v, ", ")
		}
	}
	return env
}
