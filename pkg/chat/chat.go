package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/obot-platform/nanobot/pkg/confirm"
	"github.com/obot-platform/nanobot/pkg/llm"
	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/printer"
	"github.com/obot-platform/nanobot/pkg/types"
	"github.com/obot-platform/nanobot/pkg/uuid"
)

func Chat(ctx context.Context, listenAddress string, confirmations *confirm.Service, autoConfirm bool, prompt, output string) error {
	progressToken := uuid.String()

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
			printToolCall(msg.Params)
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
				fmt.Printf("Error: %v\n", err)
			}
		}
		return nil
	}

	_, _ = fmt.Fprintln(os.Stderr)
	intro, _ := c.Session.ServerCapabilities.Experimental["nanobot/intro"].(string)
	if intro != "" {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", intro)
	}

	scanner := bufio.NewScanner(os.Stdin)

	first := true
	next := func() bool {
		if !first {
			// Arbitrary delay to prevent the progress from override the prompt.
			// This is because the progress comes through the SSE channel where as the chat response
			// comes from the POST response. So it's possible to get the POST response before the SSE
			// responses are all done. (Yeah, annoying, I know. But switching to SSE and not HTTP streaming
			// will fix this once I have SSE working.)
			time.Sleep(300 * time.Millisecond)
		}
		first = false
		fmt.Println()
		fmt.Print("> ")
		return scanner.Scan()
	}

	context.AfterFunc(ctx, func() {
		os.Exit(0)
	})

	for next() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		_, err := c.Call(ctx, types.AgentTool, map[string]any{
			"prompt": line,
		}, mcp.CallOption{
			ProgressToken: progressToken,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return nil
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

func printToolCall(params json.RawMessage) {
	var toolCall struct {
		Data struct {
			Type   string              `json:"type"`
			Input  any                 `json:"input,omitempty"`
			Error  string              `json:"error,omitempty"`
			Target string              `json:"target"`
			Output *mcp.CallToolResult `json:"output,omitempty"`
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
	if toolCall.Data.Output != nil {
		for _, content := range toolCall.Data.Output.Content {
			printer.Prefix(fmt.Sprintf("<-(%s)", toolCall.Data.Target), content.Text)
		}
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
