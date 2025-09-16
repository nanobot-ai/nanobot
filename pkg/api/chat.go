package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

func writeEvent(rw http.ResponseWriter, id any, name string, textOrData any) error {
	asMap := make(map[string]any)
	if textOrData != nil {
		if err := mcp.JSONCoerce(textOrData, &asMap); err != nil {
			return fmt.Errorf("failed to coerce data: %w", err)
		}
	}

	// we want to make sure it's all one line
	data, err := json.Marshal(asMap)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if id != nil {
		_, err = rw.Write([]byte(fmt.Sprintf("id: %s\n", id)))
		if err != nil {
			return fmt.Errorf("failed to write id line: %w", err)
		}
	}

	if name != "message" {
		_, err = rw.Write([]byte(fmt.Sprintf("event: %s\n", name)))
		if err != nil {
			return fmt.Errorf("failed to write event line: %w", err)
		}
	}

	_, err = rw.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

func printHistory(rw http.ResponseWriter, req *http.Request, client *mcp.Client, printedIDs map[string]struct{}) error {
	resources, err := client.ListResources(req.Context())
	if err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	var progressURI string
	for _, resource := range resources.Resources {
		if resource.MimeType == types.HistoryMimeType {
			if err := writeEvent(rw, nil, "history-start", nil); err != nil {
				return fmt.Errorf("failed to write history-start: %w", err)
			}

			messages, err := client.ReadResource(req.Context(), resource.URI)
			if err != nil {
				return fmt.Errorf("failed to read history: %w", err)
			}
			for _, message := range messages.Contents {
				if message.MIMEType != types.MessageMimeType {
					continue
				}
				if err := writeEvent(rw, nil, "message", message.Text); err != nil {
					return err
				}
				var id string
				if err := json.Unmarshal([]byte(message.Text), &struct {
					ID *string `json:"id"`
				}{
					ID: &id,
				}); err != nil {
					return fmt.Errorf("failed to unmarshal message: %w", err)
				}
				if id != "" {
					printedIDs[id] = struct{}{}
				}
			}
			if err := writeEvent(rw, nil, "history-end", nil); err != nil {
				return fmt.Errorf("failed to write history-start: %w", err)
			}
		} else if resource.MimeType == types.ToolResultMimeType {
			progressURI = resource.URI
		}
	}

	if progressURI != "" {
		if err := printProgressURI(rw, req, client, progressURI, printedIDs); err != nil {
			return err
		}
	}

	return nil
}

func printProgressURI(rw http.ResponseWriter, req *http.Request, client *mcp.Client, progressURI string,
	printedIDs map[string]struct{}) error {
	messages, err := client.ReadResource(req.Context(), progressURI)
	if err != nil {
		return fmt.Errorf("failed to read history: %w", err)
	}
	for _, message := range messages.Contents {
		if message.MIMEType != types.ToolResultMimeType {
			continue
		}

		var callResult types.AsyncCallResult
		if err := json.Unmarshal([]byte(message.Text), &callResult); err != nil {
			return fmt.Errorf("failed to unmarshal tool result: %w", err)
		}

		if callResult.ToolName != types.AgentTool {
			continue
		}

		if callResult.InProgress {
			if err := writeEvent(rw, nil, "chat-in-progress", nil); err != nil {
				return err
			}
		}

		for _, progressMessage := range callResult.Content {
			if progressMessage.Resource != nil && progressMessage.Resource.MIMEType == types.MessageMimeType {
				var id string
				_ = json.Unmarshal([]byte(progressMessage.Resource.Text), &struct {
					ID *string `json:"id"`
				}{
					ID: &id,
				})
				if _, ok := printedIDs[id]; ok {
					continue
				}
				if err := writeEvent(rw, nil, "message", progressMessage.Resource.Text); err != nil {
					return err
				}
			}
		}

		if !callResult.InProgress {
			if err := writeEvent(rw, nil, "chat-done", nil); err != nil {
				return err
			}
		}
	}

	return nil
}

func Events(rw http.ResponseWriter, req *http.Request) error {
	apiContext := getContext(req.Context())

	state, err := apiContext.ChatClient.Session.State()
	if err != nil {
		return err
	}

	events := make(chan mcp.Message)
	subClient, err := mcp.NewClient(req.Context(), "nanobot.ui", apiContext.MCPServer, mcp.ClientOption{
		OnElicit: func(ctx context.Context, msg mcp.Message, _ mcp.ElicitRequest) (mcp.ElicitResult, error) {
			select {
			case events <- msg:
				return mcp.ElicitResult{
					Action: "handled",
				}, nil
			case <-req.Context().Done():
				return mcp.ElicitResult{}, req.Context().Err()
			case <-ctx.Done():
				// I'm pretty sure ctx and req.Context() are the same, but just in case...
				return mcp.ElicitResult{}, ctx.Err()
			}
		},
		OnNotify: func(ctx context.Context, msg mcp.Message) error {
			select {
			case events <- msg:
			case <-req.Context().Done():
				return req.Context().Err()
			case <-ctx.Done():
				// I'm pretty sure ctx and req.Context() are the same, but just in case...
				return ctx.Err()
			}
			return nil
		},
		SessionState: state,
	})
	if err != nil {
		return err
	}
	defer subClient.Close(false)

	_, _ = subClient.SubscribeResource(req.Context(), types.ProgressURI)

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.WriteHeader(200)
	if _, f := rw.(http.Flusher); f {
		rw.(http.Flusher).Flush()
	}

	ids := map[string]struct{}{}

	// Transform chat messages into SSE events
	if err := printHistory(rw, req, subClient, ids); err != nil {
		return err
	}

	for msg := range events {
		err := printProgressMessage(rw, req, msg, subClient, ids)
		if err != nil {
			return err
		}
	}

	return nil
}

func printProgressMessage(rw http.ResponseWriter, req *http.Request, msg mcp.Message, client *mcp.Client, printedIDs map[string]struct{}) error {
	defer func() {
		if f, ok := rw.(http.Flusher); ok {
			f.Flush()
		}
	}()

	if msg.Method == "notifications/resources/updated" {
		var data struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(msg.Params, &data); err != nil {
			return fmt.Errorf("failed to unmarshal params: %w", err)
		}
		if data.URI != "" {
			return printProgressURI(rw, req, client, data.URI, printedIDs)
		}
	}

	if msg.Error != nil {
		return writeEvent(rw, nil, "error", msg.Error)
	} else if msg.Method != "" && len(msg.Params) > 0 {
		return writeEvent(rw, msg.ID, msg.Method, string(msg.Params))
	}

	return nil
}
