package resources

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/gormdsn"
	"gorm.io/gorm"
)

type Store struct {
	// db is the database connection
	db *gorm.DB
}

// NewStore creates a new artifact store with the given database connection
func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func NewStoreFromDSN(dsn string) (*Store, error) {
	db, err := gormdsn.NewDBFromDSN(dsn)
	if err != nil {
		return nil, err
	}
	s := NewStore(db)
	return s, s.Init()
}

// Init initializes the artifact store by migrating the schema
func (s *Store) Init() error {
	return s.db.AutoMigrate(&Resource{})
}

// Create creates a new artifact in the database
func (s *Store) Create(ctx context.Context, artifact *Resource) error {
	return s.db.WithContext(ctx).Create(artifact).Error
}

// Get retrieves an artifact by its ID
func (s *Store) Get(ctx context.Context, id uint) (*Resource, error) {
	var artifact Resource
	err := s.db.WithContext(ctx).First(&artifact, id).Error
	if err != nil {
		return nil, err
	}
	return &artifact, nil
}

func (s *Store) GetByUUIDAndAccountID(ctx context.Context, uuid, accountID string) (*Resource, error) {
	var artifact Resource
	err := s.db.WithContext(ctx).Where("uuid = ? and account_id = ?", uuid, accountID).First(&artifact).Error
	if err != nil {
		return nil, err
	}
	return &artifact, nil
}

// GetByNameSessionIDAndAccountID retrieves an artifact by name, session ID, and account ID
func (s *Store) GetByNameSessionIDAndAccountID(ctx context.Context, name, sessionID, accountID string) (*Resource, error) {
	var artifact Resource
	err := s.db.WithContext(ctx).Where("name = ? and session_id = ? and account_id = ?", name, sessionID, accountID).First(&artifact).Error
	if err != nil {
		return nil, err
	}
	return &artifact, nil
}

// Update updates an existing artifact in the database
func (s *Store) Update(ctx context.Context, artifact *Resource) error {
	return s.db.WithContext(ctx).Save(artifact).Error
}

// Delete deletes an artifact by its ID
func (s *Store) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&Resource{}, id).Error
}

// FindBySessionID retrieves all artifacts for a given session ID
func (s *Store) FindBySessionID(ctx context.Context, sessionID string) ([]Resource, error) {
	var artifacts []Resource
	err := s.db.WithContext(ctx).Where("session_id = ?", sessionID).Find(&artifacts).Error
	if err != nil {
		return nil, err
	}
	return artifacts, nil
}

// List retrieves all artifacts
func (s *Store) List(ctx context.Context) ([]Resource, error) {
	var artifacts []Resource
	err := s.db.WithContext(ctx).Find(&artifacts).Error
	return artifacts, err
}
