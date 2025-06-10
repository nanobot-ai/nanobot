package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/obot-platform/nanobot/pkg/complete"
	"github.com/obot-platform/nanobot/pkg/envvar"
	"github.com/obot-platform/nanobot/pkg/log"
	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/sampling"
	"github.com/obot-platform/nanobot/pkg/types"
	"github.com/obot-platform/nanobot/pkg/uuid"
)

type Service struct {
	servers    map[string]map[string]*mcp.Client
	roots      []mcp.Root
	config     types.Config
	serverLock sync.Mutex
	sampler    Sampler
}

type Sampler interface {
	Sample(ctx context.Context, sampling mcp.CreateMessageRequest, opts ...sampling.SamplerOptions) (mcp.CreateMessageResult, error)
}

type RegistryOptions struct {
	Roots []mcp.Root
}

func completeOptions(opts ...RegistryOptions) RegistryOptions {
	var options RegistryOptions
	for _, opt := range opts {
		options.Roots = append(options.Roots, opt.Roots...)
	}
	if options.Roots == nil {
		options.Roots = []mcp.Root{}
	}
	return options
}

func NewToolsService(config types.Config, opts ...RegistryOptions) *Service {
	return &Service{
		servers: make(map[string]map[string]*mcp.Client),
		config:  config,
		roots:   completeOptions(opts...).Roots,
	}
}

func (r *Service) SetSampler(sampler Sampler) {
	r.sampler = sampler
}

func (r *Service) GetDynamicInstruction(ctx context.Context, instruction types.DynamicInstructions) (string, error) {
	if !instruction.IsSet() {
		return "", nil
	}
	if !instruction.IsPrompt() {
		return instruction.Instructions, nil
	}

	session := mcp.SessionFromContext(ctx)

	prompt, err := r.GetPrompt(ctx, instruction.MCPServer, instruction.Prompt, envvar.ReplaceMap(session.EnvMap(), instruction.Args))
	if err != nil {
		return "", fmt.Errorf("failed to get prompt: %w", err)
	}
	if len(prompt.Messages) != 1 {
		return "", fmt.Errorf("prompt %s/%s returned %d messages, expected 1",
			instruction.MCPServer, instruction.Prompt, len(prompt.Messages))
	}
	return prompt.Messages[0].Content.Text, nil
}

func (r *Service) GetPrompt(ctx context.Context, target, prompt string, args map[string]string) (*mcp.GetPromptResult, error) {
	c, err := r.GetClient(ctx, target)
	if err != nil {
		return nil, err
	}

	return c.GetPrompt(ctx, prompt, args)
}

func (r *Service) GetClient(ctx context.Context, name string) (*mcp.Client, error) {
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

	var roots mcp.ListRootsResult
	if session.ClientCapabilities != nil && session.ClientCapabilities.Roots != nil {
		err := session.Exchange(ctx, "roots/list", mcp.ListRootsRequest{}, &roots)
		if err != nil {
			return nil, fmt.Errorf("failed to list roots: %w", err)
		}
	}

	mcpConfig, ok := r.config.MCPServers[name]
	if !ok {
		return nil, fmt.Errorf("MCP server %s not found in config", name)
	}

	if len(r.roots) > 0 {
		roots.Roots = append(roots.Roots, r.roots...)
	}

	clientOpts := mcp.ClientOption{
		Roots:         roots.Roots,
		Env:           session.EnvMap(),
		ParentSession: session,
		SessionID:     session.ID() + "/" + uuid.String(),
		OnRoots: func(ctx context.Context, msg mcp.Message) error {
			return msg.Reply(ctx, mcp.ListRootsResult{
				Roots: roots.Roots,
			})
		},
		OnLogging: func(ctx context.Context, logMsg mcp.LoggingMessage) error {
			data, err := json.Marshal(mcp.LoggingMessage{
				Level:  logMsg.Level,
				Logger: logMsg.Logger,
				Data: map[string]any{
					"server": name,
					"data":   logMsg.Data,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to marshal logging message: %w", err)
			}
			for session.Parent != nil {
				session = session.Parent
			}
			return session.Send(ctx, mcp.Message{
				Method: "notifications/message",
				Params: data,
			})
		},
	}
	if r.sampler != nil {
		clientOpts.OnSampling = func(ctx context.Context, samplingRequest mcp.CreateMessageRequest) (mcp.CreateMessageResult, error) {
			return r.sampler.Sample(ctx, samplingRequest, sampling.SamplerOptions{
				ProgressToken: uuid.String(),
			})
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

func (r *Service) SampleCall(ctx context.Context, agent string, args any, opts ...SampleCallOptions) (*mcp.CallToolResult, error) {
	createMessageRequest, err := r.convertToSampleRequest(agent, args)
	if err != nil {
		return nil, err
	}

	opt := complete.Complete(opts...)

	result, err := r.sampler.Sample(ctx, *createMessageRequest, sampling.SamplerOptions{
		ProgressToken: opt.ProgressToken,
		AgentOverride: opt.AgentOverride,
	})
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			result.Content,
		},
	}, nil
}

type CallOptions struct {
	ProgressToken any
	AgentOverride types.AgentCall
}

func (o CallOptions) Merge(other CallOptions) (result CallOptions) {
	result.ProgressToken = complete.Last(o.ProgressToken, other.ProgressToken)
	result.AgentOverride = complete.Merge(o.AgentOverride, other.AgentOverride)
	return
}

func (r *Service) Call(ctx context.Context, server, tool string, args any, opts ...CallOptions) (ret *mcp.CallToolResult, err error) {
	defer func() {
		if ret == nil {
			return
		}
		for i, content := range ret.Content {
			if content.Text != "" {
				var obj any
				if err := json.Unmarshal([]byte(content.Text), &obj); err == nil {
					ret.Content[i].StructuredContent = obj
				}
			}
		}
	}()

	opt := complete.Complete(opts...)
	session := mcp.SessionFromContext(ctx)

	target := server
	if tool != "" {
		target = server + "/" + tool
	}

	if session != nil && opt.ProgressToken != nil {
		callID := uuid.String()
		_ = session.SendPayload(ctx, "notifications/progress", mcp.NotificationProgressRequest{
			ProgressToken: opt.ProgressToken,
			Data: map[string]any{
				"type":   "nanobot/call",
				"id":     callID,
				"target": target,
				"input":  args,
			},
		})

		defer func() {
			if err == nil {
				_ = session.SendPayload(ctx, "notifications/progress", mcp.NotificationProgressRequest{
					ProgressToken: opt.ProgressToken,
					Data: map[string]any{
						"type":   "nanobot/call/complete",
						"id":     callID,
						"target": target,
						"output": ret,
					},
				})
			} else {
				_ = session.SendPayload(ctx, "notifications/progress", mcp.NotificationProgressRequest{
					ProgressToken: opt.ProgressToken,
					Data: map[string]any{
						"type":   "nanobot/toolcall/error",
						"id":     callID,
						"target": target,
						"error":  err.Error(),
					},
				})
			}
		}()
	}

	if _, ok := r.config.Agents[server]; ok {
		return r.SampleCall(ctx, server, args, SampleCallOptions{
			ProgressToken: opt.ProgressToken,
			AgentOverride: opt.AgentOverride,
		})
	}

	if _, ok := r.config.Flows[server]; ok {
		return r.startFlow(ctx, server, args, opt)
	}

	c, err := r.GetClient(ctx, server)
	if err != nil {
		return nil, err
	}

	return c.Call(ctx, tool, args, mcp.CallOption{
		ProgressToken: opt.ProgressToken,
	})
}

type ListToolsOptions struct {
	Servers []string
	Tools   []string
}

type ListToolsResult struct {
	Server string     `json:"server,omitempty"`
	Tools  []mcp.Tool `json:"tools,omitempty"`
}

func (r *Service) ListTools(ctx context.Context, opts ...ListToolsOptions) (result []ListToolsResult, _ error) {
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
	flowsList := slices.Sorted(maps.Keys(r.config.Flows))
	if len(opt.Servers) == 0 {
		opt.Servers = append(serverList, agentsList...)
		opt.Servers = append(opt.Servers, flowsList...)
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
					InputSchema: types.ChatInputSchema,
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

	for _, flowName := range opt.Servers {
		flow, ok := r.config.Flows[flowName]
		if !ok {
			continue
		}

		tools := filterTools(&mcp.ListToolsResult{
			Tools: []mcp.Tool{
				{
					Name:        flowName,
					Description: flow.Description,
					InputSchema: flow.Input.ToSchema(),
				},
			},
		}, opt.Tools)

		if len(tools.Tools) == 0 {
			continue
		}

		result = append(result, ListToolsResult{
			Server: flowName,
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

func (r *Service) getMatches(ref string, tools []ListToolsResult) types.ToolMappings {
	toolRef := types.ParseToolRef(ref)
	result := types.ToolMappings{}

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
				result[tool.Name] = types.TargetMapping{
					MCPServer:  toolRef.Server,
					TargetName: originalName,
					Target:     tool,
				}
			}
		}
	}

	return result
}

func (r *Service) GetEntryPoint(ctx context.Context, existingTools types.ToolMappings) (types.TargetMapping, error) {
	if tm, ok := existingTools[types.AgentTool]; ok {
		return tm, nil
	}

	entrypoint := r.config.Publish.Entrypoint
	if entrypoint == "" {
		return types.TargetMapping{}, fmt.Errorf("no entrypoint specified")
	}

	tools, err := r.listToolsForReferences(ctx, []string{entrypoint})
	if err != nil {
		return types.TargetMapping{}, err
	}

	agents := r.getMatches(entrypoint, tools)
	if len(agents) != 1 {
		return types.TargetMapping{}, fmt.Errorf("expected one agent for entrypoint %s, got %d", entrypoint, len(agents))
	}
	for _, v := range agents {
		target := v.Target.(mcp.Tool)
		target.Name = types.AgentTool
		v.Target = target
		return v, nil
	}
	panic("unreachable")
}

func (r *Service) listToolsForReferences(ctx context.Context, toolList []string) ([]ListToolsResult, error) {
	if len(toolList) == 0 {
		return nil, nil
	}

	var servers []string
	for _, ref := range toolList {
		toolRef := types.ParseToolRef(ref)
		if toolRef.Server != "" {
			servers = append(servers, toolRef.Server)
		}
	}

	return r.ListTools(ctx, ListToolsOptions{
		Servers: servers,
	})
}

func (r *Service) BuildToolMappings(ctx context.Context, toolList []string) (types.ToolMappings, error) {
	tools, err := r.listToolsForReferences(ctx, toolList)
	if err != nil {
		return nil, err
	}

	result := types.ToolMappings{}
	for _, ref := range toolList {
		maps.Copy(result, r.getMatches(ref, tools))
	}

	return result, nil
}

func hasOnlySampleKeys(args map[string]any) bool {
	for key := range args {
		if key != "prompt" && key != "attachments" {
			return false
		}
	}
	return true
}

func (r *Service) convertToSampleRequest(agent string, args any) (*mcp.CreateMessageRequest, error) {
	var sampleArgs types.SampleCallRequest
	switch args := args.(type) {
	case string:
		sampleArgs.Prompt = args
	case map[string]any:
		if hasOnlySampleKeys(args) {
			if err := types.Marshal(args, &sampleArgs); err != nil {
				return nil, fmt.Errorf("failed to marshal args: %w", err)
			}
		} else {
			if err := types.Marshal(args, &sampleArgs.Prompt); err != nil {
				return nil, fmt.Errorf("failed to marshal args to prompt: %w", err)
			}
		}
	case *types.SampleCallRequest:
		if args != nil {
			if err := types.Marshal(*args, &sampleArgs); err != nil {
				return nil, fmt.Errorf("failed to marshal args to prompt: %w", err)
			}
		}
	default:
		if err := types.Marshal(args, &sampleArgs); err != nil {
			return nil, fmt.Errorf("failed to marshal args to prompt: %w", err)
		}
	}

	var sampleRequest = mcp.CreateMessageRequest{
		MaxTokens: r.config.Agents[agent].MaxTokens,
		ModelPreferences: mcp.ModelPreferences{
			Hints: []mcp.ModelHint{
				{Name: agent},
			},
		},
	}

	if sampleArgs.Prompt != "" {
		sampleRequest.Messages = append(sampleRequest.Messages, mcp.SamplingMessage{
			Role: "user",
			Content: mcp.Content{
				Type: "text",
				Text: sampleArgs.Prompt,
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
	for session.Parent != nil {
		session = session.Parent
	}
	c := make(chan json.RawMessage, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
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
			err = session.Send(ctx, mcp.Message{
				Method: "notifications/progress",
				Params: data,
			})
			if err != nil {
				log.Errorf(ctx, "failed to send progress notification: %v", err)
			}
		}
	}()
	return c, func() {
		close(c)
		<-done
	}
}

type SampleCallOptions struct {
	ProgressToken any
	AgentOverride types.AgentCall
}

func (s SampleCallOptions) Merge(other SampleCallOptions) (result SampleCallOptions) {
	result.ProgressToken = complete.Last(s.ProgressToken, other.ProgressToken)
	result.AgentOverride = complete.Merge(s.AgentOverride, other.AgentOverride)
	return
}
