package tasks

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/session"
)

func testServer(t *testing.T) *Server {
	t.Helper()

	manager, err := session.NewManager(fmt.Sprintf("sqlite:file:%s?mode=memory&cache=shared",
		strings.NewReplacer("/", "-", " ", "-").Replace(t.Name())))
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}

	srv := NewServer("")
	if err := srv.Start(t.Context(), manager.DB); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	return srv
}

func TestCreateAndListTasks(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	task, err := srv.createTask(ctx, struct {
		Name       string `json:"name"`
		Prompt     string `json:"prompt"`
		Schedule   string `json:"schedule"`
		Timezone   string `json:"timezone"`
		Expiration string `json:"expiration,omitempty"`
		Enabled    bool   `json:"enabled,omitempty"`
	}{
		Name: "Daily Summary", Prompt: "Summarize.", Schedule: "0 9 * * *", Timezone: "America/New_York",
	})
	if err != nil {
		t.Fatalf("createTask: %v", err)
	}
	if task.URI != "task:///daily-summary" {
		t.Fatalf("task.URI = %q, want %q", task.URI, "task:///daily-summary")
	}
	if task.Schedule != "0 9 * * *" {
		t.Fatalf("task.Schedule = %q", task.Schedule)
	}
	if task.Timezone != "America/New_York" {
		t.Fatalf("task.Timezone = %q", task.Timezone)
	}

	stored, err := srv.db.GetScheduledTask(ctx, "task:///daily-summary")
	if err != nil {
		t.Fatalf("GetScheduledTask: %v", err)
	}
	if stored.Prompt != "Summarize." {
		t.Fatalf("stored.Prompt = %q", stored.Prompt)
	}

	listed, err := srv.listTasks(ctx, struct{}{})
	if err != nil {
		t.Fatalf("listTasks: %v", err)
	}
	if len(listed.Tasks) != 1 {
		t.Fatalf("len(listed.Tasks) = %d, want 1", len(listed.Tasks))
	}
	if listed.Tasks[0].URI != "task:///daily-summary" {
		t.Fatalf("listed.Tasks[0].URI = %q", listed.Tasks[0].URI)
	}
}

func TestResourcesExposeTaskMetadata(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	if _, err := srv.createTask(ctx, struct {
		Name       string `json:"name"`
		Prompt     string `json:"prompt"`
		Schedule   string `json:"schedule"`
		Timezone   string `json:"timezone"`
		Expiration string `json:"expiration,omitempty"`
		Enabled    bool   `json:"enabled,omitempty"`
	}{
		Name: "Daily Summary", Prompt: "Summarize.", Schedule: "0 9 * * *", Timezone: "America/New_York",
	}); err != nil {
		t.Fatalf("createTask: %v", err)
	}

	list, err := srv.resourcesList(ctx, mcp.Message{}, mcp.ListResourcesRequest{})
	if err != nil {
		t.Fatalf("resourcesList: %v", err)
	}
	if len(list.Resources) != 1 {
		t.Fatalf("len = %d, want 1", len(list.Resources))
	}
	if list.Resources[0].URI != "task:///daily-summary" {
		t.Fatalf("URI = %q", list.Resources[0].URI)
	}

	read, err := srv.resourcesRead(ctx, mcp.Message{}, mcp.ReadResourceRequest{URI: "task:///daily-summary"})
	if err != nil {
		t.Fatalf("resourcesRead: %v", err)
	}
	if len(read.Contents) != 1 || read.Contents[0].Text == nil {
		t.Fatal("expected text content")
	}
	if !strings.Contains(*read.Contents[0].Text, `"schedule":"0 9 * * *"`) {
		t.Fatalf("body = %q", *read.Contents[0].Text)
	}
}

func TestUpdateTask(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	task, err := srv.createTask(ctx, struct {
		Name       string `json:"name"`
		Prompt     string `json:"prompt"`
		Schedule   string `json:"schedule"`
		Timezone   string `json:"timezone"`
		Expiration string `json:"expiration,omitempty"`
		Enabled    bool   `json:"enabled,omitempty"`
	}{
		Name: "Heartbeat", Prompt: "Ping", Schedule: "0 9 * * 1,3", Timezone: "America/New_York", Enabled: true,
	})
	if err != nil {
		t.Fatalf("createTask: %v", err)
	}

	updated, err := srv.updateTask(ctx, struct {
		URI        string  `json:"uri"`
		Name       string  `json:"name,omitempty"`
		Prompt     string  `json:"prompt,omitempty"`
		Schedule   string  `json:"schedule,omitempty"`
		Timezone   string  `json:"timezone,omitempty"`
		Expiration *string `json:"expiration,omitempty"`
		Enabled    *bool   `json:"enabled,omitempty"`
	}{
		URI:    task.URI,
		Prompt: "Pong",
	})
	if err != nil {
		t.Fatalf("updateTask: %v", err)
	}
	if updated.Prompt != "Pong" {
		t.Fatalf("updated.Prompt = %q, want %q", updated.Prompt, "Pong")
	}
}

func TestContiguousTaskURIsAfterDelete(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	create := func(prompt string) *taskResult {
		t.Helper()
		task, err := srv.createTask(ctx, struct {
			Name       string `json:"name"`
			Prompt     string `json:"prompt"`
			Schedule   string `json:"schedule"`
			Timezone   string `json:"timezone"`
			Expiration string `json:"expiration,omitempty"`
			Enabled    bool   `json:"enabled,omitempty"`
		}{
			Name: "Daily Summary", Prompt: prompt, Schedule: "0 9 * * *", Timezone: "America/New_York",
		})
		if err != nil {
			t.Fatalf("createTask(%q): %v", prompt, err)
		}
		return task
	}

	first := create("First")
	second := create("Second")
	if second.URI != "task:///daily-summary-2" {
		t.Fatalf("second.URI = %q, want %q", second.URI, "task:///daily-summary-2")
	}

	if _, err := srv.deleteTask(ctx, struct {
		URI string `json:"uri"`
	}{URI: first.URI}); err != nil {
		t.Fatalf("deleteTask: %v", err)
	}

	third := create("Third")
	if third.URI != "task:///daily-summary-3" {
		t.Fatalf("third.URI = %q, want %q", third.URI, "task:///daily-summary-3")
	}
}
