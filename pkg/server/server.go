package server

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/runtime"
	"github.com/obot-platform/nanobot/pkg/tools"
	"github.com/obot-platform/nanobot/pkg/types"
)

type Server struct {
	handlers []handler
	runtime  *runtime.Runtime
}

const (
	toolMappingKey = "toolMapping"
)

func NewServer(r *runtime.Runtime) *Server {
	s := &Server{
		runtime: r,
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
		if err := json.Unmarshal(msg.Params, &payload); err != nil {
			return false, err
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
	}
}

func (s *Server) handleCallTool(ctx context.Context, msg mcp.Message, payload mcp.CallToolRequest) error {
	toolMappings, _ := msg.Session.Get(toolMappingKey).(types.ToolMappings)
	toolMapping, ok := toolMappings[payload.Name]
	if !ok {
		return fmt.Errorf("tool %s not found", payload.Name)
	}

	result, err := s.runtime.Call(ctx, toolMapping.Server, toolMapping.ToolName, payload.Arguments, mcp.CallOption{
		ProgressToken: msg.ProgressToken(),
	})
	if err != nil {
		return err
	}

	return msg.Reply(ctx, result)
}

func (s *Server) handleListTools(ctx context.Context, msg mcp.Message, _ mcp.ListToolsRequest) error {
	result := mcp.ListToolsResult{
		Tools: []mcp.Tool{},
	}

	toolMappings, _ := msg.Session.Get(toolMappingKey).(types.ToolMappings)
	for _, k := range slices.Sorted(maps.Keys(toolMappings)) {
		result.Tools = append(result.Tools, mcp.Tool(toolMappings[k].Tool))
	}

	return msg.Reply(ctx, result)
}

func (s *Server) buildToolMappings(ctx context.Context, msg mcp.Message) (types.ToolMappings, error) {
	c := s.runtime.GetConfig()

	var servers []string
	for _, ref := range c.Publish.Tools {
		toolRef := types.ParseToolRef(ref)
		if toolRef.Server != "" {
			servers = append(servers, toolRef.Server)
		}
	}

	tools, err := s.runtime.ListTools(ctx, tools.ListToolsOptions{
		Servers: servers,
	})
	if err != nil {
		return nil, err
	}

	result := types.ToolMappings{}
	for _, ref := range c.Publish.Tools {
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

func (s *Server) handlePing(ctx context.Context, msg mcp.Message, _ mcp.PingRequest) error {
	return msg.Reply(ctx, mcp.PingResult{})
}

func (s *Server) handleInitialize(ctx context.Context, msg mcp.Message, payload mcp.InitializeRequest) error {
	c := s.runtime.GetConfig()

	toolMappings, err := s.runtime.BuildToolMappings(ctx, c.Publish.Tools)
	if err != nil {
		return err
	}
	msg.Session.Set(toolMappingKey, toolMappings)

	return msg.Reply(ctx, mcp.InitializeResult{
		ProtocolVersion: payload.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Logging:   nil,
			Prompts:   nil,
			Resources: nil,
			Tools:     &mcp.ToolsServerCapability{},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    c.Publish.Name,
			Version: c.Publish.Version,
		},
		Instructions: s.runtime.GetConfig().Publish.Instructions,
	})
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) error {
	for _, h := range s.handlers {
		if ok, err := h(ctx, msg); err != nil || ok {
			return err
		}
	}
	return nil
}
