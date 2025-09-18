package sessiondata

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/config"
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
	BuildToolMappings(ctx context.Context, toolList []string, opts ...types.BuildToolMappingsOptions) (types.ToolMappings, error)
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

func (d *Data) getEntrypoints(ctx context.Context) []string {
	c := types.ConfigFromContext(ctx)
	return c.Publish.Entrypoint
}

func (d *Data) SetCurrentAgent(ctx context.Context, newAgent string) error {
	if newAgent == d.CurrentAgent(ctx) {
		return nil
	}

	entrypoints := d.getEntrypoints(ctx)
	session := mcp.SessionFromContext(ctx)
	for session.Parent != nil {
		session = session.Parent
	}

	d.Refresh(ctx)
	if newAgent == "" {
		session.Delete(types.CurrentAgentSessionKey)
		return nil
	}

	if !slices.Contains(entrypoints, newAgent) {
		return fmt.Errorf("agent %s not found in entrypoints", newAgent)
	}

	session.Set(types.CurrentAgentSessionKey, mcp.SavedString(newAgent))
	return nil
}

func (d *Data) Agents(ctx context.Context) ([]types.AgentDisplay, error) {
	var (
		session = mcp.SessionFromContext(ctx)
		agents  []types.AgentDisplay
		c       types.Config
	)

	if found := session.Get(agentsSessionKey, &agents); found {
		return agents, nil
	}

	session.Get(types.ConfigSessionKey, &c)

	for _, key := range d.getEntrypoints(ctx) {
		var (
			agentDisplay types.AgentDisplay
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

			icon, _ := c.Session.InitializeResult.Capabilities.Experimental["ai.nanobot.meta/icon"].(string)
			iconDark, _ := c.Session.InitializeResult.Capabilities.Experimental["ai.nanobot.meta/icon"].(string)
			starterMessages, _ := c.Session.InitializeResult.Capabilities.Experimental["ai.nanobot.meta/starter-messages"].(string)

			agentDisplay = types.AgentDisplay{
				Name:            complete.First(c.Session.InitializeResult.ServerInfo.Name, mcpServer.Name, mcpServer.ShortName, key),
				ShortName:       complete.First(mcpServer.ShortName, c.Session.InitializeResult.ServerInfo.Name, mcpServer.Name, key),
				Description:     strings.TrimSpace(mcpServer.Description),
				Icon:            icon,
				IconDark:        iconDark,
				StarterMessages: strings.Split(starterMessages, ","),
			}
		} else {
			continue
		}

		agents = append(agents, agentDisplay)
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
		}
	}
	return currentAgent
}

func (d *Data) setURL(ctx context.Context) {
	var (
		session = mcp.SessionFromContext(ctx)
	)

	req := mcp.RequestFromContext(ctx)
	if req == nil {
		return
	}

	url := getHostURL(req)
	session.Set(types.PublicURLSessionKey, mcp.SavedString(url))
}

func getHostURL(req *http.Request) string {
	scheme := req.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if req.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := req.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = req.Host
	}

	if originalURL := req.Header.Get("X-Original-URL"); originalURL != "" {
		if strings.HasPrefix(originalURL, "http://") || strings.HasPrefix(originalURL, "https://") {
			return originalURL
		}
		return fmt.Sprintf("%s://%s%s", scheme, host, originalURL)
	}

	return fmt.Sprintf("%s://%s%s", scheme, host, req.URL.Path)
}

func (d *Data) getAndSetConfig(ctx context.Context, defaultConfig types.ConfigFactory) (types.Config, error) {
	var (
		c        types.Config
		nctx     = types.NanobotContext(ctx)
		session  = mcp.SessionFromContext(ctx)
		profiles string
		err      error
	)

	if len(nctx.Profile) > 0 {
		profiles = strings.Join(nctx.Profile, ",")
	} else if req := mcp.RequestFromContext(ctx); req != nil && strings.Contains(req.URL.Path, "/profile/") {
		_, v, ok := strings.Cut(req.URL.Path, "/profile/")
		if ok {
			profiles = strings.TrimSpace(v)
		}
	}

	if nctx.Config != nil {
		c, err = nctx.Config(ctx, profiles)
		if err != nil {
			return c, fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		c, err = defaultConfig(ctx, profiles)
		if err != nil {
			return c, fmt.Errorf("failed to load default config: %w", err)
		}
	}

	if req := mcp.RequestFromContext(ctx); req != nil && req.URL.Path == "/mcp/ui" {
		uiConfig, _, err := config.Load(ctx, "nanobot.ui")
		if err != nil {
			return c, fmt.Errorf("failed to load ui config: %w", err)
		}
		uiConfig.Publish.Entrypoint = nil
		c, err = config.Merge(c, *uiConfig)
		if err != nil {
			return c, fmt.Errorf("failed to merge ui config: %w", err)
		}
	}

	session.Set(types.ConfigSessionKey, &c)
	return c, nil
}

func initSubscriptions(session *mcp.Session) {
	var set bool
	if session.Get("_subscriptions_initialized", &set) {
		return
	}

	session.AddFilter(func(ctx context.Context, msg *mcp.Message) (*mcp.Message, error) {
		if msg.Method != "notifications/resources/updated" {
			return msg, nil
		}

		var uri string
		err := json.Unmarshal(msg.Params, &struct {
			URI *string `json:"uri"`
		}{
			URI: &uri,
		})
		if err != nil {
			return msg, nil
		}

		subs := resourceSubscriptions{}
		if session.Get(types.ResourceSubscriptionsSessionKey, &subs) {
			_, ok := subs[uri]
			if ok {
				return msg, nil
			}
		}

		return nil, nil
	})

	session.Set("_subscriptions_initialized", true)
}

type resourceSubscriptions map[string]struct{}

func (r resourceSubscriptions) Deserialize(v any) (any, error) {
	r = resourceSubscriptions{}
	return r, mcp.JSONCoerce(v, &r)
}

func (r resourceSubscriptions) Serialize() (any, error) {
	return (map[string]struct{})(r), nil
}

func (d *Data) UnsubscribeFromResources(ctx context.Context, uris ...string) error {
	var (
		session = mcp.SessionFromContext(ctx)
		subs    resourceSubscriptions
	)
	session.Get(types.ResourceSubscriptionsSessionKey, &subs)
	if subs == nil {
		subs = resourceSubscriptions{}
	}
	for _, uri := range uris {
		delete(subs, uri)
	}

	session.Set(types.ResourceSubscriptionsSessionKey, subs)
	return nil
}

func (d *Data) SubscribeToResources(ctx context.Context, uris ...string) error {
	var (
		session = mcp.SessionFromContext(ctx)
		subs    resourceSubscriptions
	)
	session.Get(types.ResourceSubscriptionsSessionKey, &subs)
	if subs == nil {
		subs = resourceSubscriptions{}
	}

	for _, uri := range uris {
		subs[uri] = struct{}{}
	}

	session.Set(types.ResourceSubscriptionsSessionKey, subs)
	return nil
}

func (d *Data) Sync(ctx context.Context, defaultConfig types.ConfigFactory) error {
	var (
		session      = mcp.SessionFromContext(ctx)
		existingHash string
		nctx         = types.NanobotContext(ctx)
	)

	session.Set(types.AccountIDSessionKey, nctx.User.ID)

	initSubscriptions(session)

	d.setURL(ctx)

	config, err := d.getAndSetConfig(ctx, defaultConfig)
	if err != nil {
		return err
	}

	session.Get(types.ConfigHashSessionKey, &existingHash)

	digest := sha256.New()
	_ = json.NewEncoder(digest).Encode(config)
	hash := fmt.Sprintf("%x", digest.Sum(nil))

	if hash != existingHash {
		d.Refresh(ctx)
	}

	session.Set(types.ConfigHashSessionKey, mcp.SavedString(hash))
	return nil
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

func (d *Data) getPublishedMCPServers(ctx context.Context) (result []string) {
	var (
		c       types.Config
		session = mcp.SessionFromContext(ctx)
	)
	session.Get(types.ConfigSessionKey, &c)

	if currentAgent := d.CurrentAgent(ctx); currentAgent != "" {
		result = append(result, currentAgent)
	}

	result = append(result, c.Publish.MCPServers...)
	return result
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

	toolMappings, err := d.runtime.BuildToolMappings(ctx, append(d.getPublishedMCPServers(ctx), c.Publish.Tools...))
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

func (d *Data) PublishedPromptMappings(ctx context.Context, opts ...GetOption) (types.PromptMappings, error) {
	var (
		prompts = types.PromptMappings{}
		session = mcp.SessionFromContext(ctx)
		c       = types.ConfigFromContext(ctx)
	)

	if found := session.Get(promptMappingKey, &prompts); !found && complete.Complete(opts...).AllowMissing {
		return nil, nil
	} else if found {
		return prompts, nil
	}

	promptMappings, err := d.BuildPromptMappings(ctx, append(d.getPublishedMCPServers(ctx), c.Publish.Prompts...)...)
	if err != nil {
		return nil, err
	}

	session.Set(promptMappingKey, promptMappings)
	return promptMappings, nil
}

func (d *Data) BuildPromptMappings(ctx context.Context, refs ...string) (types.PromptMappings, error) {
	var (
		serverPrompts = map[string]*mcp.ListPromptsResult{}
		result        = types.PromptMappings{}
		c             = types.ConfigFromContext(ctx)
	)

	for _, ref := range refs {
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
	for _, ref := range append(d.getPublishedMCPServers(ctx), config.Publish.Resources...) {
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
	for _, ref := range append(d.getPublishedMCPServers(ctx), config.Publish.ResourceTemplates...) {
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
