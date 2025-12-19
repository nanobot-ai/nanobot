package resources

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"github.com/nanobot-ai/nanobot/pkg/version"
	"gorm.io/gorm"
)

type Server struct {
	tools        mcp.ServerTools
	toolsService *tools.Service
	store        *Store
	sessionStore *session.Store
}

func NewServer(store *Store, toolsService *tools.Service, sessionStore *session.Store) *Server {
	s := &Server{
		store:        store,
		toolsService: toolsService,
		sessionStore: sessionStore,
	}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("create_resource", "Create a resource", s.createResource),
		mcp.NewServerTool("delete_resource", "Delete a resource", s.deleteResource),
	)

	return s
}

type GetArtifactParams struct {
	ArtifactID string `json:"artifactID"`
}

type CreateArtifactParams struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Blob        string `json:"blob"`
	MimeType    string `json:"mimeType,omitempty"`
}

type DeleteResourceParams struct {
	URI string `json:"uri"`
}

// detectMimeType attempts to determine if data is text or binary
// Returns "text/plain" for UTF-8 compatible content, "application/octet-stream" for binary
func detectMimeType(data []byte) string {
	// Check if the data is valid UTF-8
	if utf8.Valid(data) {
		return "text/plain"
	}
	return "application/octet-stream"
}

func (s *Server) workspaceWrite(ctx context.Context, params CreateArtifactParams) (*mcp.Resource, error) {
	workspaceID := types.GetWorkspaceID(ctx)
	if workspaceID == "" {
		return nil, fmt.Errorf("no workspace in set in session")
	}

	path := strings.TrimPrefix(params.Name, "workspace://")
	_, err := s.toolsService.Call(ctx, "nanobot.workspace.provider", "writeTextFile", map[string]any{
		"sessionId": workspaceID,
		"path":      path,
		"content":   params.Blob,
		"encoding":  "base64",
	})
	if err != nil {
		return nil, err
	}
	return &mcp.Resource{
		URI:  params.Name,
		Name: path,
	}, nil
}

func (s *Server) workspaceDelete(ctx context.Context, uri string) (*mcp.Resource, error) {
	workspaceID := types.GetWorkspaceID(ctx)
	if workspaceID == "" {
		return nil, fmt.Errorf("no workspace in set in session")
	}

	path := strings.TrimPrefix(uri, "workspace://")
	_, err := s.toolsService.Call(ctx, "nanobot.workspace.provider", "deleteFile", map[string]any{
		"sessionId": workspaceID,
		"path":      path,
	})
	if err != nil {
		return nil, err
	}
	return &mcp.Resource{
		URI:  uri,
		Name: path,
	}, nil
}

func (s *Server) sessionDelete(ctx context.Context, uri string) (*mcp.Resource, error) {
	if s.sessionStore == nil {
		return nil, fmt.Errorf("session store not available")
	}

	_, accountID := types.GetSessionAndAccountID(ctx)

	sessionID := strings.TrimPrefix(uri, "session://")

	// Verify the session exists and belongs to this account
	sess, err := s.sessionStore.GetByIDByAccountID(ctx, sessionID, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("session not found or access denied")
	} else if err != nil {
		return nil, err
	}

	// Delete the session
	err = s.sessionStore.Delete(ctx, sess.SessionID)
	if err != nil {
		return nil, err
	}

	return &mcp.Resource{
		URI:  uri,
		Name: sess.Description,
	}, nil
}

func (s *Server) createResource(ctx context.Context, params CreateArtifactParams) (*mcp.Resource, error) {
	sessionID, accountID := types.GetSessionAndAccountID(ctx)

	data, err := base64.StdEncoding.DecodeString(params.Blob)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid base64 data: %v", err)
	}

	// Detect mimetype if not provided
	mimeType := params.MimeType
	if mimeType == "" {
		mimeType = detectMimeType(data)
	}

	if strings.HasPrefix(params.Name, "workspace://") {
		return s.workspaceWrite(ctx, params)
	}

	// Check if a resource with this name already exists
	existing, err := s.store.GetByNameSessionIDAndAccountID(ctx, params.Name, sessionID, accountID)
	if err == nil {
		// Resource exists, update it
		if params.Blob != "" {
			existing.Blob = params.Blob
		}
		if params.MimeType != "" {
			existing.MimeType = params.MimeType
		} else {
			// If mimetype not provided in update, detect it
			existing.MimeType = mimeType
		}
		if params.Description != "" {
			existing.Description = params.Description
		}

		err = s.store.Update(ctx, existing)
		if err != nil {
			return nil, err
		}

		return &mcp.Resource{
			URI:         "nanobot://resource/" + existing.UUID,
			Name:        existing.Name,
			Description: existing.Description,
			MimeType:    existing.MimeType,
			Size:        int64(len(data)),
		}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// If error is not "not found", return the error
		return nil, err
	}

	// Resource doesn't exist, create a new one
	resourceUUID := uuid.String()
	err = s.store.Create(ctx, &Resource{
		UUID:        resourceUUID,
		SessionID:   sessionID,
		AccountID:   accountID,
		Blob:        params.Blob,
		MimeType:    mimeType,
		Name:        params.Name,
		Description: params.Description,
	})
	if err != nil {
		return nil, err
	}

	return &mcp.Resource{
		URI:         "nanobot://resource/" + resourceUUID,
		Name:        params.Name,
		Description: params.Description,
		MimeType:    mimeType,
		Size:        int64(len(data)),
	}, nil
}

func (s *Server) deleteResource(ctx context.Context, params DeleteResourceParams) (*mcp.Resource, error) {
	if strings.HasPrefix(params.URI, "workspace://") {
		return s.workspaceDelete(ctx, params.URI)
	}

	if strings.HasPrefix(params.URI, "session://") {
		return s.sessionDelete(ctx, params.URI)
	}

	_, accountID := types.GetSessionAndAccountID(ctx)

	id := strings.TrimPrefix(params.URI, "nanobot://resource/")

	resource, err := s.store.GetByUUIDAndAccountID(ctx, id, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("resource not found")
	} else if err != nil {
		return nil, err
	}

	err = s.store.Delete(ctx, resource.ID)
	if err != nil {
		return nil, err
	}

	return &mcp.Resource{
		URI:         "nanobot://resource/" + resource.UUID,
		Name:        resource.Name,
		Description: resource.Description,
		MimeType:    resource.MimeType,
		Size:        int64(base64.StdEncoding.DecodedLen(len(resource.Blob))),
	}, nil
}

func (s *Server) readResource(ctx context.Context, _ mcp.Message, body mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	_, accountID := types.GetSessionAndAccountID(ctx)

	id := strings.TrimPrefix(body.URI, "nanobot://resource/")

	artifact, err := s.store.GetByUUIDAndAccountID(ctx, id, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("artifact not found")
	} else if err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{
			{
				Name:     artifact.Name,
				URI:      "nanobot://resource/" + artifact.UUID,
				MIMEType: artifact.MimeType,
				Blob:     &artifact.Blob,
			},
		},
	}, nil
}

func (s *Server) listResourcesTemplates(_ context.Context, _ mcp.Message, _ mcp.ListResourceTemplatesRequest) (*mcp.ListResourceTemplatesResult, error) {
	return &mcp.ListResourceTemplatesResult{
		ResourceTemplates: make([]mcp.ResourceTemplate, 0),
	}, nil
}

func (s *Server) listResources(ctx context.Context, _ mcp.Message, body mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	sessionID, _ := types.GetSessionAndAccountID(ctx)

	resources, err := s.store.FindBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	result := &mcp.ListResourcesResult{
		Resources: make([]mcp.Resource, 0, len(resources)),
	}

	for _, resource := range resources {
		result.Resources = append(result.Resources, mcp.Resource{
			URI:         "nanobot://resource/" + resource.UUID,
			Name:        resource.Name,
			Description: resource.Description,
			MimeType:    resource.MimeType,
			Size:        int64(base64.StdEncoding.DecodedLen(len(resource.Blob))),
		})
	}

	return result, nil
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	switch msg.Method {
	case "initialize":
		mcp.Invoke(ctx, msg, s.initialize)
	case "notifications/initialized":
		// nothing to do
	case "resources/read":
		mcp.Invoke(ctx, msg, s.readResource)
	case "resources/list":
		mcp.Invoke(ctx, msg, s.listResources)
	case "resources/templates/list":
		mcp.Invoke(ctx, msg, s.listResourcesTemplates)
	case "tools/list":
		mcp.Invoke(ctx, msg, s.tools.List)
	case "tools/call":
		mcp.Invoke(ctx, msg, s.tools.Call)
	default:
		msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage("%v", msg.Method))
	}
}

func (s *Server) initialize(ctx context.Context, _ mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	if !types.IsUISession(ctx) {
		s.tools = mcp.NewServerTools()
		return &mcp.InitializeResult{
			ProtocolVersion: params.ProtocolVersion,
			ServerInfo: mcp.ServerInfo{
				Name:    version.Name,
				Version: version.Get().String(),
			},
		}, nil
	}

	return &mcp.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Tools:     &mcp.ToolsServerCapability{},
			Resources: &mcp.ResourcesServerCapability{},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    version.Name,
			Version: version.Get().String(),
		},
	}, nil
}
