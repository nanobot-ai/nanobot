package session

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func NewStoreFromDSN(dsn string) (*Store, error) {
	var dialector gorm.Dialector

	switch {
	case strings.HasPrefix(dsn, "sqlite:") || strings.HasSuffix(dsn, ".db") || strings.Contains(dsn, ":memory:"):
		dsn = strings.TrimPrefix(dsn, "sqlite:")
		dialector = sqlite.Open(dsn)
	case strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://"):
		dialector = postgres.Open(dsn)
	case strings.HasPrefix(dsn, "mysql://") || strings.Contains(dsn, "@tcp("):
		dsn = strings.TrimPrefix(dsn, "mysql://")
		dialector = mysql.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database type in DSN: %s", dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(&Session{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Create(session *Session) error {
	return s.db.Create(session).Error
}

func (s *Store) Update(session *Session) error {
	return s.db.Save(session).Error
}

func (s *Store) FindByPrefix(sessionIDPrefix string) ([]Session, error) {
	var sessions []Session
	err := s.db.Where("session_id LIKE ?", sessionIDPrefix+"%").Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *Store) Get(id string) (*Session, error) {
	var session Session
	err := s.db.Where("session_id = ?", id).First(&session).Error
	return &session, err
}

func (s *Store) FindByTypeAndParentID(sessionType, parentID string) ([]Session, error) {
	var sessions []Session
	err := s.db.Where("type = ? AND parent_id = ?", sessionType, parentID).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *Store) List() ([]Session, error) {
	var sessions []Session
	err := s.db.Order("updated_at desc").Find(&sessions).Error
	return sessions, err
}
