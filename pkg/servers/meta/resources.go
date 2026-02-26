package meta

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/servers/workflows"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"gopkg.in/yaml.v3"
)

const (
	workflowsDir      = "workflows"
	agentURIPrefix    = "agent:///"
	chatThreadPrefix  = "chat:///threads/"
	workflowURIPrefix = "workflow:///"
	fileURIPrefix     = "file:///"
)

type workflowMeta struct {
	Name      string `yaml:"name"`
	CreatedAt string `yaml:"createdAt"`
}

func (s *Server) listAgents(ctx context.Context) (*types.AgentList, error) {
	if s.data == nil {
		return &types.AgentList{}, nil
	}

	agents, err := s.data.Agents(ctx)
	if err != nil {
		return nil, err
	}
	return &types.AgentList{
		Agents: agents,
	}, nil
}

func (s *Server) listChats(ctx context.Context) (*types.ChatList, error) {
	mcpSession := mcp.SessionFromContext(ctx)

	manager, accountID, err := s.getManagerAndAccountID(mcpSession)
	if err != nil {
		return nil, err
	}

	sessions, err := manager.DB.FindByAccount(ctx, "thread", accountID)
	if err != nil {
		return nil, err
	}

	chats := make([]types.Chat, 0, len(sessions))
	for _, s := range sessions {
		chats = append(chats, chatFromSession(&s, accountID))
	}

	return &types.ChatList{
		Chats: chats,
	}, nil
}

func (s *Server) resourcesList(ctx context.Context, _ mcp.Message, _ mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	if err := s.ensureWatchers(ctx); err != nil {
		log.Debugf(ctx, "failed to refresh meta resource watchers: %v", err)
	}

	agents, err := s.listAgents(ctx)
	if err != nil {
		return nil, err
	}

	chats, err := s.listChats(ctx)
	if err != nil {
		return nil, err
	}

	files, err := s.listFiles(ctx)
	if err != nil {
		return nil, err
	}

	workflowsList, err := s.listWorkflows(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]mcp.Resource, 0, len(agents.Agents)+len(chats.Chats)+len(files.Resources)+len(workflowsList.Resources))

	for _, agent := range agents.Agents {
		resources = append(resources, agentResource(agent))
	}
	for _, chat := range chats.Chats {
		resources = append(resources, chatResource(chat))
	}
	resources = append(resources, files.Resources...)
	resources = append(resources, workflowsList.Resources...)

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].URI < resources[j].URI
	})

	return &mcp.ListResourcesResult{
		Resources: resources,
	}, nil
}

func (s *Server) resourcesRead(ctx context.Context, _ mcp.Message, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	switch {
	case strings.HasPrefix(request.URI, agentURIPrefix):
		return s.readAgentResource(ctx, request.URI)
	case strings.HasPrefix(request.URI, chatThreadPrefix):
		return s.readChatResource(ctx, request.URI)
	case strings.HasPrefix(request.URI, fileURIPrefix):
		return s.readFileResource(ctx, request.URI)
	case strings.HasPrefix(request.URI, workflowURIPrefix):
		return s.readWorkflowResource(request.URI)
	default:
		return nil, mcp.ErrRPCInvalidParams.WithMessage("unsupported resource URI: %s", request.URI)
	}
}

func (s *Server) resourcesSubscribe(ctx context.Context, msg mcp.Message, request mcp.SubscribeRequest) (*mcp.SubscribeResult, error) {
	s.trackSession(ctx, msg.Session)
	if err := s.ensureWatchers(ctx); err != nil {
		log.Debugf(ctx, "failed to refresh meta resource watchers before subscribe: %v", err)
	}

	if err := s.validateResourceExists(ctx, request.URI); err != nil {
		return nil, err
	}

	sessionID := sessionSubscriptionKey(ctx, msg.Session)
	if sessionID == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("session ID not found")
	}

	s.subscriptions.Subscribe(sessionID, msg.Session, request.URI)
	return &mcp.SubscribeResult{}, nil
}

func (s *Server) resourcesUnsubscribe(ctx context.Context, msg mcp.Message, request mcp.UnsubscribeRequest) (*mcp.UnsubscribeResult, error) {
	sessionID := sessionSubscriptionKey(ctx, msg.Session)
	if sessionID == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("session ID not found")
	}

	s.subscriptions.Unsubscribe(sessionID, request.URI)
	return &mcp.UnsubscribeResult{}, nil
}

func (s *Server) listFiles(ctx context.Context) (*mcp.ListResourcesResult, error) {
	mcpSession := mcp.SessionFromContext(ctx)

	manager, accountID, err := s.getManagerAndAccountID(mcpSession)
	if err != nil {
		return nil, err
	}

	sessions, err := manager.DB.FindByAccount(ctx, "thread", accountID)
	if err != nil {
		return nil, err
	}

	files := make([]mcp.Resource, 0)
	for _, chatSession := range sessions {
		cwd := strings.TrimSpace(chatSession.Cwd)
		if cwd == "" {
			cwd = defaultSessionCwd(chatSession.SessionID)
		}

		info, err := os.Stat(cwd)
		if err != nil || !info.IsDir() {
			continue
		}

		_ = filepath.WalkDir(cwd, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				return nil
			}

			fileInfo, err := d.Info()
			if err != nil || !fileInfo.Mode().IsRegular() {
				return nil
			}

			relPath, err := filepath.Rel(cwd, path)
			if err != nil || relPath == "." {
				return nil
			}

			mimeType := mime.TypeByExtension(filepath.Ext(relPath))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}

			files = append(files, mcp.Resource{
				URI:      fileURI(path),
				Name:     filepath.ToSlash(relPath),
				MimeType: mimeType,
				Size:     fileInfo.Size(),
				Meta:     fileMeta(chatSession, cwd, fileInfo),
			})
			return nil
		})
	}

	sort.Slice(files, func(i, j int) bool {
		sessionI, _ := files[i].Meta["sessionId"].(string)
		sessionJ, _ := files[j].Meta["sessionId"].(string)
		if sessionI == sessionJ {
			return files[i].Name < files[j].Name
		}
		return sessionI < sessionJ
	})

	return &mcp.ListResourcesResult{
		Resources: files,
	}, nil
}

func (s *Server) listWorkflows(ctx context.Context) (*mcp.ListResourcesResult, error) {
	workflowsPath := filepath.Join(".", workflowsDir)
	return &mcp.ListResourcesResult{
		Resources: workflows.ListWorkflowResources(ctx, workflowsPath),
	}, nil
}

func (s *Server) readAgentResource(ctx context.Context, uri string) (*mcp.ReadResourceResult, error) {
	agentID, err := parseAgentURI(uri)
	if err != nil {
		return nil, err
	}

	agents, err := s.listAgents(ctx)
	if err != nil {
		return nil, err
	}

	for _, agent := range agents.Agents {
		if agent.ID != agentID {
			continue
		}

		data, err := json.Marshal(agent)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal agent %s: %w", agentID, err)
		}

		text := string(data)
		return &mcp.ReadResourceResult{
			Contents: []mcp.ResourceContent{
				{
					URI:      uri,
					Name:     resourceDisplayName(agent.Name, agent.ID),
					MIMEType: types.AgentMimeType,
					Text:     &text,
					Meta:     structToMeta(agent),
				},
			},
		}, nil
	}

	return nil, mcp.ErrRPCInvalidParams.WithMessage("agent not found: %s", uri)
}

func (s *Server) readChatResource(ctx context.Context, uri string) (*mcp.ReadResourceResult, error) {
	chatID, err := parseChatURI(uri)
	if err != nil {
		return nil, err
	}

	chats, err := s.listChats(ctx)
	if err != nil {
		return nil, err
	}

	for _, chat := range chats.Chats {
		if chat.ID != chatID {
			continue
		}

		data, err := json.Marshal(chat)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal chat %s: %w", chatID, err)
		}

		text := string(data)
		return &mcp.ReadResourceResult{
			Contents: []mcp.ResourceContent{
				{
					URI:      uri,
					Name:     resourceDisplayName(chat.Title, chat.ID),
					MIMEType: types.SessionMimeType,
					Text:     &text,
					Meta:     structToMeta(chat),
				},
			},
		}, nil
	}

	return nil, mcp.ErrRPCInvalidParams.WithMessage("chat not found: %s", uri)
}

func (s *Server) readFileResource(ctx context.Context, uri string) (*mcp.ReadResourceResult, error) {
	path, err := parseAbsoluteFileURI(uri)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("file not found: %s", uri)
	}

	chatSession, sessionCwd, relPath, err := s.sessionForFile(ctx, path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(relPath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if i := strings.IndexByte(mimeType, ';'); i >= 0 {
		mimeType = strings.TrimSpace(mimeType[:i])
	}

	rc := mcp.ResourceContent{
		URI:      uri,
		Name:     relPath,
		MIMEType: mimeType,
		Meta:     fileMeta(*chatSession, sessionCwd, info),
	}
	if _, isImage := types.ImageMimeTypes[mimeType]; isImage {
		rc.Blob = new(base64.StdEncoding.EncodeToString(content))
	} else if _, isPDF := types.PDFMimeTypes[mimeType]; isPDF {
		rc.Blob = new(base64.StdEncoding.EncodeToString(content))
	} else {
		rc.Text = new(string(content))
	}

	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{rc},
	}, nil
}

func (s *Server) readWorkflowResource(uri string) (*mcp.ReadResourceResult, error) {
	workflowName, err := parseWorkflowURI(uri)
	if err != nil {
		return nil, err
	}

	workflowPath := filepath.Join(".", workflowsDir, workflowName+".md")
	contentBytes, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("workflow not found: %s", uri)
	}

	content := string(contentBytes)
	meta, err := parseWorkflowFrontmatter(content)
	if err != nil {
		meta = workflowMeta{}
	}

	resourceMeta := map[string]any{}
	if meta.Name != "" {
		resourceMeta["name"] = meta.Name
	}
	if meta.CreatedAt != "" {
		resourceMeta["createdAt"] = meta.CreatedAt
	}

	resourceContent := mcp.ResourceContent{
		URI:      uri,
		Name:     workflowName,
		MIMEType: "text/markdown",
		Text:     &content,
	}
	if len(resourceMeta) > 0 {
		resourceContent.Meta = resourceMeta
	}

	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{resourceContent},
	}, nil
}

func (s *Server) validateResourceExists(ctx context.Context, uri string) error {
	switch {
	case strings.HasPrefix(uri, agentURIPrefix):
		agentID, err := parseAgentURI(uri)
		if err != nil {
			return err
		}
		agents, err := s.listAgents(ctx)
		if err != nil {
			return err
		}
		for _, agent := range agents.Agents {
			if agent.ID == agentID {
				return nil
			}
		}
		return mcp.ErrRPCInvalidParams.WithMessage("agent not found: %s", uri)
	case strings.HasPrefix(uri, chatThreadPrefix):
		chatID, err := parseChatURI(uri)
		if err != nil {
			return err
		}
		chats, err := s.listChats(ctx)
		if err != nil {
			return err
		}
		for _, chat := range chats.Chats {
			if chat.ID == chatID {
				return nil
			}
		}
		return mcp.ErrRPCInvalidParams.WithMessage("chat not found: %s", uri)
	case strings.HasPrefix(uri, fileURIPrefix):
		path, err := parseAbsoluteFileURI(uri)
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			return mcp.ErrRPCInvalidParams.WithMessage("file not found: %s", uri)
		}
		_, _, _, err = s.sessionForFile(ctx, path)
		return err
	case strings.HasPrefix(uri, workflowURIPrefix):
		workflowName, err := parseWorkflowURI(uri)
		if err != nil {
			return err
		}
		workflowPath := filepath.Join(".", workflowsDir, workflowName+".md")
		if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
			return mcp.ErrRPCInvalidParams.WithMessage("workflow not found: %s", uri)
		} else if err != nil {
			return fmt.Errorf("failed to stat workflow %s: %w", uri, err)
		}
		return nil
	default:
		return mcp.ErrRPCInvalidParams.WithMessage("unsupported resource URI: %s", uri)
	}
}

func (s *Server) sessionForFile(ctx context.Context, targetPath string) (*session.Session, string, string, error) {
	mcpSession := mcp.SessionFromContext(ctx)

	manager, accountID, err := s.getManagerAndAccountID(mcpSession)
	if err != nil {
		return nil, "", "", err
	}

	sessions, err := manager.DB.FindByAccount(ctx, "thread", accountID)
	if err != nil {
		return nil, "", "", err
	}

	cleanPath := filepath.Clean(targetPath)
	for i := range sessions {
		chatSession := &sessions[i]
		cwd := strings.TrimSpace(chatSession.Cwd)
		if cwd == "" {
			cwd = defaultSessionCwd(chatSession.SessionID)
		}

		absCwd, err := filepath.Abs(cwd)
		if err != nil {
			continue
		}

		relPath, err := filepath.Rel(absCwd, cleanPath)
		if err != nil || relPath == "." || relPath == "" || filepath.IsAbs(relPath) {
			continue
		}
		if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
			continue
		}

		return chatSession, absCwd, filepath.ToSlash(relPath), nil
	}

	return nil, "", "", mcp.ErrRPCInvalidParams.WithMessage("file not found: %s", fileURI(cleanPath))
}

func agentResource(agent types.AgentDisplay) mcp.Resource {
	return mcp.Resource{
		URI:         agentURI(agent.ID),
		Name:        resourceDisplayName(agent.Name, agent.ID),
		Title:       agent.Name,
		Description: agent.Description,
		MimeType:    types.AgentMimeType,
		Meta:        structToMeta(agent),
	}
}

func chatResource(chat types.Chat) mcp.Resource {
	return mcp.Resource{
		URI:         chatURI(chat.ID),
		Name:        resourceDisplayName(chat.Title, chat.ID),
		Title:       chat.Title,
		Description: chat.Title,
		MimeType:    types.SessionMimeType,
		Meta:        structToMeta(chat),
	}
}

func structToMeta(in any) map[string]any {
	meta := map[string]any{}
	if err := mcp.JSONCoerce(in, &meta); err != nil {
		return nil
	}
	return meta
}

func fileMeta(chatSession session.Session, cwd string, info os.FileInfo) map[string]any {
	meta := map[string]any{
		"sessionId":        chatSession.SessionID,
		"sessionCreatedAt": chatSession.CreatedAt.UTC().Format(time.RFC3339Nano),
		"sessionCwd":       cwd,
		"modifiedAt":       info.ModTime().UTC().Format(time.RFC3339Nano),
	}
	if chatSession.Description != "" {
		meta["sessionTitle"] = chatSession.Description
	}
	return meta
}

func agentURI(id string) string {
	return agentURIPrefix + url.PathEscape(id)
}

func chatURI(id string) string {
	return chatThreadPrefix + url.PathEscape(id)
}

func parseAgentURI(uri string) (string, error) {
	if !strings.HasPrefix(uri, agentURIPrefix) {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid agent URI format, expected %s{id}", agentURIPrefix)
	}

	id := strings.TrimPrefix(uri, agentURIPrefix)
	if id == "" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("agent ID is required")
	}

	id, err := url.PathUnescape(id)
	if err != nil {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid agent URI encoding: %s", uri)
	}
	if strings.Contains(id, "/") {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid agent ID in URI: %s", uri)
	}

	return id, nil
}

func parseChatURI(uri string) (string, error) {
	if !strings.HasPrefix(uri, chatThreadPrefix) {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid chat URI format, expected %s{id}", chatThreadPrefix)
	}

	id := strings.TrimPrefix(uri, chatThreadPrefix)
	if id == "" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("chat ID is required")
	}

	id, err := url.PathUnescape(id)
	if err != nil {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid chat URI encoding: %s", uri)
	}
	if strings.Contains(id, "/") {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid chat ID in URI: %s", uri)
	}

	return id, nil
}

func parseAbsoluteFileURI(uri string) (string, error) {
	if !strings.HasPrefix(uri, fileURIPrefix) {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid file URI, expected file:///absolute/path")
	}

	raw := strings.TrimPrefix(uri, fileURIPrefix)
	if raw == "" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("file path is required")
	}

	if unescaped, err := url.PathUnescape(raw); err == nil {
		raw = unescaped
	}

	raw = strings.TrimPrefix(raw, "/")
	path := "/" + raw
	if runtime.GOOS == "windows" && len(raw) >= 2 && raw[1] == ':' {
		path = raw
	}

	path = filepath.Clean(filepath.FromSlash(path))
	if !filepath.IsAbs(path) {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid file URI, expected absolute path: %s", uri)
	}

	return path, nil
}

func parseWorkflowURI(uri string) (string, error) {
	if !strings.HasPrefix(uri, workflowURIPrefix) {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid workflow URI format, expected workflow:///name")
	}

	workflowName := strings.TrimPrefix(uri, workflowURIPrefix)
	if workflowName == "" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("workflow name is required")
	}

	workflowName = strings.TrimSuffix(workflowName, ".md")
	clean := filepath.Clean(workflowName)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid workflow URI: %s", uri)
	}

	return filepath.ToSlash(clean), nil
}

func parseWorkflowFrontmatter(content string) (workflowMeta, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return workflowMeta{}, nil
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		return workflowMeta{}, fmt.Errorf("frontmatter missing closing delimiter")
	}

	frontmatterYAML := strings.Join(lines[1:endIdx], "\n")
	var meta workflowMeta
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &meta); err != nil {
		return workflowMeta{}, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return meta, nil
}

func resourceDisplayName(preferred, fallback string) string {
	if strings.TrimSpace(preferred) != "" {
		return preferred
	}
	return fallback
}

func defaultSessionCwd(sessionID string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(cwd, "sessions", types.SanitizeSessionDirectoryName(sessionID))
}

func fileURI(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	return fileURIPrefix + strings.TrimPrefix(filepath.ToSlash(absPath), "/")
}
