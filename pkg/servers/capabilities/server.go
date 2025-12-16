package capabilities

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/servers/workspace"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"github.com/nanobot-ai/nanobot/pkg/version"
	"gorm.io/gorm"
)

type Server struct {
	store   *workspace.Store
	tools   mcp.ServerTools
	service *tools.Service
}

type Caller interface {
	Call(ctx context.Context, server, tool string, args any, opts ...tools.CallOptions) (ret *types.CallResult, err error)
}

func NewServer(store *workspace.Store, tools *tools.Service) *Server {
	s := &Server{
		store:   store,
		service: tools,
	}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("init_session", "Initializes the session capabilities", s.initSession),
	)

	return s
}

func (s *Server) initSession(ctx context.Context, params types.SessionInitHook) (_ types.SessionInitHook, err error) {
	c := types.ConfigFromContext(ctx)
	if _, ok := c.MCPServers["nanobot.workspace"]; ok {
		params, err = s.initWorkspace(ctx, params)
		if err != nil {
			return params, err
		}
	}
	return params, nil
}

func (s *Server) initWorkspace(ctx context.Context, params types.SessionInitHook) (types.SessionInitHook, error) {
	// never reinit workspace
	if _, ok := params.Meta["workspace"]; ok {
		return params, nil
	}

	if params.Meta == nil {
		params.Meta = make(map[string]any)
	}
	params.Meta["workspace"] = map[string]any{
		"supported": true,
	}

	u, err := url.Parse(params.URL)
	if err != nil {
		return params, err
	}

	workspaceUUID := u.Query().Get("workspace")
	if workspaceUUID == "" {
		return params, nil
	}

	sessionID, accountID := types.GetSessionAndAccountID(ctx)

	// Verify the workspace exists and belongs to this
	currentWorkspace, err := s.store.GetByUUIDAndAccountID(ctx, workspaceUUID, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return params, mcp.ErrRPCInvalidParams.WithMessage("workspace not found")
	} else if err != nil {
		return params, err
	}

	newWorkspace := workspace.WorkspaceRecord{
		Model:     gorm.Model{},
		UUID:      uuid.String(),
		AccountID: accountID,
		ParentID:  &workspaceUUID,
		SessionID: sessionID,
	}

	if u.Query().Get("shared") == "true" {
		newWorkspace = *currentWorkspace
	} else {
		uri := fmt.Sprintf("%s?parentId=%s", newWorkspace.UUID, *newWorkspace.ParentID)

		_, err = s.service.Call(ctx, "nanobot.workspace.provider", "sessionCreate", map[string]any{
			"uri": uri,
		})
		if err != nil {
			return params, fmt.Errorf("failed to create workspace: %w", err)
		}

		if err := s.store.Create(ctx, &newWorkspace); err != nil {
			return params, fmt.Errorf("failed to assign new workspace: %w", err)
		}
	}

	if params.Meta == nil {
		params.Meta = make(map[string]any)
	}

	params.Meta["workspace"] = map[string]any{
		"id":        newWorkspace.UUID,
		"supported": true,
	}
	if newWorkspace.ParentID != nil {
		params.Meta["parentId"] = *newWorkspace.ParentID
	}

	return params, nil
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	switch msg.Method {
	case "initialize":
		mcp.Invoke(ctx, msg, s.initialize)
	case "notifications/initialized":
		// nothing to do
	case "tools/list":
		mcp.Invoke(ctx, msg, s.tools.List)
	case "tools/call":
		mcp.Invoke(ctx, msg, s.tools.Call)
	default:
		msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage("%v", msg.Method))
	}
}

func (s *Server) initialize(ctx context.Context, _ mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return &mcp.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Resources: &mcp.ResourcesServerCapability{},
			Tools:     &mcp.ToolsServerCapability{},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    version.Name,
			Version: version.Get().String(),
		},
	}, nil
}
