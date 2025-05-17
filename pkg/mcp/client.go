package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"slices"

	"github.com/obot-platform/nanobot/pkg/types"
)

type ClientCapabilities struct {
	Roots    RootsCapability `json:"roots,omitzero"`
	Sampling *struct{}       `json:"sampling,omitzero"`
}

type RootsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerCapabilities struct {
	Logging   *struct{}                  `json:"logging,omitempty"`
	Prompts   *PromptsServerCapability   `json:"prompts,omitempty"`
	Resources *ResourcesServerCapability `json:"resources,omitempty"`
	Tools     *ToolsServerCapability     `json:"tools,omitempty"`
}

type ToolsServerCapability struct {
	ListChanged bool `json:"listChanged"`
}

type PromptsServerCapability struct {
	ListChanged bool `json:"listChanged"`
}

type ResourcesServerCapability struct {
	Subscribe   bool `json:"subscribe"`
	ListChanged bool `json:"listChanged"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Instructions    string             `json:"instructions"`
}

type InitializeRequest struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

type PingRequest struct {
}

type PingResult struct {
}

type ModelPreferences struct {
	Hints                []ModelHint `json:"hints,omitzero"`
	CostPriority         *float64    `json:"costPriority"`
	SpeedPriority        *float64    `json:"speedPriority"`
	IntelligencePriority *float64    `json:"intelligencePriority"`
}

type ModelHint struct {
	Name string `json:"name"`
}
type CreateMessageRequest struct {
	Messages         []SamplingMessage `json:"messages,omitzero"`
	ModelPreferences ModelPreferences  `json:"modelPreferences,omitzero"`
	SystemPrompt     string            `json:"systemPrompt,omitzero"`
	IncludeContext   string            `json:"includeContext,omitempty"`
	MaxTokens        int               `json:"maxTokens,omitempty"`
	Temperature      *json.Number      `json:"temperature,omitempty"`
	StopSequences    []string          `json:"stopSequences,omitzero"`
	Metadata         map[string]any    `json:"metadata,omitempty"`
}

type SamplingMessage struct {
	Role    string  `json:"role,omitempty"`
	Content Content `json:"content,omitempty"`
}

type CreateMessageResult struct {
	Content    Content `json:"content,omitempty"`
	Role       string  `json:"role,omitempty"`
	Model      string  `json:"model,omitempty"`
	StopReason string  `json:"stopReason,omitempty"`
}

type Content struct {
	Type string `json:"type,omitempty"`

	// Text is set when type is "text"
	Text string `json:"text,omitempty"`

	// Data is set when type is "image" or "audio"
	Data string `json:"data,omitempty"`
	// MIMEType is set when type is "image" or "audio"
	MIMEType string `json:"mimeType,omitempty"`

	// Resource is set when type is "resource"
	Resource *Resource `json:"resource,omitempty"`
}

func (c *Content) ToImageURL() string {
	return "data:" + c.MIMEType + ";base64," + c.Data
}

type Resource struct {
	URI      string `json:"uri,omitempty"`
	MIMEType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema,omitzero"`
}

type CallToolResult struct {
	IsError bool      `json:"isError"`
	Content []Content `json:"content,omitzero"`
}

type CallToolRequest struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type ListToolsRequest struct {
}

type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

type Notification struct {
}

type NotificationProgressRequest struct {
	ProgressToken any          `json:"progressToken"`
	Progress      json.Number  `json:"progress"`
	Total         *json.Number `json:"total,omitempty"`
	Message       string       `json:"message,omitempty"`
	Data          any          `json:"data,omitzero"`
}

type Client struct {
	Session *Session
}

type ClientOption struct {
	OnSampling func(ctx context.Context, sampling CreateMessageRequest) (CreateMessageResult, error)
	OnMessage  func(ctx context.Context, msg Message) error
	Env        map[string]string
	SessionID  string
}

func toHandler(opts ClientOption) MessageHandler {
	return MessageHandlerFunc(func(ctx context.Context, msg Message) error {
		if msg.Method == "sampling/createMessage" && opts.OnSampling != nil {
			var param CreateMessageRequest
			if err := json.Unmarshal(msg.Params, &param); err != nil {
				return fmt.Errorf("failed to unmarshal sampling/createMessage: %w", err)
			}
			resp, err := opts.OnSampling(ctx, param)
			if err != nil {
				return fmt.Errorf("failed to handle sampling/createMessage: %w", err)
			}
			return msg.Reply(ctx, resp)
		} else if opts.OnMessage != nil {
			return opts.OnMessage(ctx, msg)
		}
		return nil
	})
}

func complete(opts ...ClientOption) ClientOption {
	var all ClientOption
	for _, opt := range opts {
		if opt.OnSampling != nil {
			if all.OnSampling != nil {
				panic("multiple OnSampling handlers provided")
			}
			all.OnSampling = opt.OnSampling
		}
		if opt.OnMessage != nil {
			if all.OnMessage != nil {
				panic("multiple OnMessage handlers provided")
			}
			all.OnMessage = opt.OnMessage
		}
		if len(opt.Env) > 0 {
			if all.Env == nil {
				all.Env = make(map[string]string)
			}
			for k, v := range opt.Env {
				all.Env[k] = v
			}
		}
		if opt.SessionID != "" {
			if all.SessionID != "" {
				panic("multiple SessionID provided")
			}
			all.SessionID = opt.SessionID
		}
	}
	return all
}

func replaceString(envs map[string]string, str string) string {
	return os.Expand(str, func(key string) string {
		if val, ok := envs[key]; ok {
			return val
		}
		return "${" + key + "}"
	})
}

func replaceMap(envs map[string]string, m map[string]string) map[string]string {
	newMap := make(map[string]string, len(m))
	for k, v := range m {
		newMap[replaceString(envs, k)] = replaceString(envs, v)
	}
	return newMap
}

func replaceEnv(envs map[string]string, command string, args []string, env map[string]string) (string, []string, []string) {
	newEnvMap := maps.Clone(envs)
	if newEnvMap == nil {
		newEnvMap = make(map[string]string, len(env))
	}
	maps.Copy(newEnvMap, replaceMap(envs, env))

	newEnv := make([]string, 0, len(env))
	for _, k := range slices.Sorted(maps.Keys(newEnvMap)) {
		newEnv = append(newEnv, fmt.Sprintf("%s=%s", k, newEnvMap[k]))
	}

	newArgs := make([]string, len(args))
	for i, arg := range args {
		newArgs[i] = replaceString(envs, arg)
	}
	return replaceString(envs, command), newArgs, newEnv
}

func NewSession(ctx context.Context, serverName string, config types.MCPServer, opts ...ClientOption) (*Session, error) {
	var (
		wire wire
		err  error
		opt  = complete(opts...)
	)

	cmd, args, env := replaceEnv(opt.Env, config.Command, config.Args, config.Env)
	headers := replaceMap(opt.Env, config.Headers)

	if config.Command == "" && config.BaseURL == "" {
		return nil, fmt.Errorf("no command or base URL provided")
	} else if config.BaseURL != "" {
		if config.Command != "" {
			if err := runCommand(ctx, cmd, args, env); err != nil {
				return nil, err
			}
		}
		wire = NewHTTPClient(config.BaseURL, headers)
	} else {
		wire, err = newStdioClient(ctx, serverName, cmd, args, env)
		if err != nil {
			return nil, err
		}
	}

	return newSession(ctx, wire, toHandler(opt), opt.SessionID, nil)
}

func runCommand(ctx context.Context, cmd string, args []string, env []string) error {
	osCmd := exec.CommandContext(ctx, cmd, args...)
	osCmd.Env = env
	osCmd.Stdout = os.Stdout
	osCmd.Stderr = os.Stderr
	if err := osCmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}
	return nil
}

func NewClient(ctx context.Context, serverName string, config types.MCPServer, opts ...ClientOption) (*Client, error) {
	var (
		opt = complete(opts...)
	)

	session, err := NewSession(ctx, serverName, config, opts...)
	if err != nil {
		return nil, err
	}

	c := &Client{
		Session: session,
	}

	var sampling *struct{}
	if opt.OnSampling != nil {
		sampling = &struct{}{}
	}
	_, err = c.Initialize(ctx, InitializeRequest{
		ProtocolVersion: "2025-03-26",
		Capabilities: ClientCapabilities{
			Sampling: sampling,
		},
		ClientInfo: ClientInfo{},
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
	return
}

func (c *Client) ListTools(ctx context.Context) (*ListToolsResult, error) {
	var tools ListToolsResult
	err := c.Session.Exchange(ctx, "tools/list", struct{}{}, &tools)
	return &tools, err
}

type CallOption struct {
	ID            string
	ProgressToken any
}

func CompleteCallOptions(opts ...CallOption) CallOption {
	var all CallOption
	for _, opt := range opts {
		if opt.ID != "" {
			all.ID = opt.ID
		}
		if opt.ProgressToken != nil {
			all.ProgressToken = opt.ProgressToken
		}
	}
	return all
}

func (c *Client) Call(ctx context.Context, tool string, args any, opts ...CallOption) (result *CallToolResult, err error) {
	opt := CompleteCallOptions(opts...)
	result = new(CallToolResult)

	err = c.Session.Exchange(ctx, "tools/call", struct {
		Name      string `json:"name"`
		Arguments any    `json:"arguments,omitempty"`
	}{
		Name:      tool,
		Arguments: args,
	}, result, ExchangeOption{
		ID:            opt.ID,
		ProgressToken: opt.ProgressToken,
	})

	return
}
