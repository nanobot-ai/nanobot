package meta

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/nanobot-ai/nanobot/pkg/fswatch"
	"github.com/nanobot-ai/nanobot/pkg/log"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/session"
	"github.com/nanobot-ai/nanobot/pkg/sessiondata"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/version"
)

type Server struct {
	tools         mcp.ServerTools
	data          *sessiondata.Data
	subscriptions *fswatch.SubscriptionManager

	fileWatchersLock sync.Mutex
	fileWatchers     map[string]*fswatch.Watcher

	workflowsWatcherLock sync.Mutex
	workflowsWatcher     *fswatch.Watcher

	sessionLock sync.Mutex
	sessions    map[string]trackedSession

	managerSubscriptionsLock sync.Mutex
	managerSubscriptions     map[*session.Manager]func()
}

type trackedSession struct {
	session   *mcp.Session
	accountID string
}

func NewServer(data *sessiondata.Data) *Server {
	s := &Server{
		data:                 data,
		subscriptions:        fswatch.NewSubscriptionManager(context.Background()),
		fileWatchers:         map[string]*fswatch.Watcher{},
		sessions:             map[string]trackedSession{},
		managerSubscriptions: map[*session.Manager]func(){},
	}

	s.tools = mcp.NewServerTools(
		mcp.NewServerTool("update_chat", "Update fields of a give chat thread", s.updateChat),
	)

	return s
}

func (s *Server) OnMessage(ctx context.Context, msg mcp.Message) {
	switch msg.Method {
	case "initialize":
		mcp.Invoke(ctx, msg, s.initialize)
	case "notifications/initialized":
		// nothing to do
	case "notifications/cancelled":
		mcp.HandleCancelled(ctx, msg)
	case "tools/list":
		mcp.Invoke(ctx, msg, s.tools.List)
	case "tools/call":
		mcp.Invoke(ctx, msg, s.tools.Call)
	case "resources/list":
		mcp.Invoke(ctx, msg, s.resourcesList)
	case "resources/read":
		mcp.Invoke(ctx, msg, s.resourcesRead)
	case "resources/subscribe":
		mcp.Invoke(ctx, msg, s.resourcesSubscribe)
	case "resources/unsubscribe":
		mcp.Invoke(ctx, msg, s.resourcesUnsubscribe)
	default:
		msg.SendError(ctx, mcp.ErrRPCMethodNotFound.WithMessage("%v", msg.Method))
	}
}

func (s *Server) initialize(ctx context.Context, msg mcp.Message, params mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	if !types.IsUISession(ctx) {
		s.tools = mcp.NewServerTools()
		return &mcp.InitializeResult{
			ProtocolVersion: params.ProtocolVersion,
			ServerInfo: mcp.ServerInfo{
				Name:    version.Name,
				Version: version.Get().String(),
			},
		}, nil
	}

	s.trackSession(ctx, msg.Session)
	s.ensureManagerEventSubscription(ctx)
	if err := s.ensureWatchers(ctx); err != nil {
		log.Debugf(ctx, "failed to initialize meta resource watchers: %v", err)
	}

	return &mcp.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Tools: &mcp.ToolsServerCapability{},
			Resources: &mcp.ResourcesServerCapability{
				Subscribe:   true,
				ListChanged: true,
			},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    version.Name,
			Version: version.Get().String(),
		},
	}, nil
}

func (s *Server) trackSession(ctx context.Context, session *mcp.Session) {
	if session == nil {
		return
	}

	sessionID := sessionSubscriptionKey(ctx, session)
	if sessionID == "" {
		return
	}

	s.subscriptions.AddSession(sessionID, session)
	accountID := sessionAccountID(ctx, session)

	s.sessionLock.Lock()
	if _, ok := s.sessions[sessionID]; !ok {
		s.sessions[sessionID] = trackedSession{
			session:   session,
			accountID: accountID,
		}
		// Auto-cleanup when the transport closes so we do not leak watchers/subscriptions.
		context.AfterFunc(session.Context(), func() {
			s.untrackSession(sessionID)
		})
	} else {
		tracked := s.sessions[sessionID]
		if tracked.session == nil {
			tracked.session = session
		}
		if tracked.accountID == "" {
			tracked.accountID = accountID
		}
		s.sessions[sessionID] = tracked
	}
	s.sessionLock.Unlock()
}

func (s *Server) untrackSession(sessionID string) {
	var closeWatchers bool

	s.sessionLock.Lock()
	delete(s.sessions, sessionID)
	if len(s.sessions) == 0 {
		closeWatchers = true
	}
	s.sessionLock.Unlock()

	if closeWatchers {
		s.closeWatchers()
		s.closeManagerSubscriptions()
	}
}

func (s *Server) ensureManagerEventSubscription(ctx context.Context) {
	mcpSession := mcp.SessionFromContext(ctx)
	if mcpSession != nil {
		s.trackSession(ctx, mcpSession)
	}

	manager, _, err := s.getManagerAndAccountID(mcpSession)
	if err != nil || manager == nil {
		return
	}

	s.managerSubscriptionsLock.Lock()
	defer s.managerSubscriptionsLock.Unlock()
	if _, ok := s.managerSubscriptions[manager]; ok {
		return
	}

	// Session manager events drive list_changed and resource updated notifications.
	unsubscribe := manager.SubscribeEvents(s.handleSessionEvent)
	s.managerSubscriptions[manager] = unsubscribe
}

func (s *Server) closeManagerSubscriptions() {
	s.managerSubscriptionsLock.Lock()
	defer s.managerSubscriptionsLock.Unlock()

	for manager, unsubscribe := range s.managerSubscriptions {
		if unsubscribe != nil {
			unsubscribe()
		}
		delete(s.managerSubscriptions, manager)
	}
}

func (s *Server) handleSessionEvent(event session.SessionEvent) {
	if event.SessionType != "thread" || event.SessionID == "" {
		return
	}

	switch event.Type {
	case session.SessionEventCreated:
		s.sendListChangedForAccount(event.AccountID)
	case session.SessionEventDeleted:
		uri := chatURI(event.SessionID)
		s.subscriptions.SendResourceUpdatedNotification(uri)
		s.subscriptions.AutoUnsubscribe(uri)
		s.sendListChangedForAccount(event.AccountID)
	case session.SessionEventUpdated:
		s.subscriptions.SendResourceUpdatedNotification(chatURI(event.SessionID))
	}
}

func (s *Server) sendListChangedForAccount(accountID string) {
	s.sessionLock.Lock()
	// Copy recipients under lock, then send outside the lock to avoid blocking all sessions.
	sessions := make([]*mcp.Session, 0, len(s.sessions))
	for _, tracked := range s.sessions {
		if tracked.session == nil {
			continue
		}
		if accountID != "" && tracked.accountID != accountID {
			continue
		}
		sessions = append(sessions, tracked.session)
	}
	s.sessionLock.Unlock()

	for _, session := range sessions {
		err := session.Send(context.Background(), mcp.Message{
			JSONRPC: "2.0",
			Method:  "notifications/resources/list_changed",
		})
		if err != nil && !errors.Is(err, mcp.ErrNoReader) {
			log.Errorf(context.Background(), "failed to send list_changed notification: %v", err)
		}
	}
}

func sessionSubscriptionKey(ctx context.Context, session *mcp.Session) string {
	sessionID, _ := types.GetSessionAndAccountID(ctx)
	if sessionID == "" && session != nil {
		sessionID = session.ID()
	}
	if sessionID == "" && session != nil {
		sessionID = fmt.Sprintf("session-%p", session)
	}
	return sessionID
}

func sessionAccountID(ctx context.Context, session *mcp.Session) string {
	_, accountID := types.GetSessionAndAccountID(ctx)
	if accountID == "" && session != nil {
		session.Get(types.AccountIDSessionKey, &accountID)
	}
	return accountID
}
