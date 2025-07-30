package agentbuilder

import (
	"context"

	"github.com/nanobot-ai/nanobot/pkg/gormdsn"
	"gorm.io/gorm"
)

type Store struct {
	// db is the database connection
	db *gorm.DB
}

// NewStore creates a new agent store with the given database connection
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

// Init initializes the agent store by migrating the schema
func (s *Store) Init() error {
	return s.db.AutoMigrate(&Agent{})
}

// Create creates a new agent in the database
func (s *Store) Create(ctx context.Context, agent *Agent) error {
	return s.db.WithContext(ctx).Create(agent).Error
}

// Get retrieves an agent by its ID
func (s *Store) Get(ctx context.Context, id uint) (*Agent, error) {
	var agent Agent
	err := s.db.WithContext(ctx).First(&agent, id).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (s *Store) UpdateConfig(ctx context.Context, agent *Agent) error {
	return s.db.WithContext(ctx).Model(agent).Updates(map[string]interface{}{
		"config":      agent.Config,
		"name":        agent.Name,
		"description": agent.Description,
		"is_public":   agent.IsPublic,
	}).Error
}

func (s *Store) GetByUUID(ctx context.Context, uuid string) (*Agent, error) {
	var agent Agent
	err := s.db.WithContext(ctx).Where("uuid = ?", uuid).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (s *Store) GetByUUIDAndAccountID(ctx context.Context, uuid, accountID string) (*Agent, error) {
	var agent Agent
	err := s.db.WithContext(ctx).Where("uuid = ? and account_id = ?", uuid, accountID).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// Delete deletes an agent by its ID
func (s *Store) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&Agent{}, id).Error
}

// FindBySessionID retrieves all agents for a given session ID
func (s *Store) FindBySessionID(ctx context.Context, sessionID string) ([]Agent, error) {
	var agents []Agent
	err := s.db.WithContext(ctx).Where("session_id like ?", sessionID).Find(&agents).Error
	if err != nil {
		return nil, err
	}
	return agents, nil
}

// FindByAccountID retrieves all agents for a given account ID
func (s *Store) FindByAccountID(ctx context.Context, accountID string) ([]Agent, error) {
	var (
		agents []Agent
		seen   = make(map[string]struct{})
	)
	err := s.db.WithContext(ctx).Where("account_id = ?", accountID).Find(&agents).Error
	if err != nil {
		return nil, err
	}
	for _, agent := range agents {
		seen[agent.UUID] = struct{}{}
	}
	var agentsFromSessionID []Agent
	err = s.db.WithContext(ctx).Where("id IN (SELECT agent_uuid FROM sessions WHERE account_id = ?)", accountID).Find(&agentsFromSessionID).Error
	if err != nil {
		return nil, err
	}

	for _, agent := range agentsFromSessionID {
		if _, exists := seen[agent.UUID]; !exists {
			agents = append(agents, agent)
			seen[agent.UUID] = struct{}{}
		}
	}
	return agents, nil
}

// List retrieves all agents
func (s *Store) List(ctx context.Context) ([]Agent, error) {
	var agents []Agent
	err := s.db.WithContext(ctx).Find(&agents).Error
	return agents, err
}
