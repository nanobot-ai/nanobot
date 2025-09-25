package resources

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"github.com/nanobot-ai/nanobot/pkg/version"
	"gorm.io/gorm"
)

type Server struct {
	tools mcp.ServerTools
	store *Store
}

func NewServer(store *Store) *Server {
	s := &Server{
		store: store,
	}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("create_resource", "Create a resource", s.createResource),
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
	MimeType    string `json:"mimeType"`
}

func (s *Server) createResource(ctx context.Context, params CreateArtifactParams) (*mcp.Resource, error) {
	sessionID, accountID := s.getSessionAndAccountID(ctx)

	data, err := base64.StdEncoding.DecodeString(params.Blob)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid base64 data: " + err.Error())
	}

	uuid := uuid.String()
	err = s.store.Create(ctx, &Resource{
		UUID:        uuid,
		SessionID:   sessionID,
		AccountID:   accountID,
		Blob:        params.Blob,
		MimeType:    params.MimeType,
		Name:        params.Name,
		Description: params.Description,
	})
	if err != nil {
		return nil, err
	}

	return &mcp.Resource{
		URI:         "nanobot://resource/" + uuid,
		Name:        params.Name,
		Description: params.Description,
		MimeType:    params.MimeType,
		Size:        int64(len(data)),
	}, nil
}

func (s *Server) readResource(ctx context.Context, _ mcp.Message, body mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	_, accountID := s.getSessionAndAccountID(ctx)

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
				URI:      "nanobot://resource/" + artifact.UUID,
				MIMEType: artifact.MimeType,
				Blob:     artifact.Blob,
			},
		},
	}, nil
}

func (s *Server) getSessionAndAccountID(ctx context.Context) (string, string) {
	var (
		session   = mcp.SessionFromContext(ctx)
		accountID string
	)
	session = session.Parent
	session.Get(types.AccountIDSessionKey, &accountID)
	return session.ID(), accountID
}

func (s *Server) listResourcesTemplates(_ context.Context, _ mcp.Message, _ mcp.ListResourceTemplatesRequest) (*mcp.ListResourceTemplatesResult, error) {
	return &mcp.ListResourceTemplatesResult{
		ResourceTemplates: make([]mcp.ResourceTemplate, 0),
	}, nil
}

func (s *Server) listResources(ctx context.Context, _ mcp.Message, body mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	sessionID, _ := s.getSessionAndAccountID(ctx)

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
		msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage(msg.Method))
	}
}

func (s *Server) initialize(_ context.Context, _ mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
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
