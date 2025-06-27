package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"

	"github.com/nanobot-ai/nanobot/pkg/expr"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/runtime"
	"github.com/nanobot-ai/nanobot/pkg/schema"
	"github.com/nanobot-ai/nanobot/pkg/tools"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

type Server struct {
	handlers       []handler
	DefaultRuntime *runtime.Runtime
}

const (
	toolMappingKey             = "toolMapping"
	currentAgentKey            = "currentAgent"
	promptMappingKey           = "promptMapping"
	resourceMappingKey         = "resourceMapping"
	resourceTemplateMappingKey = "resourceTemplateMapping"
)

func NewServer(runtime *runtime.Runtime) *Server {
	s := &Server{
		DefaultRuntime: runtime,
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
		handle[mcp.PingRequest]("ping", s.handlePing),
		handle[mcp.ListToolsRequest]("tools/list", s.handleListTools),
		handle[mcp.CallToolRequest]("tools/call", s.handleCallTool),
		handle[mcp.ListPromptsRequest]("prompts/list", s.handleListPrompts),
		handle[mcp.GetPromptRequest]("prompts/get", s.handleGetPrompt),
		handle[mcp.ListResourceTemplatesRequest]("resources/templates/list", s.handleListResourceTemplates),
		handle[mcp.ListResourcesRequest]("resources/list", s.handleListResources),
		handle[mcp.ReadResourceRequest]("resources/read", s.handleReadResource),
	}
}

func (s *Server) handleListResourceTemplates(ctx context.Context, msg mcp.Message, _ mcp.ListResourceTemplatesRequest) error {
	resourceTemplateMappings := types.ResourceTemplateMappings{}
	msg.Session.Get(resourceTemplateMappingKey, &resourceTemplateMappings)
	result := mcp.ListResourceTemplatesResult{
		ResourceTemplates: []mcp.ResourceTemplate{},
	}

	for _, k := range slices.Sorted(maps.Keys(resourceTemplateMappings)) {
		match := resourceTemplateMappings[k].Target.(templateMatch)
		result.ResourceTemplates = append(result.ResourceTemplates, match.resource)
	}

	return msg.Reply(ctx, result)
}

func (s *Server) matchResourceURITemplate(resourceTemplateMappings types.ResourceTemplateMappings, uri string) (types.TargetMapping, bool) {
	keys := slices.Sorted(maps.Keys(resourceTemplateMappings))
	for _, key := range keys {
		mapping := resourceTemplateMappings[key]
		match := mapping.Target.(templateMatch)
		if match.regexp.MatchString(uri) {
			mapping.TargetName = uri
			return mapping, true
		}
	}
	return types.TargetMapping{}, false
}

func (s *Server) handleReadResource(ctx context.Context, msg mcp.Message, payload mcp.ReadResourceRequest) error {
	resourceMappings := types.ResourceMappings{}
	msg.Session.Get(resourceMappingKey, &resourceMappings)
	resourceMapping, ok := resourceMappings[payload.URI]
	if !ok {
		resourceTemplateMappings := types.ResourceTemplateMappings{}
		msg.Session.Get(resourceTemplateMappingKey, &resourceTemplateMappings)
		resourceMapping, ok = s.matchResourceURITemplate(resourceTemplateMappings, payload.URI)
		if !ok {
			return fmt.Errorf("resource %s not found", payload.URI)
		}
	}

	r, err := s.getRuntime(msg.Session)
	if err != nil {
		return err
	}

	c, err := r.GetClient(ctx, resourceMapping.MCPServer)
	if err != nil {
		return fmt.Errorf("failed to get client for server %s: %w", resourceMapping.MCPServer, err)
	}

	result, err := c.ReadResource(ctx, resourceMapping.TargetName)
	if err != nil {
		return err
	}

	return msg.Reply(ctx, result)
}

func (s *Server) handleGetPrompt(ctx context.Context, msg mcp.Message, payload mcp.GetPromptRequest) error {
	promptMappings := types.PromptMappings{}
	msg.Session.Get(promptMappingKey, &promptMappings)
	promptMapping, ok := promptMappings[payload.Name]
	if !ok {
		return fmt.Errorf("prompt %s not found", payload.Name)
	}

	runtime, err := s.getRuntime(msg.Session)
	if err != nil {
		return err
	}

	result, err := runtime.GetPrompt(ctx, promptMapping.MCPServer, promptMapping.TargetName, payload.Arguments)
	if err != nil {
		return err
	}

	return msg.Reply(ctx, result)
}

func (s *Server) handleListResources(ctx context.Context, msg mcp.Message, _ mcp.ListResourcesRequest) error {
	resourceMappings := types.ResourceMappings{}
	msg.Session.Get(resourceMappingKey, &resourceMappings)
	result := mcp.ListResourcesResult{
		Resources: []mcp.Resource{},
	}

	for _, k := range slices.Sorted(maps.Keys(resourceMappings)) {
		result.Resources = append(result.Resources, resourceMappings[k].Target.(mcp.Resource))
	}

	return msg.Reply(ctx, result)
}

func (s *Server) handleListPrompts(ctx context.Context, msg mcp.Message, _ mcp.ListPromptsRequest) error {
	promptMappings := types.PromptMappings{}
	msg.Session.Get(promptMappingKey, &promptMappings)
	result := mcp.ListPromptsResult{
		Prompts: []mcp.Prompt{},
	}

	for _, k := range slices.Sorted(maps.Keys(promptMappings)) {
		result.Prompts = append(result.Prompts, promptMappings[k].Target.(mcp.Prompt))
	}

	return msg.Reply(ctx, result)
}

func (s *Server) handleCallTool(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) error {
	toolMappings := types.ToolMappings{}
	msg.Session.Get(toolMappingKey, &toolMappings)
	runtime, err := s.getRuntime(msg.Session)
	if err != nil {
		return err
	}

	toolMapping, ok := toolMappings[payload.Name]
	if !ok {
		return fmt.Errorf("tool %s not found", payload.Name)
	}

	if payload.Name == types.AgentTool {
		if currentAgent := ""; msg.Session.Get(currentAgentKey, &currentAgent) && currentAgent != "" {
			toolMapping = types.TargetMapping{
				MCPServer:  currentAgent,
				TargetName: currentAgent,
			}
		}
	}

	result, err := runtime.Call(ctx, toolMapping.MCPServer, toolMapping.TargetName, payload.Arguments, tools.CallOptions{
		ProgressToken: msg.ProgressToken(),
		LogData: map[string]any{
			"mcpToolName": payload.Name,
		},
	})
	if err != nil {
		return err
	}

	if payload.Name == types.AgentTool && result.ChatResponse && result.Agent != "" {
		msg.Session.Set(currentAgentKey, result.Agent)
	}

	mcpResult := mcp.CallToolResult{
		IsError: result.IsError,
		Content: result.Content,
	}

	return msg.Reply(ctx, mcpResult)
}

func (s *Server) handleListTools(ctx context.Context, msg mcp.Message, _ mcp.ListToolsRequest) error {
	result := mcp.ListToolsResult{
		Tools: []mcp.Tool{},
	}

	toolMappings := types.ToolMappings{}
	msg.Session.Get(toolMappingKey, &toolMappings)
	for _, k := range slices.Sorted(maps.Keys(toolMappings)) {
		if k == types.AgentTool {
			// entrypoint is essentially hidden
			continue
		}
		result.Tools = append(result.Tools, toolMappings[k].Target.(mcp.Tool))
	}

	return msg.Reply(ctx, result)
}

func (s *Server) handlePing(ctx context.Context, msg mcp.Message, _ mcp.PingRequest) error {
	return msg.Reply(ctx, mcp.PingResult{})
}

func (s *Server) buildResourceMappings(ctx context.Context) (types.ResourceMappings, error) {
	r, err := s.getRuntime(mcp.SessionFromContext(ctx))
	if err != nil {
		return nil, err
	}

	resourceMappings := types.ResourceMappings{}
	for _, ref := range append(r.GetConfig().Publish.Resources, r.GetConfig().Publish.MCPServers...) {
		toolRef := types.ParseToolRef(ref)
		if toolRef.Server == "" {
			continue
		}

		c, err := r.GetClient(ctx, toolRef.Server)
		if err != nil {
			return nil, err
		}
		resources, err := c.ListResources(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get resources for server %s: %w", toolRef, err)
		}

		for _, resource := range resources.Resources {
			resourceMappings[toolRef.PublishedName(resource.URI)] = types.TargetMapping{
				MCPServer:  toolRef.Server,
				TargetName: resource.URI,
				Target:     resource,
			}
		}
	}

	return resourceMappings, nil
}

func (s *Server) buildResourceTemplateMappings(ctx context.Context) (types.ResourceTemplateMappings, error) {
	resourceTemplateMappings := types.ResourceTemplateMappings{}
	r, err := s.getRuntime(mcp.SessionFromContext(ctx))
	if err != nil {
		return nil, err
	}

	for _, ref := range append(r.GetConfig().Publish.ResourceTemplates, r.GetConfig().Publish.MCPServers...) {
		toolRef := types.ParseToolRef(ref)
		if toolRef.Server == "" {
			continue
		}

		c, err := r.GetClient(ctx, toolRef.Server)
		if err != nil {
			return nil, err
		}
		resources, err := c.ListResourceTemplates(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get resources for server %s: %w", toolRef, err)
		}

		for _, resource := range resources.ResourceTemplates {
			re, err := uriToRegexp(resource.URITemplate)
			if err != nil {
				return nil, fmt.Errorf("failed to convert uri to regexp: %w", err)
			}
			resourceTemplateMappings[toolRef.PublishedName(resource.URITemplate)] = types.TargetMapping{
				MCPServer:  toolRef.Server,
				TargetName: resource.URITemplate,
				Target: templateMatch{
					regexp:   re,
					resource: resource,
				},
			}
		}
	}

	return resourceTemplateMappings, nil
}

type templateMatch struct {
	regexp   *regexp.Regexp
	resource mcp.ResourceTemplate
}

func (s *Server) buildPromptMappings(ctx context.Context) (types.PromptMappings, error) {
	serverPrompts := map[string]*mcp.ListPromptsResult{}
	session := mcp.SessionFromContext(ctx)
	r, err := s.getRuntime(session)
	if err != nil {
		return nil, err
	}

	result := types.PromptMappings{}
	c := r.GetConfig()
	for _, ref := range append(r.GetConfig().Publish.Prompts, r.GetConfig().Publish.MCPServers...) {
		toolRef := types.ParseToolRef(ref)
		if toolRef.Server == "" {
			continue
		}

		if inlinePrompt, ok := c.Prompts[toolRef.Server]; ok && toolRef.Tool == "" {
			result[toolRef.PublishedName(toolRef.Server)] = types.TargetMapping{
				MCPServer:  toolRef.Server,
				TargetName: toolRef.Server,
				Target:     inlinePrompt.ToPrompt(toolRef.PublishedName(toolRef.Server)),
			}
			continue
		}

		prompts, ok := serverPrompts[toolRef.Server]
		if !ok {
			c, err := r.GetClient(ctx, toolRef.Server)
			if err != nil {
				return nil, err
			}
			prompts, err = c.ListPrompts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get prompts for server %s: %w", toolRef, err)
			}
			serverPrompts[toolRef.Server] = prompts
		}

		for _, prompt := range prompts.Prompts {
			if prompt.Name == toolRef.Tool || toolRef.Tool == "" {
				prompt.Name = toolRef.PublishedName(prompt.Name)
				result[toolRef.PublishedName(prompt.Name)] = types.TargetMapping{
					MCPServer:  toolRef.Server,
					TargetName: prompt.Name,
					Target:     prompt,
				}
			}
		}
	}

	return result, nil
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
	envMap := session.EnvMap()
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

func (s *Server) getRuntime(session *mcp.Session) (*runtime.Runtime, error) {
	if session == nil {
		return nil, fmt.Errorf("session is nil")
	}

	var r runtime.Runtime
	if !session.Get(runtime.SessionKey, &r) {
		if s.DefaultRuntime == nil {
			return nil, fmt.Errorf("runtime not found in session")
		}
		return s.DefaultRuntime, nil
	}

	return &r, nil
}

func (s *Server) handleInitialize(ctx context.Context, msg mcp.Message, payload mcp.InitializeRequest) error {
	session := mcp.SessionFromContext(ctx)
	runtime, err := s.getRuntime(msg.Session)
	if err != nil {
		return fmt.Errorf("failed to get runtime from session: %w", err)
	}
	c := runtime.GetConfig()

	if err := reconcileEnv(session, c); err != nil {
		return err
	}

	toolMappings, err := runtime.BuildToolMappings(ctx, append(c.Publish.Tools, c.Publish.MCPServers...))
	if err != nil {
		return err
	}
	if runtime.GetConfig().Publish.Entrypoint != "" {
		toolMappings[types.AgentTool], err = runtime.GetEntryPoint(ctx, toolMappings)
		if err != nil {
			return err
		}
	} else if agent, ok := c.Agents["main"]; ok {
		toolMappings[types.AgentTool] = types.TargetMapping{
			MCPServer:  "main",
			TargetName: "main",
			Target:     agent,
		}
	}

	toolMappings = schema.ValidateToolMappings(toolMappings)
	msg.Session.Set(toolMappingKey, toolMappings)

	promptMappings, err := s.buildPromptMappings(ctx)
	if err != nil {
		return err
	}
	msg.Session.Set(promptMappingKey, promptMappings)

	resourceMappings, err := s.buildResourceMappings(ctx)
	if err != nil {
		return err
	}
	msg.Session.Set(resourceMappingKey, resourceMappings)

	resourceTemplateMappings, err := s.buildResourceTemplateMappings(ctx)
	if err != nil {
		return err
	}
	msg.Session.Set(resourceTemplateMappingKey, resourceTemplateMappings)

	var experimental map[string]any
	if c.Publish.Introduction.IsSet() {
		intro, err := runtime.GetDynamicInstruction(ctx, c.Publish.Introduction)
		if err != nil {
			return fmt.Errorf("failed to get introduction: %w", err)
		}
		experimental = map[string]any{
			"nanobot/intro": intro,
		}
	}

	return msg.Reply(ctx, mcp.InitializeResult{
		ProtocolVersion: payload.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Experimental: experimental,
			Logging:      &struct{}{},
			Prompts:      &mcp.PromptsServerCapability{},
			Resources:    &mcp.ResourcesServerCapability{},
			Tools:        &mcp.ToolsServerCapability{},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    c.Publish.Name,
			Version: c.Publish.Version,
		},
		Instructions: runtime.GetConfig().Publish.Instructions,
	})
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	for _, h := range s.handlers {
		ok, err := h(ctx, msg)
		if err != nil {
			msg.SendError(ctx, err)
		} else if ok {
			return
		}
	}

	msg.SendError(ctx, &mcp.RPCError{
		Code:    -32601,
		Message: fmt.Sprintf("method %q not allowed", msg.Method),
	})
}
