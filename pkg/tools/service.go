package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/envvar"
	"github.com/nanobot-ai/nanobot/pkg/expr"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/sampling"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

type Service struct {
	servers          map[string]map[string]*mcp.Client
	roots            []mcp.Root
	serverLock       sync.Mutex
	sampler          Sampler
	runner           mcp.Runner
	callbackHandler  mcp.CallbackHandler
	oauthRedirectURL string
	concurrency      int
}

type Sampler interface {
	Sample(ctx context.Context, sampling mcp.CreateMessageRequest, opts ...sampling.SamplerOptions) (*types.CallResult, error)
}

type RegistryOptions struct {
	Roots            []mcp.Root
	Concurrency      int
	CallbackHandler  mcp.CallbackHandler
	OAuthRedirectURL string
}

func (r RegistryOptions) Merge(other RegistryOptions) (result RegistryOptions) {
	result.Roots = append(r.Roots, other.Roots...)
	result.Concurrency = complete.Last(r.Concurrency, other.Concurrency)
	result.CallbackHandler = complete.Last(r.CallbackHandler, other.CallbackHandler)
	result.OAuthRedirectURL = complete.Last(r.OAuthRedirectURL, other.OAuthRedirectURL)
	return result
}

func (r RegistryOptions) Complete() RegistryOptions {
	if r.Concurrency == 0 {
		r.Concurrency = 10
	}
	return r
}

func NewToolsService(opts ...RegistryOptions) *Service {
	opt := complete.Complete(opts...)
	return &Service{
		servers:          make(map[string]map[string]*mcp.Client),
		roots:            opt.Roots,
		concurrency:      opt.Concurrency,
		oauthRedirectURL: opt.OAuthRedirectURL,
		callbackHandler:  opt.CallbackHandler,
	}
}

func (s *Service) SetSampler(sampler Sampler) {
	s.sampler = sampler
}

func (s *Service) GetDynamicInstruction(ctx context.Context, instruction types.DynamicInstructions) (string, error) {
	if !instruction.IsSet() {
		return "", nil
	}

	session := mcp.SessionFromContext(ctx)

	if !instruction.IsPrompt() {
		return expr.EvalString(ctx, session.EnvMap(), s.newGlobals(ctx, nil), instruction.Instructions)
	}

	prompt, err := s.GetPrompt(ctx, instruction.MCPServer, instruction.Prompt, envvar.ReplaceMap(session.EnvMap(), instruction.Args))
	if err != nil {
		return "", fmt.Errorf("failed to get prompt: %w", err)
	}
	if len(prompt.Messages) != 1 {
		return "", fmt.Errorf("prompt %s/%s returned %d messages, expected 1",
			instruction.MCPServer, instruction.Prompt, len(prompt.Messages))
	}
	return prompt.Messages[0].Content.Text, nil
}

func (s *Service) GetPrompt(ctx context.Context, target, prompt string, args map[string]string) (*mcp.GetPromptResult, error) {
	if target == "" && prompt != "" {
		target = prompt
	}

	config := types.ConfigFromContext(ctx)

	if inline, ok := config.Prompts[target]; ok && target == prompt {
		vals := map[string]any{}
		for k, v := range args {
			vals[k] = v
		}
		rendered, err := expr.EvalString(ctx, mcp.SessionFromContext(ctx).EnvMap(), s.newGlobals(ctx, vals), inline.Template)
		if err != nil {
			return nil, fmt.Errorf("failed to render inline prompt %s: %w", prompt, err)
		}
		return &mcp.GetPromptResult{
			Messages: []mcp.PromptMessage{
				{
					Role: "user",
					Content: mcp.Content{
						Type: "text",
						Text: rendered,
					},
				},
			},
		}, nil
	}

	c, err := s.GetClient(ctx, target)
	if err != nil {
		return nil, err
	}

	return c.GetPrompt(ctx, prompt, args)
}

func (s *Service) GetClient(ctx context.Context, name string) (*mcp.Client, error) {
	s.serverLock.Lock()
	defer s.serverLock.Unlock()

	session := mcp.SessionFromContext(ctx)
	if session == nil {
		return nil, fmt.Errorf("session not found in context")
	}

	config := types.ConfigFromContext(ctx)

	servers, ok := s.servers[strings.Split(session.ID(), "/")[0]]
	if !ok {
		servers = make(map[string]*mcp.Client)
		s.servers[session.ID()] = servers
	}

	c, ok := servers[name]
	if ok {
		return c, nil
	}

	var roots mcp.ListRootsResult
	if session.InitializeRequest.Capabilities.Roots != nil {
		err := session.Exchange(ctx, "roots/list", mcp.ListRootsRequest{}, &roots)
		if err != nil {
			return nil, fmt.Errorf("failed to list roots: %w", err)
		}
	}

	mcpConfig, ok := config.MCPServers[name]
	if !ok {
		return nil, fmt.Errorf("MCP server %s not found in config", name)
	}

	if len(s.roots) > 0 {
		roots.Roots = append(roots.Roots, s.roots...)
	}

	clientOpts := mcp.ClientOption{
		Roots:         roots.Roots,
		Env:           session.EnvMap(),
		ParentSession: session,
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
		Runner:           &s.runner,
		CallbackHandler:  s.callbackHandler,
		OAuthRedirectURL: s.oauthRedirectURL,
	}
	if session.InitializeRequest.Capabilities.Elicitation == nil {
		clientOpts.OnElicit = func(ctx context.Context, elicitation mcp.ElicitRequest) (result mcp.ElicitResult, _ error) {
			return mcp.ElicitResult{
				Action: "cancel",
			}, nil
		}
	} else {
		clientOpts.OnElicit = func(ctx context.Context, elicitation mcp.ElicitRequest) (result mcp.ElicitResult, _ error) {
			err := session.Exchange(ctx, "elicitation/create", elicitation, &result)
			return result, err
		}
	}
	if s.sampler != nil {
		clientOpts.OnSampling = func(ctx context.Context, samplingRequest mcp.CreateMessageRequest) (mcp.CreateMessageResult, error) {
			result, err := s.sampler.Sample(ctx, samplingRequest, sampling.SamplerOptions{
				ProgressToken: uuid.String(),
			})
			if err != nil {
				return mcp.CreateMessageResult{}, err
			}
			for _, content := range result.Content {
				return mcp.CreateMessageResult{
					Content:    content,
					Role:       "assistant",
					Model:      result.Model,
					StopReason: result.StopReason,
				}, nil
			}
			return mcp.CreateMessageResult{}, fmt.Errorf("no content returned from sampler")
		}
	}

	c, err := mcp.NewClient(ctx, name, mcpConfig, clientOpts)
	if err != nil {
		return nil, err
	}

	servers[name] = c
	s.servers[session.ID()] = servers
	return c, nil
}

func (s *Service) SampleCall(ctx context.Context, agent string, args any, opts ...SampleCallOptions) (*types.CallResult, error) {
	config := types.ConfigFromContext(ctx)
	createMessageRequest, err := s.convertToSampleRequest(config, agent, args)
	if err != nil {
		return nil, err
	}

	opt := complete.Complete(opts...)

	return s.sampler.Sample(ctx, *createMessageRequest, sampling.SamplerOptions{
		ProgressToken: opt.ProgressToken,
		AgentOverride: opt.AgentOverride,
	})
}

type CallOptions struct {
	ProgressToken any
	AgentOverride types.AgentCall
	LogData       map[string]any
	ReturnInput   bool
	ReturnOutput  bool
	Target        any
	ToolCall      *types.ToolCall
}

func (o CallOptions) Merge(other CallOptions) (result CallOptions) {
	result.ProgressToken = complete.Last(o.ProgressToken, other.ProgressToken)
	result.AgentOverride = complete.Merge(o.AgentOverride, other.AgentOverride)
	result.LogData = complete.MergeMap(o.LogData, other.LogData)
	result.ReturnInput = o.ReturnInput || other.ReturnInput
	result.ReturnOutput = o.ReturnOutput || other.ReturnOutput
	result.Target = complete.Last(o.Target, other.Target)
	result.ToolCall = complete.Last(o.ToolCall, other.ToolCall)
	return
}

func (s *Service) getTarget(ctx context.Context, config types.Config, server, tool string) (any, error) {
	if a, ok := config.Agents[server]; ok {
		return a, nil
	} else if f, ok := config.Flows[server]; ok {
		return f, nil
	}
	tools, err := s.ListTools(ctx, ListToolsOptions{
		Servers: []string{server},
		Tools:   []string{tool},
	})
	if err != nil {
		return nil, err
	}
	if len(tools) == 1 && len(tools[0].Tools) == 1 {
		return tools[0].Tools[0], nil
	}
	return nil, fmt.Errorf("unknown target %s/%s", server, tool)
}

func (s *Service) runAfter(ctx context.Context, config types.Config, target, server, tool string, ret *types.CallResult, opt CallOptions) (*types.CallResult, error) {
	var err error
	for _, flowName := range slices.Sorted(maps.Keys(config.Flows)) {
		if slices.Contains(config.Flows[flowName].After, target) ||
			slices.Contains(config.Flows[flowName].After, server) {
			newOpts := opt
			newOpts.ReturnOutput = true
			newOpts.ReturnInput = false
			newOpts.Target, err = s.getTarget(ctx, config, server, tool)
			if err != nil {
				return nil, err
			}

			newRet, err := s.Call(ctx, flowName, "", ret, newOpts)
			if err != nil {
				return nil, fmt.Errorf("failed to call after flow %s: %w", flowName, err)
			}
			if newRet.IsError {
				return newRet, nil
			}

			if len(newRet.Content) == 0 || newRet.Content[0].Text == "" {
				return nil, fmt.Errorf("after flow %s returned empty content", flowName)
			}

			if retType, ok := ret.Content[0].StructuredContent.(*types.CallResult); ok {
				ret = retType
			} else {
				var callResult types.CallResult
				if err := json.Unmarshal([]byte(ret.Content[0].Text), &callResult); err != nil {
					return nil, fmt.Errorf("failed to unmarshal call result from after flow %s: %w", flowName, err)
				}
				ret = &callResult
			}
		}
	}

	return ret, nil
}

func (s *Service) runBefore(ctx context.Context, config types.Config, target, server, tool string, args any, opt CallOptions) (any, *types.CallResult, error) {
	var err error

	for _, flowName := range slices.Sorted(maps.Keys(config.Flows)) {
		if slices.Contains(config.Flows[flowName].Before, target) ||
			slices.Contains(config.Flows[flowName].Before, server) {
			newOpts := opt
			newOpts.ReturnInput = true
			newOpts.ReturnOutput = false
			newOpts.Target, err = s.getTarget(ctx, config, server, tool)
			if err != nil {
				return nil, nil, err
			}
			ret, err := s.Call(ctx, flowName, "", args, newOpts)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to call before flow %s: %w", flowName, err)
			}
			if ret.IsError {
				return nil, ret, nil
			}
			args = ret.Content[0].StructuredContent
		}
	}

	return args, nil, nil
}

func (s *Service) Call(ctx context.Context, server, tool string, args any, opts ...CallOptions) (ret *types.CallResult, err error) {
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
	config := types.ConfigFromContext(ctx)

	target := server
	if tool != "" {
		target = server + "/" + tool
	}

	defer func() {
		if err == nil {
			ret, err = s.runAfter(ctx, config, target, server, tool, ret, opt)
		}
	}()

	targetType := "tool"
	if _, ok := config.Agents[server]; ok {
		targetType = "agent"
	} else if _, ok := config.Flows[server]; ok {
		targetType = "flow"
	}

	if session != nil && opt.ProgressToken != nil {
		var tc types.ToolCall
		if opt.ToolCall != nil {
			tc = *opt.ToolCall
		} else {
			tc.CallID = uuid.String()
			argsData, _ := json.Marshal(args)
			tc.Arguments = string(argsData)
			tc.Name, _ = opt.LogData["mcpToolName"].(string)
			if tc.Name == "" {
				tc.Name = target
			}
		}
		tc.Target = target
		tc.TargetType = targetType

		if opt.ToolCall == nil {
			_ = session.SendPayload(ctx, "notifications/progress", mcp.NotificationProgressRequest{
				ProgressToken: opt.ProgressToken,
				Meta: map[string]any{
					types.CompletionProgressMetaKey: types.CompletionProgress{
						HasMore: true,
						Item: types.CompletionItem{
							ID:       tc.CallID,
							ToolCall: &tc,
						},
					},
				},
			})
		}

		defer func() {
			tcResult := types.ToolCallResult{
				CallID: tc.CallID,
			}
			if ret != nil {
				tcResult.Output = *ret
			}
			if err != nil {
				tcResult.Output = types.CallResult{
					IsError: true,
					Content: []mcp.Content{
						{
							Type: "text",
							Text: err.Error(),
						},
					},
				}
			}
			_ = session.SendPayload(ctx, "notifications/progress", mcp.NotificationProgressRequest{
				ProgressToken: opt.ProgressToken,
				Meta: map[string]any{
					types.CompletionProgressMetaKey: types.CompletionProgress{
						Item: types.CompletionItem{
							ID:             tc.CallID,
							ToolCall:       &tc,
							ToolCallResult: &tcResult,
						},
					},
				},
			})
		}()
	}

	args, ret, err = s.runBefore(ctx, config, target, server, tool, args, opt)
	if err != nil || ret != nil {
		return ret, err
	}

	if _, ok := config.Agents[server]; ok {
		return s.SampleCall(ctx, server, args, SampleCallOptions{
			ProgressToken: opt.ProgressToken,
			AgentOverride: opt.AgentOverride,
		})
	}

	if _, ok := config.Flows[server]; ok {
		return s.startFlow(ctx, config, server, args, opt)
	}

	c, err := s.GetClient(ctx, server)
	if err != nil {
		return nil, err
	}

	mcpCallResult, err := c.Call(ctx, tool, args, mcp.CallOption{
		ProgressToken: opt.ProgressToken,
	})
	if err != nil {
		return nil, err
	}
	return &types.CallResult{
		Content: mcpCallResult.Content,
		IsError: mcpCallResult.IsError,
	}, nil
}

type ListToolsOptions struct {
	Servers []string
	Tools   []string
}

type ListToolsResult struct {
	Server string     `json:"server,omitempty"`
	Tools  []mcp.Tool `json:"tools,omitempty"`
}

func (s *Service) ListTools(ctx context.Context, opts ...ListToolsOptions) (result []ListToolsResult, _ error) {
	var (
		opt    ListToolsOptions
		config = types.ConfigFromContext(ctx)
	)
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

	serverList := slices.Sorted(maps.Keys(config.MCPServers))
	agentsList := slices.Sorted(maps.Keys(config.Agents))
	flowsList := slices.Sorted(maps.Keys(config.Flows))
	if len(opt.Servers) == 0 {
		opt.Servers = append(serverList, agentsList...)
		opt.Servers = append(opt.Servers, flowsList...)
	}

	for _, server := range opt.Servers {
		if !slices.Contains(serverList, server) {
			continue
		}

		c, err := s.GetClient(ctx, server)
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
		agent, ok := config.Agents[agentName]
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
		flow, ok := config.Flows[flowName]
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

func (s *Service) getMatches(ref string, tools []ListToolsResult) types.ToolMappings {
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

func (s *Service) GetEntryPoint(ctx context.Context, existingTools types.ToolMappings) (types.TargetMapping, error) {
	if tm, ok := existingTools[types.AgentTool]; ok {
		return tm, nil
	}

	config := types.ConfigFromContext(ctx)

	entrypoint := config.Publish.Entrypoint
	if entrypoint == "" {
		return types.TargetMapping{}, fmt.Errorf("no entrypoint specified")
	}

	tools, err := s.listToolsForReferences(ctx, []string{entrypoint})
	if err != nil {
		return types.TargetMapping{}, err
	}

	agents := s.getMatches(entrypoint, tools)
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

func (s *Service) listToolsForReferences(ctx context.Context, toolList []string) ([]ListToolsResult, error) {
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

	return s.ListTools(ctx, ListToolsOptions{
		Servers: servers,
	})
}

func (s *Service) BuildToolMappings(ctx context.Context, toolList []string) (types.ToolMappings, error) {
	tools, err := s.listToolsForReferences(ctx, toolList)
	if err != nil {
		return nil, err
	}

	result := types.ToolMappings{}
	for _, ref := range toolList {
		maps.Copy(result, s.getMatches(ref, tools))
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

func (s *Service) convertToSampleRequest(config types.Config, agent string, args any) (*mcp.CreateMessageRequest, error) {
	var (
		sampleArgs types.SampleCallRequest
	)
	switch args := args.(type) {
	case string:
		sampleArgs.Prompt = args
	case map[string]any:
		if hasOnlySampleKeys(args) {
			if err := types.JSONCoerce(args, &sampleArgs); err != nil {
				return nil, fmt.Errorf("failed to marshal args: %w", err)
			}
		} else {
			if err := types.JSONCoerce(args, &sampleArgs.Prompt); err != nil {
				return nil, fmt.Errorf("failed to marshal args to prompt: %w", err)
			}
		}
	case *types.SampleCallRequest:
		if args != nil {
			if err := types.JSONCoerce(*args, &sampleArgs); err != nil {
				return nil, fmt.Errorf("failed to marshal args to prompt: %w", err)
			}
		}
	default:
		if err := types.JSONCoerce(args, &sampleArgs); err != nil {
			return nil, fmt.Errorf("failed to marshal args to prompt: %w", err)
		}
	}

	var sampleRequest = mcp.CreateMessageRequest{
		MaxTokens: config.Agents[agent].MaxTokens,
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

type SampleCallOptions struct {
	ProgressToken any
	AgentOverride types.AgentCall
}

func (s SampleCallOptions) Merge(other SampleCallOptions) (result SampleCallOptions) {
	result.ProgressToken = complete.Last(s.ProgressToken, other.ProgressToken)
	result.AgentOverride = complete.Merge(s.AgentOverride, other.AgentOverride)
	return
}
