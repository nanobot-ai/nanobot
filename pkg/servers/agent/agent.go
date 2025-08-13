package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/sampling"
	"github.com/nanobot-ai/nanobot/pkg/sessiondata"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Server struct {
	tools      mcp.ServerTools
	data       *sessiondata.Data
	agentName  string
	multiAgent bool
	runtime    Caller
}

type Caller interface {
	Call(ctx context.Context, server, tool string, args any, opts ...tools.CallOptions) (ret *types.CallResult, err error)
	GetClient(ctx context.Context, name string) (*mcp.Client, error)
	GetPrompt(ctx context.Context, target, prompt string, args map[string]string) (*mcp.GetPromptResult, error)
}

func NewServer(d *sessiondata.Data, r Caller, name string) *Server {
	s := &Server{
		data:      d,
		agentName: name,
		runtime:   r,
	}

	s.tools = mcp.NewServerTools(
		chatCall{s: s},
	)

	return s
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
	case "resources/list":
		mcp.Invoke(ctx, msg, s.resourcesList)
	case "resources/read":
		mcp.Invoke(ctx, msg, s.resourcesRead)
	case "prompts/list":
		mcp.Invoke(ctx, msg, s.promptsList)
	case "prompts/get":
		mcp.Invoke(ctx, msg, s.promptGet)
	default:
		msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage(msg.Method))
	}
}

func messagesToResourceContents(messages []types.Message) ([]mcp.ResourceContent, error) {
	var contents []mcp.ResourceContent
	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal message: %w", err)
		}
		contents = append(contents, mcp.ResourceContent{
			URI:      fmt.Sprintf(types.MessageURI, msg.ID),
			MIMEType: types.MessageMimeType,
			Text:     string(data),
		})
	}
	return contents, nil
}

func (s *Server) readHistory(ctx context.Context) (ret []mcp.ResourceContent, _ error) {
	chatData, err := getChatCall{
		s: s,
	}.getChatLocal(ctx)
	if err != nil {
		return nil, err
	}

	return messagesToResourceContents(chatData.Messages)
}

func (s *Server) readProgress(ctx context.Context) (ret []mcp.ResourceContent, _ error) {
	var (
		progress types.CompletionResponse
		session  = mcp.SessionFromContext(ctx)
	)

	if !session.Get("progress", &progress) {
		return nil, nil
	}

	callResult, err := sampling.CompletionResponseToCallResult(&progress)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(types.AsyncCallResult{
		IsError:       callResult.IsError,
		Content:       callResult.Content,
		InProgress:    progress.HasMore,
		ToolName:      "chat",
		ProgressToken: progress.ProgressToken,
	})
	if err != nil {
		return nil, err
	}

	return []mcp.ResourceContent{
		{
			URI:      types.ProgressURI,
			MIMEType: types.ToolResultMimeType,
			Text:     string(data),
		},
	}, nil
}

func (s *Server) promptGet(ctx context.Context, _ mcp.Message, payload mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := types.ConfigFromContext(ctx)
	agent := c.Agents[s.agentName]

	promptMappings, err := s.data.BuildPromptMappings(ctx, slices.Concat(agent.MCPServers, agent.Prompts)...)
	if err != nil {
		return nil, err
	}

	promptMapping, ok := promptMappings[payload.Name]
	if !ok {
		return nil, fmt.Errorf("prompt %s not found", payload.Name)
	}

	return s.runtime.GetPrompt(ctx, promptMapping.MCPServer, promptMapping.TargetName, payload.Arguments)
}

func (s *Server) promptsList(ctx context.Context, _ mcp.Message, _ mcp.ListPromptsRequest) (*mcp.ListPromptsResult, error) {
	c := types.ConfigFromContext(ctx)
	agent := c.Agents[s.agentName]
	result := &mcp.ListPromptsResult{}

	prompts, err := s.data.BuildPromptMappings(ctx, slices.Concat(agent.MCPServers, agent.Prompts)...)
	if err != nil {
		return nil, err
	}

	for _, prompt := range prompts {
		result.Prompts = append(result.Prompts, prompt.Target)
	}

	return result, nil
}

func (s *Server) resourcesRead(ctx context.Context, _ mcp.Message, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	var (
		contents []mcp.ResourceContent
		err      error
	)
	switch request.URI {
	case types.HistoryURI:
		contents, err = s.readHistory(ctx)
	case types.ProgressURI:
		contents, err = s.readProgress(ctx)
	}
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: contents,
	}, nil
}

func (s *Server) resourcesList(ctx context.Context, _ mcp.Message, request mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	return &mcp.ListResourcesResult{
		Resources: []mcp.Resource{
			{
				URI:         types.HistoryURI,
				Name:        "chat-history",
				Title:       "Chat History",
				Description: "The chat history for the current agent.",
				MimeType:    types.HistoryMimeType,
			},
			{
				URI:         types.ProgressURI,
				Name:        "chat-progress",
				Title:       "Chat Streaming Progress",
				Description: "The streaming content of the current or last chat exchange.",
				MimeType:    types.ToolResultMimeType,
			},
		},
	}, nil
}

func (s *Server) initialize(ctx context.Context, _ mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	agents, err := s.data.Agents(ctx)
	if err != nil {
		return nil, err
	}

	if len(agents) <= 1 {
		delete(s.tools, "set_current_agent")
	} else {
		s.multiAgent = true
	}

	return &mcp.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Tools:     &mcp.ToolsServerCapability{},
			Prompts:   &mcp.PromptsServerCapability{},
			Resources: &mcp.ResourcesServerCapability{},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    version.Name,
			Version: version.Get().String(),
		},
	}, nil
}
