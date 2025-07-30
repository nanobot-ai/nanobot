package sessiondata

import (
	"context"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/schema"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

const (
	toolMappingKey               = "toolMapping"
	promptMappingKey             = "promptMapping"
	resourceMappingKey           = "resourceMapping"
	resourceTemplateMappingKey   = "resourceTemplateMapping"
	agentsSessionKey             = "agents"
	currentAgentTargetSessionKey = "currentAgentTargetMapping"
)

type Data struct {
	runtime RuntimeMeta
}

func NewData(runtime RuntimeMeta) *Data {
	return &Data{
		runtime: runtime,
	}
}

type RuntimeMeta interface {
	BuildToolMappings(ctx context.Context, toolList []string) (types.ToolMappings, error)
	GetClient(ctx context.Context, name string) (*mcp.Client, error)
}

type GetOption struct {
	AllowMissing bool
}

func (g GetOption) Merge(other GetOption) (result GetOption) {
	result.AllowMissing = complete.Last(g.AllowMissing, other.AllowMissing)
	return
}

func WithAllowMissing() GetOption {
	return GetOption{
		AllowMissing: true,
	}
}

func (d *Data) SetCurrentAgent(ctx context.Context, currentAgent string) error {
	session := mcp.SessionFromContext(ctx)
	for session.Parent != nil {
		session = session.Parent
	}
	d.Refresh(ctx)
	if currentAgent == "" {
		session.Delete(types.CurrentAgentSessionKey)
	} else {
		mappings, err := d.getEntrypointMapping(ctx)
		if err != nil {
			return fmt.Errorf("failed to build tool mappings: %w", err)
		}
		if _, ok := mappings[currentAgent]; !ok {
			return fmt.Errorf("current agent %s not found in tool mappings", currentAgent)
		}
		session.Set(types.CurrentAgentSessionKey, mcp.SavedString(currentAgent))
	}
	return nil
}

func (d *Data) GetCurrentAgentTargetMapping(ctx context.Context) (string, string, error) {
	var target types.TargetMapping[mcp.Tool]
	session := mcp.SessionFromContext(ctx)
	if found := session.Get(currentAgentTargetSessionKey, &target); found {
		return target.MCPServer, target.TargetName, nil
	}

	mappings, err := d.getEntrypointMapping(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to build tool mappings: %w", err)
	}

	currentAgent := d.CurrentAgent(ctx)

	target, ok := mappings[currentAgent]
	if !ok {
		return "", "", fmt.Errorf("current agent %s not found in tool mappings", currentAgent)
	}

	session.Set(currentAgentTargetSessionKey, &target)
	return target.MCPServer, target.TargetName, nil
}

func (d *Data) getEntrypointMapping(ctx context.Context) (types.ToolMappings, error) {
	var (
		session = mcp.SessionFromContext(ctx)
		c       types.Config
	)

	session.Get(types.ConfigSessionKey, &c)
	entrypoints := c.Publish.Entrypoint
	if _, ok := c.Agents[types.DefaultAgentName]; ok && len(entrypoints) == 0 {
		entrypoints = []string{types.DefaultAgentName}
	}

	return d.runtime.BuildToolMappings(ctx, entrypoints)
}

func (d *Data) Agents(ctx context.Context) (types.Agents, error) {
	var (
		session = mcp.SessionFromContext(ctx)
		agents  = types.Agents{}
		c       types.Config
	)

	if found := session.Get(agentsSessionKey, &agents); found {
		return agents, nil
	}

	session.Get(types.ConfigSessionKey, &c)

	mapping, err := d.getEntrypointMapping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build tool mappings: %w", err)
	}

	for agentKey, target := range mapping {
		var (
			agentDisplay types.AgentDisplay
			key          = target.MCPServer
		)

		if agent, ok := c.Agents[key]; ok {
			agentDisplay = agent.ToDisplay()
			agentDisplay.Name = complete.First(agentDisplay.Name, agentDisplay.ShortName, key)
		} else if mcpServer, ok := c.MCPServers[key]; ok {
			c, err := d.runtime.GetClient(ctx, key)
			if err != nil {
				return agents, err
			}

			name := c.Session.InitializeResult.ServerInfo.Name
			if name == "" {
				name = key
			}

			agentDisplay = types.AgentDisplay{
				Name:        complete.First(c.Session.InitializeResult.ServerInfo.Name, mcpServer.Name, mcpServer.ShortName, key),
				ShortName:   complete.First(mcpServer.ShortName, c.Session.InitializeResult.ServerInfo.Name, mcpServer.Name, key),
				Description: target.Target.Description,
			}
		} else {
			continue
		}
		agents[agentKey] = agentDisplay
	}

	session.Set(agentsSessionKey, &agents)
	return agents, nil
}

func (d *Data) CurrentAgent(ctx context.Context) string {
	var (
		session      = mcp.SessionFromContext(ctx)
		currentAgent string
		c            types.Config
	)
	if !session.Get(types.CurrentAgentSessionKey, &currentAgent) {
		session.Get(types.ConfigSessionKey, &c)
		if len(c.Publish.Entrypoint) > 0 {
			currentAgent = c.Publish.Entrypoint[0]
		} else if _, ok := c.Agents[types.DefaultAgentName]; ok {
			currentAgent = types.DefaultAgentName
		}
	}
	return currentAgent
}

func (d *Data) Refresh(ctx context.Context) {
	session := mcp.SessionFromContext(ctx)
	session.Delete(toolMappingKey)
	session.Delete(types.CurrentAgentSessionKey)
	session.Delete(promptMappingKey)
	session.Delete(resourceMappingKey)
	session.Delete(resourceTemplateMappingKey)
	session.Delete(agentsSessionKey)
	session.Delete(currentAgentTargetSessionKey)
}

func (d *Data) ToolMapping(ctx context.Context, opts ...GetOption) (types.ToolMappings, error) {
	var (
		session      = mcp.SessionFromContext(ctx)
		toolMappings = types.ToolMappings{}
	)

	if found := session.Get(toolMappingKey, &toolMappings); !found && complete.Complete(opts...).AllowMissing {
		return nil, nil
	} else if found {
		return toolMappings, nil
	}

	var c types.Config
	session.Get(types.ConfigSessionKey, &c)

	toolMappings, err := d.runtime.BuildToolMappings(ctx, append(c.Publish.Tools, c.Publish.MCPServers...))
	if err != nil {
		return nil, err
	}

	toolMappings = schema.ValidateToolMappings(toolMappings)
	session.Set(toolMappingKey, toolMappings)

	return toolMappings, nil
}

func (d *Data) ResourceTemplateMappings(ctx context.Context, opts ...GetOption) (types.ResourceTemplateMappings, error) {
	var (
		resourceTemplates = types.ResourceTemplateMappings{}
		session           = mcp.SessionFromContext(ctx)
		c                 types.Config
	)

	if found := session.Get(resourceTemplateMappingKey, &resourceTemplates); !found && complete.Complete(opts...).AllowMissing {
		return nil, nil
	} else if found {
		return resourceTemplates, nil
	}

	session.Get(types.ConfigSessionKey, &c)

	resourceTemplateMappings, err := d.buildResourceTemplateMappings(ctx, c)
	if err != nil {
		return nil, err
	}
	session.Set(resourceTemplateMappingKey, resourceTemplateMappings)

	return resourceTemplateMappings, nil
}

func (d *Data) ResourceMappings(ctx context.Context, opts ...GetOption) (types.ResourceMappings, error) {
	var (
		session = mcp.SessionFromContext(ctx)
		c       types.Config
	)

	session.Get(types.ConfigSessionKey, &c)

	resourceMappings, err := d.buildResourceMappings(ctx, c)
	if err != nil {
		return nil, err
	}

	return resourceMappings, nil
}

func (d *Data) PromptMappings(ctx context.Context, opts ...GetOption) (types.PromptMappings, error) {
	var (
		prompts = types.PromptMappings{}
		session = mcp.SessionFromContext(ctx)
	)

	if found := session.Get(promptMappingKey, &prompts); !found && complete.Complete(opts...).AllowMissing {
		return nil, nil
	} else if found {
		return prompts, nil
	}

	promptMappings, err := d.buildPromptMappings(ctx)
	if err != nil {
		return nil, err
	}

	session.Set(promptMappingKey, promptMappings)
	return promptMappings, nil
}

func (d *Data) buildPromptMappings(ctx context.Context) (types.PromptMappings, error) {
	var (
		serverPrompts = map[string]*mcp.ListPromptsResult{}
		result        = types.PromptMappings{}
		c             = types.ConfigFromContext(ctx)
	)

	for _, ref := range append(c.Publish.Prompts, c.Publish.MCPServers...) {
		toolRef := types.ParseToolRef(ref)
		if toolRef.Server == "" {
			continue
		}

		if inlinePrompt, ok := c.Prompts[toolRef.Server]; ok && toolRef.Tool == "" {
			result[toolRef.PublishedName(toolRef.Server)] = types.TargetMapping[mcp.Prompt]{
				MCPServer:  toolRef.Server,
				TargetName: toolRef.Server,
				Target:     inlinePrompt.ToPrompt(toolRef.PublishedName(toolRef.Server)),
			}
			continue
		}

		prompts, ok := serverPrompts[toolRef.Server]
		if !ok {
			c, err := d.runtime.GetClient(ctx, toolRef.Server)
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
				result[toolRef.PublishedName(prompt.Name)] = types.TargetMapping[mcp.Prompt]{
					MCPServer:  toolRef.Server,
					TargetName: prompt.Name,
					Target:     prompt,
				}
			}
		}
	}

	return result, nil
}

func (d *Data) buildResourceMappings(ctx context.Context, config types.Config) (types.ResourceMappings, error) {
	resourceMappings := types.ResourceMappings{}
	for _, ref := range append(config.Publish.Resources, config.Publish.MCPServers...) {
		toolRef := types.ParseToolRef(ref)
		if toolRef.Server == "" {
			continue
		}

		c, err := d.runtime.GetClient(ctx, toolRef.Server)
		if err != nil {
			return nil, err
		}
		resources, err := c.ListResources(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get resources for server %s: %w", toolRef, err)
		}

		for _, resource := range resources.Resources {
			resourceMappings[toolRef.PublishedName(resource.URI)] = types.TargetMapping[mcp.Resource]{
				MCPServer:  toolRef.Server,
				TargetName: resource.URI,
				Target:     resource,
			}
		}
	}

	return resourceMappings, nil
}

func (d *Data) buildResourceTemplateMappings(ctx context.Context, config types.Config) (types.ResourceTemplateMappings, error) {
	resourceTemplateMappings := types.ResourceTemplateMappings{}
	for _, ref := range append(config.Publish.ResourceTemplates, config.Publish.MCPServers...) {
		toolRef := types.ParseToolRef(ref)
		if toolRef.Server == "" {
			continue
		}

		c, err := d.runtime.GetClient(ctx, toolRef.Server)
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
			resourceTemplateMappings[toolRef.PublishedName(resource.URITemplate)] = types.TargetMapping[types.TemplateMatch]{
				MCPServer:  toolRef.Server,
				TargetName: resource.URITemplate,
				Target: types.TemplateMatch{
					Regexp:           re,
					ResourceTemplate: resource,
				},
			}
		}
	}

	return resourceTemplateMappings, nil
}
