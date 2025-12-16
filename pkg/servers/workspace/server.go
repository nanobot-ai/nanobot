package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"github.com/nanobot-ai/nanobot/pkg/version"
	"gorm.io/gorm"
)

var emptyTools mcp.ServerTools

type Server struct {
	store        *Store
	sessionStore *session.Store
	tools        mcp.ServerTools
	toolsService *tools.Service
}

func NewServer(store *Store, sessionStore *session.Store, tools *tools.Service) *Server {
	s := &Server{
		store:        store,
		sessionStore: sessionStore,
		toolsService: tools,
	}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("create_workspace", "Create a new workspace in the database", s.createWorkspace),
		mcp.NewServerTool("update_workspace", "Update an existing workspace in the database", s.updateWorkspace),
		mcp.NewServerTool("delete_workspace", "Delete a workspace from the database", s.deleteWorkspace),
	)

	return s
}

// dbWorkspaceToDisplay converts a database WorkspaceRecord to types.Workspace
func dbWorkspaceToDisplay(workspace *WorkspaceRecord) types.Workspace {
	display := types.Workspace{
		ID:    workspace.UUID,
		Name:  workspace.Name,
		Order: workspace.Order,
		Color: workspace.Color,
	}

	// Extract icons JSON
	if len(workspace.Icons) > 0 {
		var icons []mcp.Icon
		if err := json.Unmarshal(workspace.Icons, &icons); err == nil {
			display.Icons = icons
		}
	}

	// Extract attributes JSON
	if len(workspace.Attributes) > 0 {
		var attrs map[string]interface{}
		if err := json.Unmarshal(workspace.Attributes, &attrs); err == nil {
			display.Attributes = attrs
		}
	}

	return display
}

func (s *Server) listResourcesTemplates(ctx context.Context, _ mcp.Message, _ mcp.ListResourceTemplatesRequest) (*mcp.ListResourceTemplatesResult, error) {
	if s.isInWorkspace(ctx) {
		return &mcp.ListResourceTemplatesResult{
			ResourceTemplates: []mcp.ResourceTemplate{
				{
					URITemplate: "session://session/{uuid}",
					Name:        "Nanobot Sessions",
					Description: "Access session data by session ID",
					MimeType:    types.SessionMimeType,
				},
			},
		}, nil
	}
	return &mcp.ListResourceTemplatesResult{
		ResourceTemplates: []mcp.ResourceTemplate{
			{
				URITemplate: "nanobot://workspaces/{uuid}",
				Name:        "Nanobot Workspace",
				Description: "Access workspace configuration and metadata by UUID",
				MimeType:    types.WorkspaceMimeType,
			},
		},
	}, nil
}

func (s *Server) listSessions(ctx context.Context, accountID string) (*mcp.ListResourcesResult, error) {
	result := &mcp.ListResourcesResult{
		Resources: make([]mcp.Resource, 0),
	}

	// Get current workspace ID
	currentWorkspaceID := types.GetWorkspaceID(ctx)
	if currentWorkspaceID == "" {
		return result, nil
	}

	// Find all workspace records with their session data where parent_id matches the current workspace's parent_id
	workspaces, err := s.store.FindByParentIDWithSessions(ctx, currentWorkspaceID)
	if err != nil {
		return nil, err
	}

	// Return the session IDs from those workspace records
	for _, workspace := range workspaces {
		if workspace.SessionID != "" {
			resource := mcp.Resource{
				URI:      "session://" + workspace.SessionID,
				Name:     workspace.SessionDescription,
				MimeType: types.SessionMimeType,
				Meta: types.Meta(map[string]any{
					"order":       workspace.Order,
					"color":       workspace.Color,
					"workspaceId": workspace.UUID,
				}),
			}

			// Parse and add icons if available
			if len(workspace.Icons) > 0 {
				var icons []mcp.Icon
				if err := json.Unmarshal(workspace.Icons, &icons); err == nil {
					resource.Icons.Icons = icons
				}
			}

			result.Resources = append(result.Resources, resource)
		}
	}

	return result, nil
}

func (s *Server) listResources(ctx context.Context, _ mcp.Message, _ mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	_, accountID := types.GetSessionAndAccountID(ctx)

	result := &mcp.ListResourcesResult{
		Resources: make([]mcp.Resource, 0),
	}

	if s.isInWorkspace(ctx) {
		// return sessions
		return s.listSessions(ctx, accountID)
	}

	// Get workspaces from database store
	workspaces, err := s.store.FindByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	for _, workspace := range workspaces {
		resource := mcp.Resource{
			URI:      "nanobot://workspaces/" + workspace.UUID,
			Name:     workspace.Name,
			Title:    workspace.Name,
			MimeType: types.WorkspaceMimeType,
			Meta: types.Meta(map[string]any{
				"order": workspace.Order,
				"color": workspace.Color,
			}),
		}

		// Parse and add icons if available
		if len(workspace.Icons) > 0 {
			var icons []mcp.Icon
			if err := json.Unmarshal(workspace.Icons, &icons); err == nil {
				resource.Icons.Icons = icons
			}
		}

		result.Resources = append(result.Resources, resource)
	}

	return result, nil
}

func (s *Server) readSession(ctx context.Context, sessionUUID, accountID string) (*mcp.ReadResourceResult, error) {
	// Get session from database by UUID and verify account ownership
	sess, err := s.sessionStore.GetByIDByAccountID(ctx, sessionUUID, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("session not found")
	} else if err != nil {
		return nil, err
	}

	// Build the response data
	responseData := map[string]any{
		"id":        sess.SessionID,
		"createdAt": sess.CreatedAt.Format(time.RFC3339Nano),
		"updatedAt": sess.UpdatedAt.Format(time.RFC3339Nano),
		"title":     sess.Description,
	}

	// Look up the associated workspace by session ID
	workspace, err := s.store.GetBySessionID(ctx, sess.SessionID)
	if err == nil && workspace != nil {
		if workspace.ParentID != nil && *workspace.ParentID != "" {
			responseData["workspaceId"] = *workspace.ParentID
		}
		responseData["sessionWorkspaceId"] = workspace.UUID
	}

	// Marshal the session to JSON
	data, err := json.Marshal(responseData)
	if err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{
			{
				Name:     sess.Description,
				URI:      "session://" + sess.SessionID,
				MIMEType: types.SessionMimeType,
				Text:     &[]string{string(data)}[0],
			},
		},
	}, nil
}

func (s *Server) readResource(ctx context.Context, _ mcp.Message, body mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	_, accountID := types.GetSessionAndAccountID(ctx)

	sessionFromURI, ok := strings.CutPrefix(body.URI, "session://")
	if ok {
		return s.readSession(ctx, sessionFromURI, accountID)
	}

	// Handle nanobot://workspaces/ URIs
	id := strings.TrimPrefix(body.URI, "nanobot://workspaces/")

	// Get workspace from database
	workspace, err := s.store.GetByUUIDAndAccountID(ctx, id, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("workspace not found")
	} else if err != nil {
		return nil, err
	}

	// Convert to WorkspaceDisplay
	display := dbWorkspaceToDisplay(workspace)

	// Marshal the WorkspaceDisplay to JSON
	data, err := json.Marshal(display)
	if err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{
			{
				Name:     display.Name,
				URI:      "nanobot://workspaces/" + display.ID,
				MIMEType: types.WorkspaceMimeType,
				Text:     &[]string{string(data)}[0],
			},
		},
	}, nil
}

type CreateWorkspaceParams struct {
	Name       string         `json:"name"`
	Order      int            `json:"order,omitempty"`
	Color      string         `json:"color,omitempty"`
	Icons      []mcp.Icon     `json:"icons,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

func (s *Server) createWorkspace(ctx context.Context, params CreateWorkspaceParams) (*types.Workspace, error) {
	_, accountID := types.GetSessionAndAccountID(ctx)

	if params.Name == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("name is required")
	}

	// Marshal attributes to JSON
	var attributesJSON []byte
	var err error
	if params.Attributes != nil {
		attributesJSON, err = json.Marshal(params.Attributes)
		if err != nil {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid attributes: %v", err)
		}
	}

	// Marshal icons to JSON
	var iconsJSON []byte
	if params.Icons != nil {
		iconsJSON, err = json.Marshal(params.Icons)
		if err != nil {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid icons: %v", err)
		}
	}

	workspaceUUID := uuid.String()
	workspace := &WorkspaceRecord{
		UUID:       workspaceUUID,
		AccountID:  accountID,
		Name:       params.Name,
		Order:      params.Order,
		Color:      params.Color,
		Icons:      iconsJSON,
		Attributes: attributesJSON,
	}

	c := types.ConfigFromContext(ctx)

	_, err = s.toolsService.Call(ctx, "nanobot.workspace.provider", "sessionCreate", map[string]any{
		"uri": fmt.Sprintf("%s?parentId=%s&baseUri=%s", workspaceUUID, c.WorkspaceID, c.WorkspaceBaseURI),
	})

	if err := s.store.Create(ctx, workspace); err != nil {
		return nil, err
	}

	display := dbWorkspaceToDisplay(workspace)
	return &display, nil
}

type UpdateWorkspaceParams struct {
	URI        string         `json:"uri"`
	Name       string         `json:"name,omitempty"`
	Order      *int           `json:"order,omitempty"`
	Color      string         `json:"color,omitempty"`
	Icons      []mcp.Icon     `json:"icons,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

func (s *Server) updateWorkspace(ctx context.Context, params UpdateWorkspaceParams) (*types.Workspace, error) {
	_, accountID := types.GetSessionAndAccountID(ctx)

	if params.URI == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("uri is required")
	}

	// Extract UUID from URI (portion after nanobot://workspaces/)
	workspaceUUID := strings.TrimPrefix(params.URI, "nanobot://workspaces/")
	if workspaceUUID == params.URI {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid uri format, expected nanobot://workspaces/{uuid}")
	}

	// Get existing workspace and verify ownership
	workspace, err := s.store.GetByUUIDAndAccountID(ctx, workspaceUUID, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("workspace not found")
	} else if err != nil {
		return nil, err
	}

	// Update fields if provided
	if params.Name != "" {
		workspace.Name = params.Name
	}
	if params.Order != nil {
		workspace.Order = *params.Order
	}
	if params.Color != "" {
		workspace.Color = params.Color
	}
	if params.Icons != nil {
		iconsJSON, err := json.Marshal(params.Icons)
		if err != nil {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid icons: %v", err)
		}
		workspace.Icons = iconsJSON
	}
	if params.Attributes != nil {
		attributesJSON, err := json.Marshal(params.Attributes)
		if err != nil {
			return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid attributes: %v", err)
		}
		workspace.Attributes = attributesJSON
	}

	// Update in database
	if err := s.store.Update(ctx, workspace); err != nil {
		return nil, err
	}

	display := dbWorkspaceToDisplay(workspace)
	return &display, nil
}

type DeleteWorkspaceParams struct {
	URI string `json:"uri"`
}

func (s *Server) deleteWorkspace(ctx context.Context, params DeleteWorkspaceParams) (string, error) {
	_, accountID := types.GetSessionAndAccountID(ctx)

	if params.URI == "" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("uri is required")
	}

	// Extract UUID from URI (portion after nanobot://workspaces/)
	workspaceUUID := strings.TrimPrefix(params.URI, "nanobot://workspaces/")
	if workspaceUUID == params.URI {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid uri format, expected nanobot://workspaces/{uuid}")
	}

	// Verify the workspace exists and belongs to this account before deleting
	workspace, err := s.store.GetByUUIDAndAccountID(ctx, workspaceUUID, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", mcp.ErrRPCInvalidParams.WithMessage("workspace not found")
	} else if err != nil {
		return "", err
	}

	// Delete the workspace
	if err := s.store.Delete(ctx, workspace.ID); err != nil {
		return "", err
	}

	return "Workspace deleted successfully", nil
}

func (s *Server) isInWorkspace(ctx context.Context) bool {
	return types.GetWorkspaceID(ctx) != ""
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
		if s.isInWorkspace(ctx) {
			mcp.Invoke(ctx, msg, emptyTools.List)
		} else {
			mcp.Invoke(ctx, msg, s.tools.List)
		}
	case "tools/call":
		if s.isInWorkspace(ctx) {
			mcp.Invoke(ctx, msg, emptyTools.List)
		} else {
			mcp.Invoke(ctx, msg, s.tools.Call)
		}
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
			Resources: &mcp.ResourcesServerCapability{},
			Tools:     &mcp.ToolsServerCapability{},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    version.Name,
			Version: version.Get().String(),
		},
	}, nil
}
