package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/sampling"
	"github.com/obot-platform/nanobot/pkg/types"
)

type Registry struct {
	env        map[string]string
	servers    map[string]map[string]*mcp.Client
	config     types.Config
	serverLock sync.Mutex
	sampler    Sampler
}

type Sampler interface {
	Sample(ctx context.Context, sampling mcp.CreateMessageRequest, opts ...sampling.SamplerOptions) (mcp.CreateMessageResult, error)
}

func NewRegistry(env map[string]string, config types.Config) *Registry {
	return &Registry{
		servers: make(map[string]map[string]*mcp.Client),
		env:     env,
		config:  config,
	}
}

func (r *Registry) SetSampler(sampler Sampler) {
	r.sampler = sampler
}

func (r *Registry) GetClient(ctx context.Context, name string) (*mcp.Client, error) {
	r.serverLock.Lock()
	defer r.serverLock.Unlock()

	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return nil, fmt.Errorf("session not found in context")
	}

	servers, ok := r.servers[strings.Split(session.ID(), "/")[0]]
	if !ok {
		servers = make(map[string]*mcp.Client)
		r.servers[session.ID()] = servers
	}

	s, ok := servers[name]
	if ok {
		return s, nil
	}

	mcpConfig, ok := r.config.MCPServers[name]
	if !ok {
		return nil, fmt.Errorf("MCP server %s not found in config", name)
	}

	clientOpts := mcp.ClientOption{
		Env:       r.env,
		SessionID: session.ID() + "/" + uuid.New().String(),
	}
	if r.sampler != nil {
		clientOpts.OnSampling = func(ctx context.Context, sampling mcp.CreateMessageRequest) (mcp.CreateMessageResult, error) {
			return r.sampler.Sample(ctx, sampling)
		}
	}

	c, err := mcp.NewClient(ctx, name, mcpConfig, clientOpts)
	if err != nil {
		return nil, err
	}

	servers[name] = c
	r.servers[session.ID()] = servers
	return c, nil
}

func (r *Registry) SampleCall(ctx context.Context, agent string, args any, opts ...SampleCallOptions) (*mcp.CallToolResult, error) {
	createMessageRequest, err := r.convertToSampleRequest(agent, args)
	if err != nil {
		return nil, err
	}

	opt := completeSampleCallOptions(opts...)
	sampleOpt := sampling.SamplerOptions{
		Continue: opt.Chat,
	}

	if opt.ProgressToken != nil {
		var cancel func()
		sampleOpt.Progress, cancel = setupProgress(ctx, opt.ProgressToken)
		defer cancel()
	}

	result, err := r.sampler.Sample(ctx, *createMessageRequest, sampleOpt)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			result.Content,
		},
	}, nil
}

func (r *Registry) Call(ctx context.Context, server, tool string, args any, opts ...mcp.CallOption) (*mcp.CallToolResult, error) {
	if _, ok := r.config.Agents[server]; ok {
		opt := mcp.CompleteCallOptions(opts...)
		return r.SampleCall(ctx, server, args, SampleCallOptions{
			ProgressToken: opt.ProgressToken,
		})
	}

	c, err := r.GetClient(ctx, server)
	if err != nil {
		return nil, err
	}

	return c.Call(ctx, tool, args, opts...)
}

type ListToolsOptions struct {
	Servers []string
	Tools   []string
}

type ListToolsResult struct {
	Server string     `json:"server,omitempty"`
	Tools  []mcp.Tool `json:"tools,omitempty"`
}

func (r *Registry) ListTools(ctx context.Context, opts ...ListToolsOptions) (result []ListToolsResult, _ error) {
	var opt ListToolsOptions
	for _, o := range opts {
		for _, server := range o.Servers {
			if server != "" {
				opt.Servers = append(opt.Servers, server)
			}
		}
		for _, tool := range o.Tools {
			if tool != "" {
				opt.Tools = append(opt.Tools, tool)
			}
		}
	}

	serverList := slices.Sorted(maps.Keys(r.config.MCPServers))
	agentsList := slices.Sorted(maps.Keys(r.config.Agents))
	if len(opt.Servers) == 0 {
		opt.Servers = append(serverList, agentsList...)
	}

	for _, server := range opt.Servers {
		if !slices.Contains(serverList, server) {
			continue
		}

		c, err := r.GetClient(ctx, server)
		if err != nil {
			return nil, err
		}

		tools, err := c.ListTools(ctx)
		if err != nil {
			return nil, err
		}

		tools = filterTools(tools, opt.Tools)

		if len(tools.Tools) == 0 {
			continue
		}

		result = append(result, ListToolsResult{
			Server: server,
			Tools:  tools.Tools,
		})
	}

	for _, agentName := range opt.Servers {
		agent, ok := r.config.Agents[agentName]
		if !ok {
			continue
		}

		tools := filterTools(&mcp.ListToolsResult{
			Tools: []mcp.Tool{
				{
					Name:        agentName,
					Description: agent.Description,
					InputSchema: types.AgentInputSchema,
				},
			},
		}, opt.Tools)

		if len(tools.Tools) == 0 {
			continue
		}

		result = append(result, ListToolsResult{
			Server: agentName,
			Tools:  tools.Tools,
		})
	}

	return
}

func filterTools(tools *mcp.ListToolsResult, filter []string) *mcp.ListToolsResult {
	if len(filter) == 0 {
		return tools
	}
	var filteredTools mcp.ListToolsResult
	for _, tool := range tools.Tools {
		if slices.Contains(filter, tool.Name) {
			filteredTools.Tools = append(filteredTools.Tools, tool)
		}
	}
	return &filteredTools
}

func (r *Registry) BuildToolMappings(ctx context.Context, toolList []string) (types.ToolMappings, error) {
	var servers []string
	for _, ref := range toolList {
		toolRef := types.ParseToolRef(ref)
		if toolRef.Server != "" {
			servers = append(servers, toolRef.Server)
		}
	}

	tools, err := r.ListTools(ctx, ListToolsOptions{
		Servers: servers,
	})
	if err != nil {
		return nil, err
	}

	result := types.ToolMappings{}
	for _, ref := range toolList {
		toolRef := types.ParseToolRef(ref)

		for _, t := range tools {
			if t.Server != toolRef.Server {
				continue
			}
			for _, tool := range t.Tools {
				if toolRef.Tool == "" || tool.Name == toolRef.Tool {
					originalName := tool.Name
					if toolRef.As != "" {
						tool.Name = toolRef.As
					}
					result[tool.Name] = types.ToolMapping{
						Server:   toolRef.Server,
						ToolName: originalName,
						Tool:     types.ToolDefinition(tool),
					}
				}
			}
		}
	}

	return result, nil
}

func (r *Registry) convertToSampleRequest(agent string, args any) (*mcp.CreateMessageRequest, error) {
	var sampleArgs types.SampleCallRequest
	if err := types.Marshal(args, &sampleArgs); err != nil {
		return nil, fmt.Errorf("failed to marshal args: %w", err)
	}

	var sampleRequest = mcp.CreateMessageRequest{
		MaxTokens: r.config.Agents[agent].MaxTokens,
		ModelPreferences: mcp.ModelPreferences{
			Hints: []mcp.ModelHint{
				{Name: agent},
			},
		},
	}

	if sampleArgs.Text != "" {
		sampleRequest.Messages = append(sampleRequest.Messages, mcp.SamplingMessage{
			Role: "user",
			Content: mcp.Content{
				Type: "text",
				Text: sampleArgs.Text,
			},
		})
	}

	for _, attachment := range sampleArgs.Attachments {
		if !strings.HasPrefix(attachment.URL, "data:") {
			return nil, fmt.Errorf("invalid attachment URL: %s, only data URI are supported", attachment.URL)
		}
		parts := strings.Split(strings.TrimPrefix(attachment.URL, "data:"), ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid attachment URL: %s, only data URI are supported", attachment.URL)
		}
		mimeType := parts[0]
		if mimeType != "" {
			attachment.MimeType = mimeType
		}
		data := parts[1]
		data, ok := strings.CutPrefix(data, "base64,")
		if !ok {
			return nil, fmt.Errorf("invalid attachment URL: %s, only base64 data URI are supported", attachment.URL)
		}
		sampleRequest.Messages = append(sampleRequest.Messages, mcp.SamplingMessage{
			Role: "user",
			Content: mcp.Content{
				Type:     "image",
				Data:     data,
				MIMEType: attachment.MimeType,
			},
		})
	}

	return &sampleRequest, nil
}

func setupProgress(ctx context.Context, progressToken any) (chan json.RawMessage, func()) {
	session := mcp.SessionFromContext(ctx)
	c := make(chan json.RawMessage, 1)
	go func() {
		var counter int
		for payload := range c {
			counter++
			data, err := json.Marshal(mcp.NotificationProgressRequest{
				ProgressToken: progressToken,
				Progress:      json.Number(fmt.Sprint(counter)),
				Data:          payload,
			})
			if err != nil {
				continue
			}
			_ = session.Send(ctx, mcp.Message{
				Method: "notifications/progress",
				Params: data,
			})
		}
	}()
	return c, func() {
		close(c)
	}
}

type SampleCallOptions struct {
	ProgressToken any
	Chat          bool
}

func completeSampleCallOptions(opts ...SampleCallOptions) SampleCallOptions {
	var opt SampleCallOptions
	for _, o := range opts {
		if o.ProgressToken != nil {
			opt.ProgressToken = o.ProgressToken
		}
		if o.Chat {
			opt.Chat = true
		}
	}
	return opt
}
