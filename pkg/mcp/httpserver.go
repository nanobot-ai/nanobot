package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

type HTTPServer struct {
	MessageHandler MessageHandler
	sessions       sync.Map
}

func NewHTTPServer(handler MessageHandler) *HTTPServer {
	return &HTTPServer{
		MessageHandler: handler,
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

	s, ok := h.sessions.Load(id)
	if !ok {
		http.Error(rw, "Session not found", http.StatusNotFound)
		return
	}

	session := s.(*serverSession)
	rw.Header().Set("Content-Type", "text/event-stream")
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
	}
}

func (h *HTTPServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		h.streamEvents(rw, req)
		return
	}

	streamingID := req.Header.Get("Mcp-Session-Id")
	sseID := req.URL.Query().Get("id")

	if streamingID != "" && req.Method == http.MethodDelete {
		session, ok := h.sessions.LoadAndDelete(streamingID)
		if !ok {
			http.Error(rw, "Session not found", http.StatusNotFound)
			return
		}

		sseSession := session.(*serverSession)
		sseSession.session.Close()
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
		session, ok := h.sessions.Load(streamingID)
		if !ok {
			http.Error(rw, "Session not found", http.StatusNotFound)
			return
		}

		sseSession := session.(*serverSession)
		response, err := sseSession.Exchange(req.Context(), msg)
		if errors.Is(err, ErrNoResponse) {
			rw.WriteHeader(http.StatusAccepted)
			return
		} else if err != nil {
			response = Message{
				JSONRPC: msg.JSONRPC,
				ID:      msg.ID,
				Error: &Error{
					Code:    http.StatusInternalServerError,
					Message: err.Error(),
				},
			}
		}

		rw.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(rw).Encode(response); err != nil {
			http.Error(rw, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
		return
	} else if sseID != "" {
		session, ok := h.sessions.Load(sseID)
		if !ok {
			http.Error(rw, "Session not found", http.StatusNotFound)
			return
		}

		sseSession := session.(*serverSession)
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

	session, err := newServerSession(context.Background(), h.MessageHandler)
	if err != nil {
		http.Error(rw, "Failed to create session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := session.Exchange(req.Context(), msg)
	if err != nil {
		http.Error(rw, "Failed to handle message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.sessions.Store(session.session.sessionID, session)

	rw.Header().Set("Mcp-Session-Id", session.session.sessionID)
	rw.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(rw).Encode(resp); err != nil {
		http.Error(rw, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
