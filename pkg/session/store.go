package session

import (
	"context"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/gormdsn"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"gorm.io/gorm"
)

const (
	ManagerSessionKey = "sessionManager"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func NewStoreFromDSN(dsn string) (*Store, error) {
	db, err := gormdsn.NewDBFromDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	if err := db.AutoMigrate(&Session{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Create(ctx context.Context, session *Session) error {
	if session.SessionID == "" {
		session.SessionID = session.State.ID
	}
	if session.SessionID == "" {
		session.SessionID = uuid.String()
		session.State.ID = session.SessionID
	}
	if session.State.ID == "" {
		session.State.ID = session.SessionID
	}
	if session.Type == "" {
		session.Type = "thread"
	}
	return s.db.WithContext(ctx).Create(session).Error
}

func (s *Store) Update(ctx context.Context, session *Session) error {
	return s.db.WithContext(ctx).Save(session).Error
}

func (s *Store) FindByPrefix(ctx context.Context, sessionIDPrefix string) ([]Session, error) {
	var sessions []Session
	if sessionIDPrefix == "last" {
		err := s.db.WithContext(ctx).Order("updated_at desc").First(&sessions).Error
		return sessions, err
	}
	err := s.db.WithContext(ctx).Where("session_id LIKE ?", sessionIDPrefix+"%").Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	return s.db.WithContext(ctx).Where("session_id = ?", id).Delete(&Session{}).Error
}

func (s *Store) Get(ctx context.Context, id string) (*Session, error) {
	var session Session
	err := s.db.WithContext(ctx).Where("session_id = ?", id).First(&session).Error
	return &session, err
}

func (s *Store) GetByIDByAccountID(ctx context.Context, id, accountID string) (*Session, error) {
	var session Session
	err := s.db.WithContext(ctx).Where("session_id = ? and account_id = ?", id, accountID).First(&session).Error
	return &session, err
}

func (s *Store) FindByAccount(ctx context.Context, sessionType, accountID string) ([]Session, error) {
	var sessions []Session
	err := s.db.WithContext(ctx).Where("type = ? and account_id = ?", sessionType, accountID).
		Order("created_at desc").Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *Store) List(ctx context.Context) ([]Session, error) {
	var sessions []Session
	err := s.db.WithContext(ctx).Order("updated_at desc").Find(&sessions).Error
	return sessions, err
}
