package session

import (
	"context"
	"fmt"

	"github.com/nanobot-ai/nanobot/pkg/gormdsn"
	"gorm.io/gorm"
)

const (
	StoreSessionKey = "sessionStore"
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

func (s *Store) Get(ctx context.Context, id string) (*Session, error) {
	var session Session
	err := s.db.WithContext(ctx).Where("session_id = ?", id).First(&session).Error
	return &session, err
}

func (s *Store) FindByAccount(ctx context.Context, accountID string) ([]Session, error) {
	var sessions []Session
	err := s.db.WithContext(ctx).Where("account_id = ?", accountID).
		Select("id", "created_at", "session_id", "account_id").Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *Store) FindByTypeAndParentID(ctx context.Context, sessionType, parentID string) ([]Session, error) {
	var sessions []Session
	err := s.db.WithContext(ctx).Where("type = ? AND parent_id = ?", sessionType, parentID).Find(&sessions).Error
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
