package workspace

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/gormdsn"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// WorkspaceRecord represents a workspace stored in the database
type WorkspaceRecord struct {
	gorm.Model
	// UUID is a unique identifier for the workspace
	UUID string `json:"uuid" gorm:"uniqueIndex;not null"`
	// AccountID is the ID of the account that owns this workspace
	AccountID string `json:"accountID" gorm:"index;not null"`
	// Name is the name of the workspace
	Name string `json:"name" gorm:"not null"`
	// Order is the display order of the workspace
	Order int `json:"order" gorm:"default:0"`
	// Color is the color theme for the workspace
	Color string `json:"color,omitempty"`
	// Icon is the icon identifier for the workspace (deprecated, use Icons instead)
	Icon string `json:"icon,omitempty"`
	// Icons is an array of icon definitions for the workspace
	Icons datatypes.JSON `json:"icons"`
	// Attributes is a JSON object containing workspace-specific attributes
	Attributes datatypes.JSON `json:"attributes"`
	// ParentID is the workspace UUID this workspace is created from
	ParentID *string `json:"parentID,omitempty"`
	// BaseURI is the external resource ID of the overlay base of the workspace
	BaseURI string `json:"baseURI,omitempty"`
	// SessionID the associated session ID for this workspace
	SessionID string `json:"sessionID,omitempty"`
}

// TableName overrides the default table name to be "workspaces"
func (WorkspaceRecord) TableName() string {
	return "workspaces"
}

type Store struct {
	// db is the database connection
	db *gorm.DB
}

// NewStore creates a new workspace store with the given database connection
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

// Init initializes the workspace store by migrating the schema
func (s *Store) Init() error {
	return s.db.AutoMigrate(&WorkspaceRecord{})
}

// Create creates a new workspace in the database
func (s *Store) Create(ctx context.Context, workspace *WorkspaceRecord) error {
	return s.db.WithContext(ctx).Create(workspace).Error
}

// Get retrieves a workspace by its ID
func (s *Store) Get(ctx context.Context, id uint) (*WorkspaceRecord, error) {
	var workspace WorkspaceRecord
	err := s.db.WithContext(ctx).First(&workspace, id).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func (s *Store) GetByUUIDAndAccountID(ctx context.Context, uuid, accountID string) (*WorkspaceRecord, error) {
	var workspace WorkspaceRecord
	err := s.db.WithContext(ctx).Where("uuid = ? and account_id = ?", uuid, accountID).First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

// Update updates a workspace in the database
func (s *Store) Update(ctx context.Context, workspace *WorkspaceRecord) error {
	return s.db.WithContext(ctx).Save(workspace).Error
}

// Delete deletes a workspace by its ID
func (s *Store) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&WorkspaceRecord{}, id).Error
}

// FindByAccountID retrieves all workspaces for a given account ID
func (s *Store) FindByAccountID(ctx context.Context, accountID string) ([]WorkspaceRecord, error) {
	var workspaces []WorkspaceRecord
	err := s.db.WithContext(ctx).Where("account_id = ? and parent_id is null", accountID).Order("`order` asc, created_at desc").Find(&workspaces).Error
	if err != nil {
		return nil, err
	}
	return workspaces, nil
}

// FindByParentID retrieves all workspace records for a given parent ID
func (s *Store) FindByParentID(ctx context.Context, parentID string) ([]WorkspaceRecord, error) {
	var workspaces []WorkspaceRecord
	err := s.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&workspaces).Error
	if err != nil {
		return nil, err
	}
	return workspaces, nil
}

// WorkspaceWithSession combines workspace and session data
type WorkspaceWithSession struct {
	WorkspaceRecord
	SessionDescription string `gorm:"column:session_description"`
}

// FindByParentIDWithSessions retrieves all workspace records with their session data for a given parent ID
func (s *Store) FindByParentIDWithSessions(ctx context.Context, parentID string) ([]WorkspaceWithSession, error) {
	var results []WorkspaceWithSession
	err := s.db.WithContext(ctx).
		Table("workspaces").
		Select("workspaces.*, sessions.description as session_description").
		Joins("LEFT JOIN sessions ON sessions.session_id = workspaces.session_id").
		Where("workspaces.parent_id = ? AND sessions.deleted_at is null", parentID).
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// GetBySessionID retrieves a workspace by its session ID
func (s *Store) GetBySessionID(ctx context.Context, sessionID string) (*WorkspaceRecord, error) {
	var workspace WorkspaceRecord
	err := s.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}
