package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/nanobot-ai/nanobot/pkg/expr"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/runtime"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/sessiondata"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type Server struct {
	handlers []handler
	runtime  *runtime.Runtime
	data     *sessiondata.Data
	config   types.ConfigFactory
	manager  *session.Manager
}

func NewServer(runtime *runtime.Runtime, config types.ConfigFactory, manager *session.Manager) *Server {
	s := &Server{
		runtime: runtime,
		data:    sessiondata.NewData(runtime),
		config:  config,
		manager: manager,
	}
	s.init()
	return s
}

type handler func(ctx context.Context, msg mcp.Message) (bool, error)

func handle[T any](method string, handler func(ctx context.Context, req mcp.Message, payload T) error) handler {
	return func(ctx context.Context, msg mcp.Message) (bool, error) {
		if msg.Method != method {
			return false, nil
		}
		var payload T
		if len(msg.Params) > 0 && !bytes.Equal(msg.Params, []byte("null")) {
			if err := json.Unmarshal(msg.Params, &payload); err != nil {
				return false, err
			}
		}
		return true, handler(ctx, msg, payload)
	}
}

func (s *Server) init() {
	s.handlers = []handler{
		handle[mcp.InitializeRequest]("initialize", s.handleInitialize),
		handle[mcp.Notification]("notifications/initialized", s.handleInitialized),
		handle[mcp.PingRequest]("ping", s.handlePing),
		handle[mcp.ListToolsRequest]("tools/list", s.handleListTools),
		handle[mcp.CallToolRequest]("tools/call", s.handleCallTool),
		handle[mcp.ListPromptsRequest]("prompts/list", s.handleListPrompts),
		handle[mcp.GetPromptRequest]("prompts/get", s.handleGetPrompt),
		handle[mcp.ListResourceTemplatesRequest]("resources/templates/list", s.handleListResourceTemplates),
		handle[mcp.ListResourcesRequest]("resources/list", s.handleListResources),
		handle[mcp.ReadResourceRequest]("resources/read", s.handleReadResource),
		handle[mcp.SubscribeRequest]("resources/subscribe", s.handleResourcesSubscribe),
		handle[mcp.UnsubscribeRequest]("resources/unsubscribe", s.handleResourcesUnsubscribe),
	}
}

func (s *Server) handleResourcesUnsubscribe(ctx context.Context, msg mcp.Message, payload mcp.UnsubscribeRequest) error {
	err := s.data.UnsubscribeFromResources(ctx, payload.URI)
	if err != nil {
		return err
	}
	return msg.Reply(ctx, map[string]any{})
}

func (s *Server) handleResourcesSubscribe(ctx context.Context, msg mcp.Message, payload mcp.SubscribeRequest) error {
	err := s.data.SubscribeToResources(ctx, payload.URI)
	if err != nil {
		return err
	}
	return msg.Reply(ctx, map[string]any{})
}

func (s *Server) handleListResourceTemplates(ctx context.Context, msg mcp.Message, _ mcp.ListResourceTemplatesRequest) error {
	resourceTemplateMappings, err := s.data.PublishedResourceTemplateMappings(ctx)
	if err != nil {
		return err
	}

	result := mcp.ListResourceTemplatesResult{
		ResourceTemplates: []mcp.ResourceTemplate{},
	}

	for _, k := range slices.Sorted(maps.Keys(resourceTemplateMappings)) {
		match := resourceTemplateMappings[k].Target
		result.ResourceTemplates = append(result.ResourceTemplates, match.ResourceTemplate)
	}

	return msg.Reply(ctx, result)
}

func (s *Server) handleReadResource(ctx context.Context, msg mcp.Message, payload mcp.ReadResourceRequest) error {
	target, resourceName, err := s.data.MatchPublishedResource(ctx, payload.URI)
	if err != nil {
		return fmt.Errorf("failed to read resource %s: %v", payload.URI, err)
	}

	c, err := s.runtime.GetClient(ctx, target)
	if err != nil {
		return fmt.Errorf("failed to get client for server %s: %w", target, err)
	}

	result, err := c.ReadResource(ctx, resourceName)
	if err != nil {
		return err
	}

	return msg.Reply(ctx, result)
}

func (s *Server) handleGetPrompt(ctx context.Context, msg mcp.Message, payload mcp.GetPromptRequest) error {
	promptMappings, err := s.data.PublishedPromptMappings(ctx)
	if err != nil {
		return err
	}

	promptMapping, ok := promptMappings[payload.Name]
	if !ok {
		return fmt.Errorf("prompt %s not found", payload.Name)
	}

	result, err := s.runtime.GetPrompt(ctx, promptMapping.MCPServer, promptMapping.TargetName, payload.Arguments)
	if err != nil {
		return err
	}

	return msg.Reply(ctx, result)
}

func (s *Server) handleListResources(ctx context.Context, msg mcp.Message, _ mcp.ListResourcesRequest) error {
	resourceMappings, err := s.data.PublishedResourceMappings(ctx)
	if err != nil {
		return err
	}

	result := mcp.ListResourcesResult{
		Resources: []mcp.Resource{},
	}

	for _, k := range slices.Sorted(maps.Keys(resourceMappings)) {
		result.Resources = append(result.Resources, resourceMappings[k].Target)
	}

	return msg.Reply(ctx, result)
}

func (s *Server) handleListPrompts(ctx context.Context, msg mcp.Message, _ mcp.ListPromptsRequest) error {
	s.data.Refresh(ctx)
	promptMappings, err := s.data.PublishedPromptMappings(ctx)
	if err != nil {
		return err
	}

	result := mcp.ListPromptsResult{
		Prompts: []mcp.Prompt{},
	}

	for _, k := range slices.Sorted(maps.Keys(promptMappings)) {
		result.Prompts = append(result.Prompts, promptMappings[k].Target)
	}

	return msg.Reply(ctx, result)
}

func (s *Server) handleCallTool(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) error {
	toolMappings, err := s.data.ToolMapping(ctx)
	if err != nil {
		return err
	}

	toolMapping, ok := toolMappings[payload.Name]
	if !ok {
		s.data.Refresh(ctx)
		toolMappings, err = s.data.ToolMapping(ctx)
		if err != nil {
			return err
		}
		toolMapping, ok = toolMappings[payload.Name]
		if !ok {
			return fmt.Errorf("tool %s not found", payload.Name)
		}
	}

	result, err := s.runtime.Call(ctx, toolMapping.MCPServer, toolMapping.TargetName, payload.Arguments, tools.CallOptions{
		ProgressToken: msg.ProgressToken(),
		LogData: map[string]any{
			"mcpToolName": payload.Name,
		},
		Meta: msg.Meta(),
	})
	if err != nil {
		return err
	}

	mcpResult := mcp.CallToolResult{
		StructuredContent: result.StructuredContent,
		IsError:           result.IsError,
		Content:           result.Content,
	}

	return msg.Reply(ctx, mcpResult)
}

func (s *Server) handleListTools(ctx context.Context, msg mcp.Message, _ mcp.ListToolsRequest) error {
	result := mcp.ListToolsResult{
		Tools: []mcp.Tool{},
	}

	toolMappings, err := s.data.ToolMapping(ctx)
	if err != nil {
		return err
	}

	for _, k := range slices.Sorted(maps.Keys(toolMappings)) {
		result.Tools = append(result.Tools, toolMappings[k].Target)
	}

	return msg.Reply(ctx, result)
}

func (s *Server) handlePing(ctx context.Context, msg mcp.Message, _ mcp.PingRequest) error {
	return msg.Reply(ctx, mcp.PingResult{})
}

func getEnvVal(envMap map[string]string, envKey string, envDef types.EnvDef) string {
	val, ok := expr.Lookup(envMap, envKey)
	if ok {
		return val
	}

	if envDef.UseBearerToken {
		bearer, ok := envMap["http:bearer-token"]
		if ok && bearer != "" {
			return bearer
		}
	}

	if !envDef.Optional {
		return ""
	}

	return envDef.Default
}

func reconcileEnv(session *mcp.Session, c types.Config) error {
	envMap := session.GetEnvMap()
	var missing []string
	for envKey, envDef := range c.Env {
		envVal := getEnvVal(envMap, envKey, envDef)
		if envVal == "" && !envDef.Optional {
			missing = append(missing, envKey)
			continue
		}
		envMap[envKey] = envVal
	}

	if len(missing) == 0 {
		return nil
	}

	var missingEnv []any
	for _, key := range missing {
		values := map[string]any{
			"name":        key,
			"description": c.Env[key].Description,
			"default":     c.Env[key].Default,
		}
		if len(c.Env[key].Options) > 0 {
			values["options"] = c.Env[key].Options
		}
		missingEnv = append(missingEnv, values)
	}
	return &mcp.RPCError{
		Code:    -32602,
		Message: fmt.Sprintf("missing required environment variables: %v", missing),
		DataObject: map[string]any{
			"missingEnv": missingEnv,
		},
	}
}

func (s *Server) handleInitialized(ctx context.Context, msg mcp.Message, payload mcp.Notification) error {
	return nil
}

func (s *Server) handleInitialize(ctx context.Context, msg mcp.Message, payload mcp.InitializeRequest) error {
	session := mcp.SessionFromContext(ctx)
	c := types.ConfigFromContext(ctx)

	if err := reconcileEnv(session, c); err != nil {
		return err
	}

	s.data.Refresh(ctx)

	var experimental map[string]any
	if c.Publish.Introduction.IsSet() {
		intro, err := s.runtime.GetDynamicInstruction(ctx, c.Publish.Introduction)
		if err != nil {
			return fmt.Errorf("failed to get introduction: %w", err)
		}
		experimental = map[string]any{
			"ai.nanobot.meta/intro": intro,
		}
	}

	return msg.Reply(ctx, mcp.InitializeResult{
		ProtocolVersion: payload.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Experimental: experimental,
			//Logging:      &struct{}{},
			Prompts: &mcp.PromptsServerCapability{},
			Resources: &mcp.ResourcesServerCapability{
				Subscribe: true,
			},
			Tools: &mcp.ToolsServerCapability{},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    c.Publish.Name,
			Version: c.Publish.Version,
		},
		Instructions: c.Publish.Instructions,
	})
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	if err := s.data.Sync(ctx, s.config); err != nil {
		msg.SendError(ctx, err)
		return
	}

	mcp.SessionFromContext(ctx).Set(session.ManagerSessionKey, s.manager)

	for _, h := range s.handlers {
		ok, err := h(ctx, msg)
		if err != nil {
			msg.SendError(ctx, err)
			return
		} else if ok {
			return
		}
	}

	msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage("%s", msg.Method))
}
