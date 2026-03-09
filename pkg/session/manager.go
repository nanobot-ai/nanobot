package session

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var sessionTracer = otel.Tracer("nanobot/session")

func NewManager(dsn string) (*Manager, error) {
	store, err := NewStoreFromDSN(dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		ctx:          ctx,
		close:        cancel,
		DB:           store,
		root:         &Session{},
		liveSessions: make(map[string]liveSession),
	}, nil
}

type Manager struct {
	ctx   context.Context
	close context.CancelFunc
	DB    *Store
	root  *Session

	liveSessionsLock sync.Mutex
	liveSessions     map[string]liveSession
}

type liveSession struct {
	session *mcp.ServerSession
	count   int
	cancel  context.CancelFunc
}

func (m *Manager) newRecord(id, accountID string) *Session {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	return &Session{
		SessionID: id,
		AccountID: accountID,
		Cwd:       cwd,
	}
}

func (m *Manager) loadAttributesFromRecord(stored *Session, session *mcp.ServerSession) {
	session.GetSession().Set(types.DescriptionSessionKey, stored.Description)
	session.GetSession().Set(types.AccountIDSessionKey, stored.AccountID)
}

func (m *Manager) saveAttributesToRecord(stored *Session, session *mcp.ServerSession) error {
	var (
		config types.Config
	)

	session.GetSession().Get(types.DescriptionSessionKey, &stored.Description)
	session.GetSession().Get(types.ConfigSessionKey, &config)

	stored.Config = ConfigWrapper(config)
	return nil
}

func (m *Manager) Store(ctx context.Context, id string, session *mcp.ServerSession) error {
	if id == "" {
		return nil
	}

	ctx, span := sessionTracer.Start(ctx, "session.store", trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session.id", id)))
	defer span.End()

	var accountID string
	session.GetSession().Get(types.AccountIDSessionKey, &accountID)

	var create bool
	stored, err := m.DB.Get(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		stored = m.newRecord(id, accountID)
		create = true
	} else if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if stored.AccountID != accountID {
		err := fmt.Errorf("session %s not found for account %s", id, accountID)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if err := m.saveAttributesToRecord(stored, session); err != nil {
		err = fmt.Errorf("failed to save attributes to session record: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	state, err := session.GetSession().State()
	if err != nil {
		err = fmt.Errorf("failed to get session state: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	stored.State = *(*State)(state)

	if create {
		span.SetAttributes(attribute.Bool("session.create", true))
		if err := m.DB.Create(ctx, stored); err != nil {
			err = fmt.Errorf("failed to create session record: %w", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		m.liveSessionsLock.Lock()
		live, ok := m.liveSessions[id]
		if ok {
			if live.session != nil {
				live.session.Close(false)
			}
			live.count++
			live.session = session

			m.liveSessions[id] = live
		} else {
			m.liveSessions[id] = liveSession{
				session: session,
				count:   1,
			}
		}
		m.liveSessionsLock.Unlock()
	} else {
		span.SetAttributes(attribute.Bool("session.create", false))
		if err := m.DB.Update(ctx, stored); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
	}

	m.loadAttributesFromRecord(stored, session)
	span.SetStatus(codes.Ok, "")
	return nil
}

func (m *Manager) ExtractID(req *http.Request) string {
	id := req.Header.Get("Mcp-Session-Id")
	if id != "" {
		return id
	}
	id = req.Header.Get("X-Nanobot-Session-Id")
	if id != "" {
		return id
	}
	parts := strings.Split(req.URL.Path, "/")
	for i, part := range parts {
		if i > 0 && parts[i-1] == "agents" {
			continue
		}

		if uuid.ValidUUID(part) {
			return part
		}
	}
	return ""
}

func checkAccount(ctx context.Context, serverSession *mcp.ServerSession) bool {
	var (
		account        string
		nanobotContext = types.NanobotContext(ctx)
	)
	serverSession.GetSession().Get(types.AccountIDSessionKey, &account)
	return account == nanobotContext.User.ID
}

func (m *Manager) Acquire(ctx context.Context, server mcp.MessageHandler, id string) (ret *mcp.ServerSession, found bool, retErr error) {
	m.liveSessionsLock.Lock()
	live, ok := m.liveSessions[id]
	if ok {
		select {
		case <-live.session.GetSession().Context().Done():
			m.liveSessionsLock.Unlock()
			return nil, false, nil
		default:
		}

		if !checkAccount(ctx, live.session) {
			m.liveSessionsLock.Unlock()
			return nil, false, nil
		}

		live.count++
		m.liveSessions[id] = live
		m.liveSessionsLock.Unlock()
		return live.session, true, nil
	}
	m.liveSessionsLock.Unlock()

	serverSession, ok, err := m.loadSessionFromDatabase(ctx, server, id)
	if err != nil || !ok {
		return nil, false, err
	}

	if !checkAccount(ctx, serverSession) {
		return nil, false, nil
	}

	m.liveSessionsLock.Lock()
	live, ok = m.liveSessions[id]
	if ok {
		serverSession.Close(false)
		live.count++
		m.liveSessions[id] = live
		m.liveSessionsLock.Unlock()
		return live.session, true, nil
	}
	m.liveSessions[id] = liveSession{
		session: serverSession,
		count:   1,
	}
	m.liveSessionsLock.Unlock()

	return serverSession, true, err
}

func (m *Manager) Release(session *mcp.ServerSession) {
	m.liveSessionsLock.Lock()
	defer m.liveSessionsLock.Unlock()

	live, ok := m.liveSessions[session.ID()]
	if ok {
		live.count--
		if live.count == 0 {
			ctx, cancel := context.WithCancel(context.Background())
			live.cancel = cancel

			go func(ctx context.Context, sessionID string) {
				defer cancel()
				select {
				case <-ctx.Done():
					return
				case <-time.After(10 * time.Second):
				}

				m.liveSessionsLock.Lock()
				defer m.liveSessionsLock.Unlock()

				live, ok := m.liveSessions[sessionID]
				if ok && live.count == 0 {
					delete(m.liveSessions, sessionID)
					live.session.Close(false)
				}
			}(ctx, session.ID())
		} else if live.cancel != nil {
			live.cancel()
			live.cancel = nil
		}

		m.liveSessions[session.ID()] = live
	} else {
		session.Close(false)
	}
}

func (m *Manager) loadSessionFromDatabase(ctx context.Context, server mcp.MessageHandler, id string) (*mcp.ServerSession, bool, error) {
	storedSession, err := m.DB.Get(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	if storedSession.State.Attributes == nil {
		storedSession.State.Attributes = make(map[string]any)
	} else {
		storedSession.State.Attributes[".keys"] = slices.Collect(maps.Keys(storedSession.State.Attributes))
	}

	serverSession, err := mcp.NewExistingServerSession(m.ctx,
		mcp.SessionState(storedSession.State), server)
	if err != nil {
		return nil, false, err
	}

	m.loadAttributesFromRecord(storedSession, serverSession)
	return serverSession, true, nil
}

func (m *Manager) LoadAndDelete(ctx context.Context, server mcp.MessageHandler, id string) (*mcp.ServerSession, bool, error) {
	session, found, err := m.Acquire(ctx, server, id)
	if !found || err != nil {
		return session, found, err
	}
	defer m.Release(session)

	err = m.DB.Delete(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("failed to delete session: %w", err)
	}
	return session, true, nil
}
