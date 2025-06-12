package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/chat/prompter"
	"github.com/nanobot-ai/nanobot/pkg/confirm"
	"github.com/nanobot-ai/nanobot/pkg/llm"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/printer"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
)

func Chat(ctx context.Context, listenAddress string, confirmations *confirm.Service,
	autoConfirm bool, prompt, output string, reload func(*mcp.Client) error) error {
	progressToken := uuid.String()

	promptDone, promptDoneCancel := context.WithCancel(ctx)
	defer promptDoneCancel()

	c, err := mcp.NewClient(ctx, "nanobot", mcp.Server{
		BaseURL: "http://" + listenAddress,
		Headers: nil,
	}, mcp.ClientOption{
		OnLogging: func(ctx context.Context, logMsg mcp.LoggingMessage) error {
			return handleLog(logMsg, confirmations, autoConfirm)
		},
		OnNotify: func(ctx context.Context, msg mcp.Message) error {
			if llm.PrintProgress(msg.Params) {
				return nil
			}
			printToolCall(msg.Params, promptDoneCancel)
			return nil
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create chat client: %w", err)
	}

	if prompt != "" {
		_, _ = fmt.Fprintf(os.Stderr, "> %s\n", prompt)
		resp, err := c.Call(ctx, types.AgentTool, map[string]any{
			"prompt": prompt,
		}, mcp.CallOption{
			ProgressToken: progressToken,
		})
		if err != nil {
			return fmt.Errorf("failed to call agent tool: %w", err)
		}
		if output != "" {
			var out io.Writer
			if output == "-" {
				out = os.Stdout
			} else {
				f, err := os.Create(output)
				if err != nil {
					return fmt.Errorf("failed to open output file: %w", err)
				}
				defer f.Close()
				out = f
			}
			if err := PrintResult(out, resp); err != nil {
				log.Errorf(ctx, "error printing: %v", err)
			}
		}
		<-promptDone.Done()
		return nil
	}

	_, _ = fmt.Fprintln(os.Stderr)
	intro, _ := c.Session.InitializeResult.Capabilities.Experimental["nanobot/intro"].(string)
	if intro != "" {
		printer.Prefix("INTRO", intro+"\n")
	}

	context.AfterFunc(ctx, func() {
		os.Exit(0)
	})

	for {
		line, err := prompter.ReadInput()
		if err != nil {
			return err
		}

		if strings.Fields(line)[0] == "/reload" {
			if err := reload(c); err != nil {
				log.Errorf(ctx, "Error reloading: %v", err)
			} else {
				log.Infof(ctx, "Reloaded successfully.")
			}
			continue
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		_, err = c.Call(ctx, types.AgentTool, map[string]any{
			"prompt": line,
		}, mcp.CallOption{
			ProgressToken: progressToken,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

var always = map[string]struct{}{}

func handleConfirm(data map[string]any, confirmations *confirm.Service, autoConfirm bool) error {
	request, _ := data["request"].(map[string]any)
	id, _ := request["id"].(string)

	if id == "" {
		return nil
	}

	if autoConfirm {
		confirmations.Reply(id, true)
		return nil
	}

	mcpServer, _ := request["mcpServer"].(string)
	toolName, _ := request["toolName"].(string)
	invocation, _ := request["invocation"].(map[string]any)
	args, _ := invocation["arguments"].(string)

	if _, ok := always[mcpServer+"/"+toolName]; ok {
		confirmations.Reply(id, true)
		return nil
	}

	_, _ = fmt.Fprintf(os.Stderr, "! Allow call to tool (%s) on MCP Server (%s)\n", toolName, mcpServer)
	if args != "" && args != "{}" {
		argsData := make(map[string]any)
		if err := json.Unmarshal([]byte(args), &argsData); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "!  args: %s\n", args)
		} else if len(argsData) > 0 {
			_, _ = fmt.Fprintf(os.Stderr, "!  args:\n")
			for k, v := range argsData {
				if _, ok := v.(string); !ok {
					vData, err := json.Marshal(v)
					if err == nil {
						v = string(vData)
					}
				}
				_, _ = fmt.Fprintf(os.Stderr, "!    %s: %s\n", k, v)
			}
		}
	}

	for {
		_, _ = fmt.Fprintf(os.Stderr, "!  (y/n/a) ? ")
		line, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
		if err != nil {
			return err
		}
		switch strings.TrimSpace(strings.ToLower(string(line))) {
		case "y", "yes":
			confirmations.Reply(id, true)
			return nil
		case "n", "no":
			confirmations.Reply(id, false)
			return nil
		case "a", "always":
			always[mcpServer+"/"+toolName] = struct{}{}
			confirmations.Reply(id, true)
			return nil
		}
	}
}

func printToolCall(params json.RawMessage, seenAgentOut func()) {
	var toolCall struct {
		Data struct {
			Type   string              `json:"type"`
			Input  any                 `json:"input,omitempty"`
			Error  string              `json:"error,omitempty"`
			Target string              `json:"target"`
			Output *mcp.CallToolResult `json:"output,omitempty"`
			Data   struct {
				MCPToolName string `json:"mcpToolName"`
			}
		} `json:"data"`
	}
	if err := json.Unmarshal(params, &toolCall); err != nil || !strings.HasPrefix(toolCall.Data.Type, "nanobot/call") {
		return
	}
	server, tool, _ := strings.Cut(toolCall.Data.Target, "/")
	if server == tool {
		toolCall.Data.Target = server
	}
	if toolCall.Data.Input != nil {
		var text string
		_ = types.Marshal(toolCall.Data.Input, &text)
		printer.Prefix(fmt.Sprintf("->(%s)", toolCall.Data.Target), text)
	}
	if toolCall.Data.Output != nil && toolCall.Data.Data.MCPToolName != types.AgentTool {
		for _, content := range toolCall.Data.Output.Content {
			printer.Prefix(fmt.Sprintf("<-(%s)", toolCall.Data.Target), content.Text)
		}
	}
	if toolCall.Data.Output != nil && toolCall.Data.Data.MCPToolName == types.AgentTool {
		seenAgentOut()
	}
}

func handleLog(msg mcp.LoggingMessage, confirmations *confirm.Service, autoConfirm bool) error {
	printed := false
	dataMap, ok := msg.Data.(map[string]any)
	msgType, _ := dataMap["type"].(string)

	if msgType == "nanobot/confirm" {
		return handleConfirm(dataMap, confirmations, autoConfirm)
	}

	if ok {
		server, serverOK := dataMap["server"].(string)
		data, dataOK := dataMap["data"].(map[string]any)
		dataString, dataStringOK := dataMap["data"].(string)
		if serverOK && dataOK {
			dataBytes, _ := json.Marshal(data)
			_, _ = fmt.Fprintf(os.Stderr, "%s(%s): %s\n", msg.Level, server, string(dataBytes))
			printed = true
		} else if serverOK && dataStringOK {
			_, _ = fmt.Fprintf(os.Stderr, "%s(%s): %s\n", msg.Level, server, dataString)
			printed = true
		}
	}

	if !printed {
		dataString, dataStringOK := msg.Data.(string)
		if dataStringOK {
			_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n", msg.Level, dataString)
		} else {
			dataBytes, _ := json.Marshal(msg.Data)
			_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n", msg.Level, string(dataBytes))
		}
	}

	return nil
}
