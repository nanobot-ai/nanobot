package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/fswatch"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"github.com/nanobot-ai/nanobot/pkg/version"
	"gorm.io/gorm"
)

// taskResult is the agent-facing JSON shape for a scheduled task.
// Keeps gorm.Model internals out of the API.
type taskResult struct {
	URI        string     `json:"uri"`
	Name       string     `json:"name"`
	Prompt     string     `json:"prompt"`
	Schedule   string     `json:"schedule"`
	Timezone   string     `json:"timezone"`
	Expiration string     `json:"expiration,omitempty"`
	Enabled    bool       `json:"enabled"`
	LastRunAt  *time.Time `json:"lastRunAt,omitempty"`
	NextRunAt  *time.Time `json:"nextRunAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func toResult(task session.ScheduledTask) taskResult {
	var expiration string
	if task.ExpiresAt != nil {
		if loc, err := time.LoadLocation(task.Timezone); err == nil {
			expiration = task.ExpiresAt.In(loc).Format(time.DateOnly)
		}
	}
	return taskResult{
		URI:        task.TaskURI,
		Name:       task.Name,
		Prompt:     task.Prompt,
		Schedule:   task.Schedule,
		Timezone:   task.Timezone,
		Expiration: expiration,
		Enabled:    task.Enabled,
		LastRunAt:  task.LastRunAt,
		NextRunAt:  task.NextRunAt,
		CreatedAt:  task.CreatedAt,
		UpdatedAt:  task.UpdatedAt,
	}
}

// Server is a singleton that manages scheduled tasks. It handles MCP tools and
// resources, CRUD persistence, scheduling, and task execution via loopback HTTP.
type Server struct {
	*fswatch.SubscriptionManager
	tools       mcp.ServerTools
	db          *session.Store
	loopbackURL string
	ctx         context.Context
	cancel      context.CancelFunc
	startOnce   sync.Once
	wg          sync.WaitGroup
	mu          sync.Mutex
	jobs        map[string]*job
}

type job struct {
	reschedule chan struct{}
	cancel     context.CancelFunc
}

// NewServer creates the task server. The DB and scheduler are initialized later
// via Start.
func NewServer(loopbackURL string) *Server {
	s := &Server{
		SubscriptionManager: fswatch.NewSubscriptionManager(context.Background()),
		loopbackURL:         loopbackURL,
		jobs:                make(map[string]*job),
	}
	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("listScheduledTasks", "List scheduled tasks", s.listTasks),
		mcp.NewServerTool("createScheduledTask", "Create a scheduled task", s.createTask),
		mcp.NewServerTool("updateScheduledTask", "Update a scheduled task", s.updateTask),
		mcp.NewServerTool("deleteScheduledTask", "Delete a scheduled task", s.deleteTask),
		mcp.NewServerTool("startScheduledTask", "Start a scheduled task now", s.startTask),
	)
	return s
}

// Start sets the DB and loads persisted tasks.
func (s *Server) Start(ctx context.Context, db *session.Store) error {
	var err error
	s.startOnce.Do(func() {
		s.ctx, s.cancel = context.WithCancel(ctx)
		s.db = db

		var tasks []session.ScheduledTask
		tasks, err = db.ListScheduledTasks(ctx)
		if err != nil {
			return
		}

		for _, task := range tasks {
			if task.Enabled {
				s.scheduleTask(task.TaskURI)
			}
		}
	})

	return err
}

// Stop shuts down all scheduled goroutines, waiting up to the ctx deadline.
func (s *Server) Stop(ctx context.Context) error {
	s.cancel()
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// OnMessage dispatches MCP messages.
func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	switch msg.Method {
	case "initialize":
		mcp.Invoke(ctx, msg, s.initialize)
	case "notifications/initialized":
	case "notifications/cancelled":
		mcp.HandleCancelled(ctx, msg)
	case "resources/list":
		mcp.Invoke(ctx, msg, s.resourcesList)
	case "resources/read":
		mcp.Invoke(ctx, msg, s.resourcesRead)
	case "resources/subscribe":
		mcp.Invoke(ctx, msg, s.resourcesSubscribe)
	case "resources/unsubscribe":
		mcp.Invoke(ctx, msg, s.resourcesUnsubscribe)
	case "tools/list":
		mcp.Invoke(ctx, msg, s.tools.List)
	case "tools/call":
		mcp.Invoke(ctx, msg, s.tools.Call)
	default:
		msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage("%v", msg.Method))
	}
}

func (s *Server) listTasks(ctx context.Context, _ struct{}) (struct {
	Tasks []taskResult `json:"tasks"`
}, error) {
	var zero struct {
		Tasks []taskResult `json:"tasks"`
	}
	tasks, err := s.db.ListScheduledTasks(ctx)
	if err != nil {
		return zero, err
	}
	results := make([]taskResult, 0, len(tasks))
	for _, t := range tasks {
		results = append(results, toResult(t))
	}
	return struct {
		Tasks []taskResult `json:"tasks"`
	}{Tasks: results}, nil
}

func (s *Server) createTask(ctx context.Context, params struct {
	Name       string `json:"name"`
	Prompt     string `json:"prompt"`
	Schedule   string `json:"schedule"`
	Timezone   string `json:"timezone"`
	Expiration string `json:"expiration,omitempty"`
	Enabled    bool   `json:"enabled,omitempty"`
}) (*taskResult, error) {
	if params.Name == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("name is required")
	}
	if params.Prompt == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("prompt is required")
	}
	if params.Timezone == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("timezone is required")
	}

	spec, loc, err := parseSchedule(params.Schedule, params.Timezone)
	if err != nil {
		return nil, err
	}

	expiresAt, err := parseExpiration(params.Expiration, loc)
	if err != nil {
		return nil, err
	}
	if err := validateSchedule(params.Schedule, expiresAt != nil); err != nil {
		return nil, err
	}

	taskURI, err := s.db.NextScheduledTaskURI(ctx, params.Name)
	if err != nil {
		return nil, err
	}

	task := session.ScheduledTask{
		TaskURI:   taskURI,
		Name:      params.Name,
		Prompt:    params.Prompt,
		Schedule:  params.Schedule,
		Timezone:  params.Timezone,
		ExpiresAt: expiresAt,
		Enabled:   params.Enabled,
		NextRunAt: nextRunAt(spec, loc, expiresAt, time.Now()),
	}

	if err := s.db.CreateScheduledTask(ctx, &task); err != nil {
		return nil, fmt.Errorf("failed to create: %w", err)
	}

	if task.Enabled {
		s.scheduleTask(taskURI)
	}
	s.SendListChangedNotification()

	result := toResult(task)
	return &result, nil
}

func (s *Server) updateTask(ctx context.Context, params struct {
	URI        string  `json:"uri"`
	Name       string  `json:"name,omitempty"`
	Prompt     string  `json:"prompt,omitempty"`
	Schedule   string  `json:"schedule,omitempty"`
	Timezone   string  `json:"timezone,omitempty"`
	Expiration *string `json:"expiration,omitempty"`
	Enabled    *bool   `json:"enabled,omitempty"`
}) (*taskResult, error) {
	if params.URI == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("uri is required")
	}

	task, err := s.db.GetScheduledTask(ctx, params.URI)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("task %q not found", params.URI)
	}
	if err != nil {
		return nil, err
	}

	if params.Name != "" {
		task.Name = params.Name
	}
	if params.Prompt != "" {
		task.Prompt = params.Prompt
	}
	if params.Schedule != "" {
		task.Schedule = params.Schedule
	}
	if params.Timezone != "" {
		task.Timezone = params.Timezone
	}
	if params.Enabled != nil {
		task.Enabled = *params.Enabled
	}

	// Validate final state and recompute NextRunAt.
	spec, loc, err := parseSchedule(task.Schedule, task.Timezone)
	if err != nil {
		return nil, err
	}
	if params.Expiration != nil {
		expiresAt, err := parseExpiration(*params.Expiration, loc)
		if err != nil {
			return nil, err
		}
		task.ExpiresAt = expiresAt
	}
	if err := validateSchedule(task.Schedule, task.ExpiresAt != nil); err != nil {
		return nil, err
	}
	if task.Enabled {
		task.NextRunAt = nextRunAt(spec, loc, task.ExpiresAt, time.Now())
	} else {
		task.NextRunAt = nil
	}

	if err := s.db.UpdateScheduledTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update: %w", err)
	}

	if task.Enabled {
		s.scheduleTask(task.TaskURI)
	} else {
		s.cancelTask(task.TaskURI)
	}
	s.SendListChangedNotification()
	s.SendResourceUpdatedNotification(task.TaskURI)

	result := toResult(*task)
	return &result, nil
}

func (s *Server) deleteTask(ctx context.Context, params struct {
	URI string `json:"uri"`
}) (string, error) {
	if params.URI == "" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("uri is required")
	}
	if err := s.db.DeleteScheduledTask(ctx, params.URI); err != nil {
		return "", fmt.Errorf("failed to delete: %w", err)
	}
	s.cancelTask(params.URI)
	s.SendListChangedNotification()
	s.AutoUnsubscribe(params.URI)
	return fmt.Sprintf("%s deleted", params.URI), nil
}

func (s *Server) startTask(ctx context.Context, params struct {
	URI string `json:"uri"`
}) (*struct {
	Message   string `json:"message"`
	SessionID string `json:"sessionId"`
	TaskURI   string `json:"taskURI"`
}, error) {
	if params.URI == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("uri is required")
	}
	task, err := s.db.GetScheduledTask(ctx, params.URI)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("task %q not found", params.URI)
	}
	if err != nil {
		return nil, err
	}
	sessionID, err := s.startChat(ctx, *task)
	if err != nil {
		return nil, err
	}
	return &struct {
		Message   string `json:"message"`
		SessionID string `json:"sessionId"`
		TaskURI   string `json:"taskURI"`
	}{
		Message:   "Task successfully started",
		SessionID: sessionID,
		TaskURI:   params.URI,
	}, nil
}

func (s *Server) initialize(ctx context.Context, msg mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	result := &mcp.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities:    mcp.ServerCapabilities{Tools: &mcp.ToolsServerCapability{}},
		ServerInfo:      mcp.ServerInfo{Name: version.Name, Version: version.Get().String()},
	}
	if types.IsUISession(ctx) || types.IsChatSession(ctx) {
		s.AddSession(msg.Session.Root().ID(), msg.Session.Root())
		result.Capabilities.Resources = &mcp.ResourcesServerCapability{Subscribe: true, ListChanged: true}
	}
	return result, nil
}

func (s *Server) resourcesList(ctx context.Context, _ mcp.Message, _ mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	tasks, err := s.db.ListScheduledTasks(ctx)
	if err != nil {
		return nil, err
	}
	resources := make([]mcp.Resource, 0, len(tasks))
	for _, t := range tasks {
		taskMeta := map[string]any{
			"createdAt": t.CreatedAt.Format(time.RFC3339),
			"enabled":   t.Enabled,
			"schedule":  t.Schedule,
			"timezone":  t.Timezone,
		}
		if t.ExpiresAt != nil {
			if loc, err := time.LoadLocation(t.Timezone); err == nil {
				taskMeta["expiration"] = t.ExpiresAt.In(loc).Format(time.DateOnly)
			}
		}
		resources = append(resources, mcp.Resource{
			URI:         t.TaskURI,
			Name:        t.Name,
			MimeType:    "application/json",
			Annotations: &mcp.Annotations{LastModified: t.UpdatedAt},
			Meta: map[string]any{
				types.MetaPrefix + "task": taskMeta,
			},
		})
	}
	return &mcp.ListResourcesResult{Resources: resources}, nil
}

func (s *Server) resourcesRead(ctx context.Context, _ mcp.Message, req mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	task, err := s.db.GetScheduledTask(ctx, req.URI)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("task %q not found", req.URI)
	}
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(toResult(*task))
	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{{
			URI:      req.URI,
			Name:     task.Name,
			MIMEType: "application/json",
			Text:     new(string(data)),
		}},
	}, nil
}

func (s *Server) resourcesSubscribe(_ context.Context, msg mcp.Message, req mcp.SubscribeRequest) (*mcp.SubscribeResult, error) {
	s.Subscribe(msg.Session.Root().ID(), msg.Session.Root(), req.URI)
	return &mcp.SubscribeResult{}, nil
}

func (s *Server) resourcesUnsubscribe(ctx context.Context, _ mcp.Message, req mcp.UnsubscribeRequest) (*mcp.UnsubscribeResult, error) {
	s.Unsubscribe(mcp.SessionFromContext(ctx).Root().ID(), req.URI)
	return &mcp.UnsubscribeResult{}, nil
}

// scheduleTask reschedules an existing goroutine to re-read from DB, or spawns a new one.
func (s *Server) scheduleTask(taskURI string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if j, ok := s.jobs[taskURI]; ok {
		select {
		case j.reschedule <- struct{}{}:
		default:
		}
		return
	}

	ctx, cancel := context.WithCancel(s.ctx)
	reschedule := make(chan struct{}, 1)
	s.jobs[taskURI] = &job{reschedule: reschedule, cancel: cancel}
	s.wg.Go(func() {
		defer func() {
			s.mu.Lock()
			if j, ok := s.jobs[taskURI]; ok && j.reschedule == reschedule {
				delete(s.jobs, taskURI)
			}
			s.mu.Unlock()
		}()

		for {
			task, err := s.db.GetScheduledTask(ctx, taskURI)
			if err != nil || !task.Enabled {
				return
			}

			spec, loc, err := parseSchedule(task.Schedule, task.Timezone)
			if err != nil {
				return
			}

			next := nextRunAt(spec, loc, task.ExpiresAt, time.Now())
			if next == nil {
				return
			}

			timer := time.NewTimer(time.Until(*next))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-reschedule:
				timer.Stop()
				continue
			case <-timer.C:
			}

			// Re-read for freshest prompt/name before executing.
			task, err = s.db.GetScheduledTask(ctx, taskURI)
			if err != nil {
				return
			}

			now := time.Now().UTC()
			next = nextRunAt(spec, loc, task.ExpiresAt, now)
			if err := s.db.RecordScheduledTaskRun(ctx, taskURI, now, next); err != nil {
				slog.Error("scheduled task: failed to record run", "task_uri", taskURI, "error", err)
			}

			s.SendResourceUpdatedNotification(taskURI)
			s.SendListChangedNotification()

			if _, err := s.startChat(ctx, *task); err != nil {
				slog.Error("scheduled task: failed to run", "task_uri", taskURI, "error", err)
			}
		}
	})
}

// cancelTask stops a scheduled goroutine and removes it from the jobs map.
func (s *Server) cancelTask(taskURI string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if j, ok := s.jobs[taskURI]; ok {
		j.cancel()
		delete(s.jobs, taskURI)
	}
}

func (s *Server) startChat(ctx context.Context, task session.ScheduledTask) (string, error) {
	client, err := mcp.NewClient(ctx, "nanobot-scheduler", mcp.Server{
		BaseURL: s.loopbackURL,
		Headers: map[string]string{
			"X-Nanobot-Description": task.Name,
			"X-Nanobot-Task-URI":    task.TaskURI,
		},
	}, mcp.ClientOption{
		ClientName: "nanobot-scheduler",
		OnMessage: func(_ context.Context, msg mcp.Message) error {
			return nil
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer client.Close(false)
	_, err = client.Call(ctx, types.AgentTool+"nanobot", map[string]any{
		"prompt": task.Prompt,
	}, mcp.CallOption{
		ProgressToken: uuid.String(),
		Meta:          map[string]any{types.AsyncMetaKey: true},
	})

	return client.Session.ID(), err
}
