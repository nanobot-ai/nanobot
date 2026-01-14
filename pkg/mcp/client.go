package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/envvar"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	Session       *Session
	toolOverrides ToolOverrides
}

func (c *Client) Close(deleteSession bool) {
	if c.Session != nil {
		c.Session.Close(deleteSession)
	}
}

type SessionState struct {
	ID                string            `json:"id,omitempty"`
	InitializeResult  InitializeResult  `json:"initializeResult,omitzero"`
	InitializeRequest InitializeRequest `json:"initializeRequest,omitzero"`
	Attributes        map[string]any    `json:"attributes,omitempty"`
}

type ClientOption struct {
	HTTPClientOptions
	Roots         func(ctx context.Context) ([]Root, error)
	OnSampling    func(ctx context.Context, sampling CreateMessageRequest) (CreateMessageResult, error)
	OnElicit      func(ctx context.Context, msg Message, req ElicitRequest) (ElicitResult, error)
	OnRoots       func(ctx context.Context, msg Message) error
	OnLogging     func(ctx context.Context, logMsg LoggingMessage) error
	OnMessage     func(ctx context.Context, msg Message) error
	OnNotify      func(ctx context.Context, msg Message) error
	Env           map[string]string
	ParentSession *Session
	SessionState  *SessionState
	Runner        *Runner
	ClientName    string
	ClientVersion string
	Wire          Wire
	HookRunner    HookRunner
	ignoreEvents  bool
}

func (c ClientOption) Complete() ClientOption {
	if c.Runner == nil {
		c.Runner = &Runner{}
	}
	if c.ClientCredLookup == nil {
		c.ClientCredLookup = NewClientLookupFromEnv()
	}
	if c.TokenStorage == nil {
		c.TokenStorage = NewDefaultLocalStorage()
	}
	if c.OAuthClientName == "" {
		c.OAuthClientName = "Nanobot MCP Client"
	}
	if c.ClientName == "" {
		c.ClientName = "nanobot"
		c.ClientVersion = version.Get().String()
	} else {
		c.ClientName += fmt.Sprintf(" (via nanobot %s)", version.Get().String())
	}
	c.ignoreEvents = c.OnMessage == nil && c.OnNotify == nil && c.OnLogging == nil &&
		c.OnRoots == nil && c.OnSampling == nil && c.OnElicit == nil
	return c
}

func (c ClientOption) Merge(other ClientOption) (result ClientOption) {
	result.OnSampling = c.OnSampling
	if other.OnSampling != nil {
		result.OnSampling = other.OnSampling
	}
	result.OnRoots = c.OnRoots
	if other.OnRoots != nil {
		result.OnRoots = other.OnRoots
	}
	result.Roots = c.Roots
	if other.Roots != nil {
		result.Roots = other.Roots
	}
	result.OnLogging = c.OnLogging
	if other.OnLogging != nil {
		result.OnLogging = other.OnLogging
	}
	result.OnMessage = c.OnMessage
	if other.OnMessage != nil {
		result.OnMessage = other.OnMessage
	}
	result.OnNotify = c.OnNotify
	if other.OnNotify != nil {
		result.OnNotify = other.OnNotify
	}
	result.OnElicit = c.OnElicit
	if other.OnElicit != nil {
		result.OnElicit = other.OnElicit
	}
	result.CallbackHandler = complete.Last(c.CallbackHandler, other.CallbackHandler)
	result.ClientCredLookup = complete.Last(c.ClientCredLookup, other.ClientCredLookup)
	result.TokenStorage = complete.Last(c.TokenStorage, other.TokenStorage)
	result.ClientName = complete.Last(c.ClientName, other.ClientName)
	result.ClientVersion = complete.Last(c.ClientVersion, other.ClientVersion)
	result.OAuthRedirectURL = complete.Last(c.OAuthRedirectURL, other.OAuthRedirectURL)
	result.TokenExchangeEndpoint = complete.Last(c.TokenExchangeEndpoint, other.TokenExchangeEndpoint)
	result.TokenExchangeClientID = complete.Last(c.TokenExchangeClientID, other.TokenExchangeClientID)
	result.TokenExchangeClientSecret = complete.Last(c.TokenExchangeClientSecret, other.TokenExchangeClientSecret)
	result.OAuthClientName = complete.Last(c.OAuthClientName, other.OAuthClientName)
	result.Env = complete.MergeMap(c.Env, other.Env)
	result.SessionState = complete.Last(c.SessionState, other.SessionState)
	result.ParentSession = complete.Last(c.ParentSession, other.ParentSession)
	result.Runner = complete.Last(c.Runner, other.Runner)
	result.Wire = complete.Last(c.Wire, other.Wire)
	result.HookRunner = complete.Last(c.HookRunner, other.HookRunner)

	return result
}

type Server struct {
	Name        string `json:"name,omitempty"`
	ShortName   string `json:"shortName,omitempty"`
	Description string `json:"description,omitempty"`

	Image        string            `json:"image,omitempty"`
	Dockerfile   string            `json:"dockerfile,omitempty"`
	Source       ServerSource      `json:"source,omitzero"`
	Sandboxed    bool              `json:"sandboxed,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Command      string            `json:"command,omitempty"`
	Args         []string          `json:"args,omitempty"`
	BaseURL      string            `json:"url,omitempty"`
	Ports        []string          `json:"ports,omitempty"`
	ReversePorts []int             `json:"reversePorts,omitempty"`
	Cwd          string            `json:"cwd,omitempty"`
	Workdir      string            `json:"workdir,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`

	// If providing tool overrides, any tools not included will be implicitly disabled.
	// If providing no tool overrides, all tools will be enabled.
	ToolOverrides ToolOverrides `json:"toolOverrides,omitzero"`

	Hooks Hooks `json:"hooks,omitzero"`
}

func (s Server) MarshalJSON() ([]byte, error) {
	if s.Cwd == "." {
		s.Cwd = ""
	}
	type Alias Server
	return json.Marshal((Alias)(s))
}

type ToolOverrides map[string]ToolOverride

type ToolOverride struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	// The input schema is replaced if set here, and no translation is performed.
	// Therefore, whatever is replaced here needs to be understood by the MCP server.
	InputSchema json.RawMessage `json:"inputSchema,omitempty"`
}

type ServerSource struct {
	Repo      string `json:"repo,omitempty"`
	Tag       string `json:"tag,omitempty"`
	Commit    string `json:"commit,omitempty"`
	Branch    string `json:"branch,omitempty"`
	SubPath   string `json:"subPath,omitempty"`
	Reference string `json:"reference,omitempty"`
}

func (s *ServerSource) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		// If the data is a string, treat it as a repo URL
		var subPath string
		if err := json.Unmarshal(data, &subPath); err != nil {
			return fmt.Errorf("failed to unmarshal server source: %w", err)
		}
		s.SubPath = subPath
		return nil
	}
	type Alias ServerSource
	return json.Unmarshal(data, (*Alias)(s))
}

func toHandler(opts ClientOption) MessageHandler {
	return MessageHandlerFunc(func(ctx context.Context, msg Message) {
		if msg.Method == "sampling/createMessage" && opts.OnSampling != nil {
			var param CreateMessageRequest
			if err := json.Unmarshal(msg.Params, &param); err != nil {
				msg.SendError(ctx, fmt.Errorf("failed to unmarshal sampling/createMessage: %w", err))
				return
			}
			go func() {
				resp, err := opts.OnSampling(ctx, param)
				if err != nil {
					if errors.Is(err, ErrNoReader) {
						msg.SendError(ctx, ErrRPCMethodNotFound.RPCError().WithMessage("%s", msg.Method))
					} else {
						msg.SendError(ctx, fmt.Errorf("failed to handle sampling/createMessage: %w", err))
					}
					return
				}
				err = msg.Reply(ctx, resp)
				if err != nil {
					log.Errorf(ctx, "failed to reply to sampling/createMessage: %v", err)
				}
			}()
		} else if msg.Method == "elicitation/create" && opts.OnElicit != nil {
			var param ElicitRequest
			if err := json.Unmarshal(msg.Params, &param); err != nil {
				msg.SendError(ctx, fmt.Errorf("failed to unmarshal elicitation/create: %w", err))
				return
			}
			go func() {
				resp, err := opts.OnElicit(ctx, msg, param)
				if err != nil {
					if errors.Is(err, ErrNoReader) {
						msg.SendError(ctx, ErrRPCMethodNotFound.RPCError().WithMessage("%s", msg.Method))
					} else {
						msg.SendError(ctx, fmt.Errorf("failed to handle elicitation/create: %w", err))
					}
					return
				}
				// Give the client caller a way to handle the elicitation out of bounds
				if resp.Action != "handled" {
					err = msg.Reply(ctx, resp)
					if err != nil {
						log.Errorf(ctx, "failed to reply to elicitation/create: %v", err)
					}
				}
			}()
		} else if msg.Method == "roots/list" && opts.OnRoots != nil {
			go func() {
				if err := opts.OnRoots(ctx, msg); err != nil && !errors.Is(err, ErrNoReader) {
					msg.SendError(ctx, fmt.Errorf("failed to handle roots/list: %w", err))
				}
			}()
		} else if msg.Method == "notifications/message" && opts.OnLogging != nil {
			var param LoggingMessage
			if err := json.Unmarshal(msg.Params, &param); err != nil {
				msg.SendError(ctx, fmt.Errorf("failed to unmarshal notifications/message: %w", err))
				return
			}
			if err := opts.OnLogging(ctx, param); err != nil && !errors.Is(err, ErrNoReader) {
				msg.SendError(ctx, fmt.Errorf("failed to handle notifications/message: %w", err))
			}
		} else if strings.HasPrefix(msg.Method, "notifications/") && opts.OnNotify != nil {
			if err := opts.OnNotify(ctx, msg); err != nil && !errors.Is(err, ErrNoReader) {
				log.Errorf(ctx, "failed to handle notification: %v", err)
			}
		} else if opts.OnMessage != nil {
			if err := opts.OnMessage(ctx, msg); err != nil && !errors.Is(err, ErrNoReader) {
				log.Errorf(ctx, "failed to handle message: %v", err)
			}
		}
	})
}

func waitForURL(ctx context.Context, serverName, baseURL string) error {
	if baseURL == "" {
		return fmt.Errorf("base URL is empty for server %s", serverName)
	}

	for i := range 120 {
		if i%20 == 0 {
			log.Infof(ctx, "Waiting for server %s at %s to be ready...", serverName, baseURL)
		}
		resp, err := http.Get(baseURL)
		if err != nil {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled while waiting for server %s at %s: %w", serverName, baseURL, ctx.Err())
			case <-time.After(500 * time.Millisecond):
			}
		} else {
			_ = resp.Body.Close()
			log.Infof(ctx, "Server %s at %s is ready", serverName, baseURL)
			return nil
		}
	}

	return fmt.Errorf("server %s at %s did not respond within the timeout period", serverName, baseURL)
}

func NewSession(ctx context.Context, serverName string, config Server, opts ...ClientOption) (*Session, error) {
	var (
		wire Wire
		err  error
		opt  = complete.Complete(opts...)
	)

	if opt.Wire != nil {
		wire = opt.Wire
	} else if config.Command == "" && config.BaseURL == "" {
		return nil, fmt.Errorf("no command or base URL provided")
	} else if config.BaseURL != "" {
		if (opt.CallbackHandler != nil) != (opt.OAuthRedirectURL != "") {
			return nil, fmt.Errorf("must specify both or neither callback server and OAuth redirect URL")
		}

		if config.Command != "" {
			var err error
			config, err = opt.Runner.Run(ctx, opt.Roots, opt.Env, serverName, config)
			if err != nil {
				return nil, err
			}
			if err := waitForURL(ctx, serverName, config.BaseURL); err != nil {
				return nil, err
			}
		}
		headers := envvar.ReplaceMap(opt.Env, config.Headers)
		if opt.SessionState != nil && opt.SessionState.ID != "" {
			if headers == nil {
				headers = make(map[string]string)
			}
			headers["Mcp-Session-Id"] = opt.SessionState.ID
		}
		wire, err = newHTTPClient(serverName, config, opt.HTTPClientOptions, opt.SessionState, headers, !opt.ignoreEvents)
		if err != nil {
			return nil, err
		}
	} else {
		wire, err = newStdioClient(ctx, opt.Roots, opt.Env, serverName, config, opt.Runner)
		if err != nil {
			return nil, err
		}
	}

	return newSession(ctx, wire, toHandler(opt), opt.SessionState, opt.HookRunner, config.Hooks, opt.ParentSession)
}

func NewClient(ctx context.Context, serverName string, config Server, opts ...ClientOption) (*Client, error) {
	var (
		opt     = complete.Complete(opts...)
		session *Session
		err     error
	)

	session, err = NewSession(ctx, serverName, config, opt)
	if err != nil {
		return nil, err
	}

	abortCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-abortCtx.Done():
		case <-ctx.Done():
			// Abort the session if the ctx closes while creating the clients
			session.Close(false)
		}
	}()

	c := &Client{
		Session:       session,
		toolOverrides: config.ToolOverrides,
	}

	var (
		sampling     *SamplingCapability
		roots        *RootsCapability
		elicitations *struct{}
	)
	if opt.OnSampling != nil {
		sampling = &SamplingCapability{
			Context: &struct{}{},
			Tools:   &struct{}{},
		}
	}
	if opt.OnRoots != nil {
		roots = &RootsCapability{}
	}
	if opt.OnElicit != nil {
		elicitations = &struct{}{}
	}
	if opt.SessionState == nil {
		_, err = c.Initialize(ctx, InitializeRequest{
			ProtocolVersion: "2025-11-25",
			Capabilities: ClientCapabilities{
				Sampling:    sampling,
				Roots:       roots,
				Elicitation: elicitations,
			},
			ClientInfo: ClientInfo{
				Name:    opt.ClientName,
				Version: opt.ClientVersion,
			},
		})
		return c, err
	}

	return c, nil
}

// startSpan creates a new span for client operations and returns a context, span, and cleanup function
func (c *Client) startSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span, func(error, string, string)) {
	tracer := otel.Tracer("mcp.client")
	ctx, span := tracer.Start(ctx, "mcp.client."+operation,
		trace.WithAttributes(attrs...),
	)

	endSpan := func(err error, successMsg, errorMsg string) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, errorMsg)
		} else {
			span.SetStatus(codes.Ok, successMsg)
		}
		span.End()
	}

	return ctx, span, endSpan
}

func (c *Client) Initialize(ctx context.Context, param InitializeRequest) (result InitializeResult, err error) {
	ctx, _, endSpan := c.startSpan(ctx, "initialize",
		attribute.String("mcp.protocol_version", param.ProtocolVersion),
		attribute.String("mcp.client_name", param.ClientInfo.Name),
	)
	defer func() { endSpan(err, "initialized", "initialization failed") }()

	err = c.Session.Exchange(ctx, "initialize", param, &result)
	if err == nil {
		err = c.Session.Send(ctx, Message{
			Method: "notifications/initialized",
		})
	}
	return
}

func (c *Client) ReadResource(ctx context.Context, uri string) (result *ReadResourceResult, err error) {
	ctx, _, endSpan := c.startSpan(ctx, "read_resource",
		attribute.String("mcp.resource.uri", uri),
	)
	defer func() { endSpan(err, "resource read", "read resource failed") }()

	result = &ReadResourceResult{}
	err = c.Session.Exchange(ctx, "resources/read", ReadResourceRequest{
		URI: uri,
	}, result)
	return
}

func (c *Client) ListResourceTemplates(ctx context.Context) (result *ListResourceTemplatesResult, err error) {
	ctx, span, endSpan := c.startSpan(ctx, "list_resource_templates")
	defer func() { endSpan(err, "resource templates listed", "list resource templates failed") }()

	result = &ListResourceTemplatesResult{}
	if c.Session.InitializeResult.Capabilities.Resources == nil {
		return result, nil
	}
	err = c.Session.Exchange(ctx, "resources/templates/list", struct{}{}, result)
	if err == nil {
		span.SetAttributes(attribute.Int("mcp.resource_templates.count", len(result.ResourceTemplates)))
	}
	return
}

func (c *Client) ListResources(ctx context.Context) (result *ListResourcesResult, err error) {
	ctx, span, endSpan := c.startSpan(ctx, "list_resources")
	defer func() { endSpan(err, "resources listed", "list resources failed") }()

	result = &ListResourcesResult{}
	if c.Session.InitializeResult.Capabilities.Resources == nil {
		return result, nil
	}
	err = c.Session.Exchange(ctx, "resources/list", struct{}{}, result)
	if err == nil {
		span.SetAttributes(attribute.Int("mcp.resources.count", len(result.Resources)))
	}
	return
}

func (c *Client) SubscribeResource(ctx context.Context, uri string) (result *SubscribeResult, err error) {
	ctx, _, endSpan := c.startSpan(ctx, "subscribe_resource",
		attribute.String("mcp.resource.uri", uri),
	)
	defer func() { endSpan(err, "resource subscribed", "subscribe resource failed") }()

	result = &SubscribeResult{}
	err = c.Session.Exchange(ctx, "resources/subscribe", SubscribeRequest{URI: uri}, result)
	return
}

func (c *Client) UnsubscribeResource(ctx context.Context, uri string) (result *UnsubscribeResult, err error) {
	ctx, _, endSpan := c.startSpan(ctx, "unsubscribe_resource",
		attribute.String("mcp.resource.uri", uri),
	)
	defer func() { endSpan(err, "resource unsubscribed", "unsubscribe resource failed") }()

	result = &UnsubscribeResult{}
	err = c.Session.Exchange(ctx, "resources/unsubscribe", UnsubscribeRequest{URI: uri}, result)
	return
}

func (c *Client) ListPrompts(ctx context.Context) (result *ListPromptsResult, err error) {
	ctx, span, endSpan := c.startSpan(ctx, "list_prompts")
	defer func() { endSpan(err, "prompts listed", "list prompts failed") }()

	result = &ListPromptsResult{}
	if c.Session.InitializeResult.Capabilities.Prompts == nil {
		return result, nil
	}
	err = c.Session.Exchange(ctx, "prompts/list", struct{}{}, result)
	if err == nil {
		span.SetAttributes(attribute.Int("mcp.prompts.count", len(result.Prompts)))
	}
	return
}

func (c *Client) GetPrompt(ctx context.Context, name string, args map[string]string) (result *GetPromptResult, err error) {
	ctx, span, endSpan := c.startSpan(ctx, "get_prompt",
		attribute.String("mcp.prompt.name", name),
		attribute.Int("mcp.prompt.args_count", len(args)),
	)
	defer func() { endSpan(err, "prompt retrieved", "get prompt failed") }()

	result = &GetPromptResult{}
	err = c.Session.Exchange(ctx, "prompts/get", GetPromptRequest{
		Name:      name,
		Arguments: args,
	}, result)
	if err == nil {
		span.SetAttributes(attribute.Int("mcp.prompt.messages_count", len(result.Messages)))
	}
	return
}

func (c *Client) ListTools(ctx context.Context) (result *ListToolsResult, err error) {
	ctx, span, endSpan := c.startSpan(ctx, "list_tools")
	defer func() { endSpan(err, "tools listed", "list tools failed") }()

	result = &ListToolsResult{}
	if c.Session.InitializeResult.Capabilities.Tools == nil {
		return result, nil
	}

	err = c.Session.Exchange(ctx, "tools/list", struct{}{}, result)
	if err == nil && len(c.toolOverrides) > 0 {
		originalCount := len(result.Tools)
		filtered := result.Tools[:0] // reuse the backing array
		for _, tool := range result.Tools {
			override, ok := c.toolOverrides[tool.Name]
			if !ok {
				// If there are tool overrides, but this tool is not there, then skip it.
				continue
			}

			tool.Name = complete.First(override.Name, tool.Name)
			tool.Description = complete.First(override.Description, tool.Description)
			if len(override.InputSchema) > 0 {
				tool.InputSchema = override.InputSchema
			}

			filtered = append(filtered, tool)
		}
		result.Tools = filtered
		span.SetAttributes(
			attribute.Int("mcp.tools.original_count", originalCount),
			attribute.Int("mcp.tools.filtered_count", len(filtered)),
		)
	}

	if err == nil {
		span.SetAttributes(attribute.Int("mcp.tools.count", len(result.Tools)))
	}
	return
}

func (c *Client) Ping(ctx context.Context) (result *PingResult, err error) {
	ctx, _, endSpan := c.startSpan(ctx, "ping")
	defer func() { endSpan(err, "ping successful", "ping failed") }()

	result = &PingResult{}
	err = c.Session.Exchange(ctx, "ping", struct{}{}, result)
	return
}

type CallOption struct {
	ProgressToken any
	Meta          map[string]any
}

func (c CallOption) Merge(other CallOption) (result CallOption) {
	result.ProgressToken = complete.Last(c.ProgressToken, other.ProgressToken)
	result.Meta = complete.MergeMap(c.Meta, other.Meta)
	return
}

func (c *Client) Call(ctx context.Context, tool string, args any, opts ...CallOption) (result *CallToolResult, err error) {
	ctx, span, endSpan := c.startSpan(ctx, "call_tool",
		attribute.String("mcp.tool.name", tool),
	)
	defer func() {
		if err != nil {
			endSpan(err, "", "tool call failed")
		} else {
			span.SetAttributes(attribute.Bool("mcp.tool.is_error", result.IsError))
			if result.IsError {
				endSpan(nil, "", "tool returned error")
			} else {
				endSpan(nil, "tool call succeeded", "")
			}
		}
	}()

	opt := complete.Complete(opts...)
	result = new(CallToolResult)

	for name, o := range c.toolOverrides {
		if o.Name != "" && tool == o.Name {
			tool = name
			span.SetAttributes(attribute.String("mcp.tool.original_name", name))
			break
		}
	}

	err = c.Session.Exchange(ctx, "tools/call", struct {
		Name      string         `json:"name"`
		Arguments any            `json:"arguments,omitempty"`
		Meta      map[string]any `json:"_meta,omitempty"`
	}{
		Name:      tool,
		Arguments: args,
		Meta:      opt.Meta,
	}, result, ExchangeOption{
		ProgressToken: opt.ProgressToken,
	})

	return
}

func (c *Client) SetLogLevel(ctx context.Context, level string) (err error) {
	ctx, _, endSpan := c.startSpan(ctx, "set_log_level",
		attribute.String("mcp.log_level", level),
	)
	defer func() { endSpan(err, "log level set", "set log level failed") }()

	if c.Session.InitializeResult.Capabilities.Logging == nil {
		// Logging is not supported, don't error.
		return nil
	}

	err = c.Session.Exchange(ctx, "logging/setLevel", SetLogLevelRequest{
		Level: level,
	}, &SetLogLevelResult{})

	return
}
