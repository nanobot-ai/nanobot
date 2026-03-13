package skills

import (
	"context"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Server struct {
	configDir        string
	tools            mcp.ServerTools
	newClient        func(context.Context) (*obotClient, error)
	confirmOverwrite func(context.Context, string) (bool, error)
}

func NewServer(configDir string) *Server {
	s := &Server{
		configDir:        configDir,
		newClient:        newClient,
		confirmOverwrite: defaultConfirmOverwrite,
	}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("searchSkills",
			"Search the Obot skill catalog for skills available to the current agent.",
			s.searchSkills),
		mcp.NewServerTool("installSkill",
			"Download and install a skill from Obot into the local skills workspace.",
			s.installSkill),
		mcp.NewServerTool("deleteSkill",
			"Delete an installed skill by its URI.",
			s.deleteSkill),
	)

	return s
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	switch msg.Method {
	case "initialize":
		mcp.Invoke(ctx, msg, s.initialize)
	case "notifications/initialized":
	case "notifications/cancelled":
		mcp.HandleCancelled(ctx, msg)
	case "tools/list":
		mcp.Invoke(ctx, msg, s.listTools)
	case "tools/call":
		mcp.Invoke(ctx, msg, s.callTool)
	default:
		msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage("%v", msg.Method))
	}
}

func (s *Server) initialize(ctx context.Context, _ mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	if !s.enabled(ctx) {
		return &mcp.InitializeResult{
			ProtocolVersion: params.ProtocolVersion,
			Capabilities:    mcp.ServerCapabilities{},
			ServerInfo: mcp.ServerInfo{
				Name:    version.Name,
				Version: version.Get().String(),
			},
		}, nil
	}

	return &mcp.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Tools: &mcp.ToolsServerCapability{},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    version.Name,
			Version: version.Get().String(),
		},
	}, nil
}

type searchSkillsParams struct {
	Query  string `json:"query,omitempty"`
	RepoID string `json:"repoID,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type searchSkillsResult struct {
	Items []searchSkillsResultItem `json:"items"`
}

type searchSkillsResultItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	RepoID      string `json:"repoID,omitempty"`
	RepoURL     string `json:"repoURL,omitempty"`
	RepoRef     string `json:"repoRef,omitempty"`
	CommitSHA   string `json:"commitSHA,omitempty"`
}

func (s *Server) searchSkills(ctx context.Context, params searchSkillsParams) (*searchSkillsResult, error) {
	client, err := s.newClient(ctx)
	if err != nil {
		return nil, err
	}

	items, err := client.SearchSkills(ctx, params.Query, params.RepoID, params.Limit)
	if err != nil {
		return nil, err
	}

	result := &searchSkillsResult{
		Items: make([]searchSkillsResultItem, 0, len(items)),
	}
	for _, item := range items {
		result.Items = append(result.Items, searchSkillsResultItem{
			ID:          item.ID,
			Name:        item.Name,
			DisplayName: item.DisplayName,
			Description: item.Description,
			RepoID:      item.RepoID,
			RepoURL:     item.RepoURL,
			RepoRef:     item.RepoRef,
			CommitSHA:   item.CommitSHA,
		})
	}

	return result, nil
}

func (s *Server) listTools(ctx context.Context, msg mcp.Message, _ mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	if !s.enabled(ctx) {
		return &mcp.ListToolsResult{Tools: []mcp.Tool{}}, nil
	}
	return s.tools.List(ctx, msg, mcp.ListToolsRequest{})
}

func (s *Server) callTool(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.enabled(ctx) {
		return nil, fmt.Errorf("OBOT_URL is not configured — nanobot skill tools require an Obot platform connection")
	}
	return s.tools.Call(ctx, msg, payload)
}

func (s *Server) enabled(ctx context.Context) bool {
	_, err := s.newClient(ctx)
	return err == nil
}
