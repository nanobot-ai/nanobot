package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/envvar"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Client struct {
	Session *Session
}

type ClientOption struct {
	Roots         []Root
	OnSampling    func(ctx context.Context, sampling CreateMessageRequest) (CreateMessageResult, error)
	OnRoots       func(ctx context.Context, msg Message) error
	OnLogging     func(ctx context.Context, logMsg LoggingMessage) error
	OnMessage     func(ctx context.Context, msg Message) error
	OnNotify      func(ctx context.Context, msg Message) error
	Env           map[string]string
	ParentSession *Session
	SessionID     string
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
	result.Env = complete.MergeMap(c.Env, other.Env)
	result.SessionID = complete.Last(c.SessionID, other.SessionID)
	result.ParentSession = complete.Last(c.ParentSession, other.ParentSession)
	result.Roots = append(c.Roots, other.Roots...)
	return result
}

type Server struct {
	Image        string            `json:"image,omitempty"`
	Dockerfile   string            `json:"dockerfile,omitempty"`
	Source       ServerSource      `json:"source,omitempty"`
	Unsandboxed  bool              `json:"unsandboxed,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Command      string            `json:"command,omitempty"`
	Args         []string          `json:"args,omitempty"`
	BaseURL      string            `json:"url,omitempty"`
	Ports        []string          `json:"ports,omitempty"`
	ReversePorts []int             `json:"reversePorts"`
	Workdir      string            `json:"workdir,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
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
					msg.SendError(ctx, fmt.Errorf("failed to handle sampling/createMessage: %w", err))
					return
				}
				err = msg.Reply(ctx, resp)
				if err != nil {
					log.Errorf(ctx, "failed to reply to sampling/createMessage: %v", err)
				}
			}()
		} else if msg.Method == "roots/list" && opts.OnRoots != nil {
			go func() {
				if err := opts.OnRoots(ctx, msg); err != nil {
					msg.SendError(ctx, fmt.Errorf("failed to handle roots/list: %w", err))
					return
				}
			}()
		} else if msg.Method == "notifications/message" && opts.OnLogging != nil {
			var param LoggingMessage
			if err := json.Unmarshal(msg.Params, &param); err != nil {
				msg.SendError(ctx, fmt.Errorf("failed to unmarshal notifications/message: %w", err))
				return
			}
			if err := opts.OnLogging(ctx, param); err != nil {
				msg.SendError(ctx, fmt.Errorf("failed to handle notifications/message: %w", err))
				return
			}
		} else if strings.HasPrefix(msg.Method, "notifications/") && opts.OnNotify != nil {
			if err := opts.OnNotify(ctx, msg); err != nil {
				log.Errorf(ctx, "failed to handle notification: %v", err)
			}
		} else if opts.OnMessage != nil {
			if err := opts.OnMessage(ctx, msg); err != nil {
				log.Errorf(ctx, "failed to handle message: %v", err)
			}
		}
	})
}

func waitForURL(ctx context.Context, serverName, baseURL string) error {
	if baseURL == "" {
		return fmt.Errorf("base URL is empty for server %s", serverName)
	}

	for i := 0; i < 120; i++ {
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
		wire wire
		err  error
		opt  = complete.Complete(opts...)
	)

	if config.Command == "" && config.BaseURL == "" {
		return nil, fmt.Errorf("no command or base URL provided")
	} else if config.BaseURL != "" {
		if config.Command != "" {
			var err error
			config, err = r.Run(ctx, opt.Roots, opt.Env, serverName, config)
			if err != nil {
				return nil, err
			}
			if err := waitForURL(ctx, serverName, config.BaseURL); err != nil {
				return nil, err
			}
		}
		wire = NewHTTPClient(serverName, config.BaseURL, envvar.ReplaceMap(opt.Env, config.Headers))
	} else {
		wire, err = newStdioClient(ctx, opt.Roots, opt.Env, serverName, config)
		if err != nil {
			return nil, err
		}
	}

	session, err := newSession(ctx, wire, toHandler(opt), opt.SessionID, nil)
	if err != nil {
		return nil, err
	}
	session.Parent = opt.ParentSession
	return session, nil
}

func NewClient(ctx context.Context, serverName string, config Server, opts ...ClientOption) (*Client, error) {
	var (
		opt = complete.Complete(opts...)
	)

	session, err := NewSession(context.Background(), serverName, config, opts...)
	if err != nil {
		return nil, err
	}

	c := &Client{
		Session: session,
	}

	var (
		sampling *struct{}
		roots    *RootsCapability
	)
	if opt.OnSampling != nil {
		sampling = &struct{}{}
	}
	if opt.OnRoots != nil {
		roots = &RootsCapability{}
	}
	_, err = c.Initialize(ctx, InitializeRequest{
		ProtocolVersion: "2025-03-26",
		Capabilities: ClientCapabilities{
			Sampling: sampling,
			Roots:    roots,
		},
		ClientInfo: ClientInfo{
			Name:    "nanobot",
			Version: version.Get().String(),
		},
	})
	return c, err
}

func (c *Client) Initialize(ctx context.Context, param InitializeRequest) (result InitializeResult, err error) {
	err = c.Session.Exchange(ctx, "initialize", param, &result)
	if err == nil {
		err = c.Session.Send(ctx, Message{
			Method: "notifications/initialized",
		})
	}
	c.Session.InitializeResult = &result
	return
}

func (c *Client) ReadResource(ctx context.Context, uri string) (*ReadResourceResult, error) {
	var result ReadResourceResult
	err := c.Session.Exchange(ctx, "resources/read", ReadResourceRequest{
		URI: uri,
	}, &result)
	return &result, err
}

func (c *Client) ListResourceTemplates(ctx context.Context) (*ListResourceTemplatesResult, error) {
	var result ListResourceTemplatesResult
	if c.Session.InitializeResult == nil || c.Session.InitializeResult.Capabilities.Resources == nil {
		return &result, nil
	}
	err := c.Session.Exchange(ctx, "resources/templates/list", struct{}{}, &result)
	return &result, err
}

func (c *Client) ListResources(ctx context.Context) (*ListResourcesResult, error) {
	var result ListResourcesResult
	if c.Session.InitializeResult == nil || c.Session.InitializeResult.Capabilities.Resources == nil {
		return &result, nil
	}
	err := c.Session.Exchange(ctx, "resources/list", struct{}{}, &result)
	return &result, err
}

func (c *Client) ListPrompts(ctx context.Context) (*ListPromptsResult, error) {
	var prompts ListPromptsResult
	if c.Session.InitializeResult == nil || c.Session.InitializeResult.Capabilities.Prompts == nil {
		return &prompts, nil
	}
	err := c.Session.Exchange(ctx, "prompts/list", struct{}{}, &prompts)
	return &prompts, err
}

func (c *Client) GetPrompt(ctx context.Context, name string, args map[string]string) (*GetPromptResult, error) {
	var result GetPromptResult
	err := c.Session.Exchange(ctx, "prompts/get", GetPromptRequest{
		Name:      name,
		Arguments: args,
	}, &result)
	return &result, err
}

func (c *Client) ListTools(ctx context.Context) (*ListToolsResult, error) {
	var tools ListToolsResult
	err := c.Session.Exchange(ctx, "tools/list", struct{}{}, &tools)
	return &tools, err
}

type CallOption struct {
	ProgressToken any
}

func (c CallOption) Merge(other CallOption) (result CallOption) {
	result.ProgressToken = complete.Last(c.ProgressToken, other.ProgressToken)
	return
}

func (c *Client) Call(ctx context.Context, tool string, args any, opts ...CallOption) (result *CallToolResult, err error) {
	opt := complete.Complete(opts...)
	result = new(CallToolResult)

	err = c.Session.Exchange(ctx, "tools/call", struct {
		Name      string `json:"name"`
		Arguments any    `json:"arguments,omitempty"`
	}{
		Name:      tool,
		Arguments: args,
	}, result, ExchangeOption{
		ProgressToken: opt.ProgressToken,
	})

	return
}
