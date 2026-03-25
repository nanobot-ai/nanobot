package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nanobot-ai/nanobot/pkg/gormdsn"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func NewStoreFromDSN(dsn string) (store *Store, err error) {
	db, err := gormdsn.NewDBFromDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	if err := tx.AutoMigrate(&Session{}, &Token{}, &WorkflowRun{}, &ScheduledTask{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	if err := migrateSessionWorkflowURIs(tx); err != nil {
		return nil, fmt.Errorf("failed to migrate session workflow URIs: %w", err)
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

// GetScheduledTask returns a scheduled task by its task URI.
func (s *Store) GetScheduledTask(ctx context.Context, taskURI string) (*ScheduledTask, error) {
	var task ScheduledTask
	err := s.db.WithContext(ctx).Where("task_uri = ?", taskURI).First(&task).Error
	return &task, err
}

// ListScheduledTasks returns all scheduled tasks ordered newest-first.
func (s *Store) ListScheduledTasks(ctx context.Context) ([]ScheduledTask, error) {
	var tasks []ScheduledTask
	err := s.db.WithContext(ctx).
		Order("created_at desc").
		Find(&tasks).Error
	return tasks, err
}

// CreateScheduledTask inserts a new scheduled task record.
func (s *Store) CreateScheduledTask(ctx context.Context, task *ScheduledTask) error {
	return s.db.WithContext(ctx).Create(task).Error
}

// UpdateScheduledTask persists definition changes to an existing scheduled task.
// Only writes configuration columns — never touches LastRunAt.
func (s *Store) UpdateScheduledTask(ctx context.Context, task *ScheduledTask) error {
	return s.db.WithContext(ctx).
		Model(task).
		Select("Name", "Prompt", "Schedule", "Timezone", "Enabled", "ExpiresAt", "NextRunAt").
		Updates(task).Error
}

// RecordScheduledTaskRun records when a scheduled task last ran and when it will next run.
func (s *Store) RecordScheduledTaskRun(ctx context.Context, taskURI string, lastRunAt time.Time, nextRunAt *time.Time) error {
	return s.db.WithContext(ctx).
		Model(&ScheduledTask{}).
		Where("task_uri = ?", taskURI).
		Updates(map[string]any{
			"last_run_at": lastRunAt,
			"next_run_at": nextRunAt,
		}).Error
}

// DeleteScheduledTask deletes a scheduled task by its task URI.
func (s *Store) DeleteScheduledTask(ctx context.Context, taskURI string) error {
	return s.db.WithContext(ctx).
		Where("task_uri = ?", taskURI).
		Delete(&ScheduledTask{}).Error
}

// NextScheduledTaskURI builds the next stable task URI for the given name.
func (s *Store) NextScheduledTaskURI(ctx context.Context, name string) (string, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	var (
		base     strings.Builder
		lastDash bool
	)
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			base.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && base.Len() > 0 {
			base.WriteByte('-')
			lastDash = true
		}
	}

	slug := strings.Trim(base.String(), "-")
	if slug == "" {
		slug = "task"
	}

	prefix := "task:///"
	var ids []string
	err := s.db.WithContext(ctx).
		Model(&ScheduledTask{}).
		Where("task_uri = ? OR task_uri LIKE ?", prefix+slug, prefix+slug+"-%").
		Pluck("task_uri", &ids).Error
	if err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return prefix + slug, nil
	}

	maxSuffix := 1
	for _, id := range ids {
		if suffix, ok := strings.CutPrefix(id, prefix+slug+"-"); ok {
			if n, err := strconv.Atoi(suffix); err == nil && n > maxSuffix {
				maxSuffix = n
			}
		}
	}

	return fmt.Sprintf("%s%s-%d", prefix, slug, maxSuffix+1), nil
}

// ListWorkflowURIs returns the URIs of the workflows that have been run in each of the given sessions.
func (s *Store) ListWorkflowURIs(ctx context.Context, sessionIDs ...string) (map[string][]string, error) {
	if len(sessionIDs) == 0 {
		return nil, nil
	}

	var runs []WorkflowRun
	err := s.db.WithContext(ctx).
		Where("session_id IN ?", sessionIDs).
		Order("session_id ASC, workflow_uri ASC").
		Find(&runs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string, len(sessionIDs))
	for _, run := range runs {
		result[run.SessionID] = append(result[run.SessionID], run.WorkflowURI)
	}

	return result, nil
}

// AddWorkflowRun records a workflow run for a session and ignores duplicates.
func (s *Store) AddWorkflowRun(ctx context.Context, sessionID, workflowURI string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	if workflowURI == "" {
		return fmt.Errorf("workflow URI cannot be empty")
	}

	run := WorkflowRun{
		SessionID:   sessionID,
		WorkflowURI: workflowURI,
	}

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&run).Error
}

func (s *Store) List(ctx context.Context) ([]Session, error) {
	var sessions []Session
	err := s.db.WithContext(ctx).Order("updated_at desc").Find(&sessions).Error
	return sessions, err
}

func (s *Store) GetTokenConfig(ctx context.Context, url string) (*oauth2.Config, *oauth2.Token, error) {
	var (
		accountID    string
		token        Token
		oauth2Config oauth2.Config
		oauth2Token  oauth2.Token
	)
	session := mcp.SessionFromContext(ctx)
	if !session.Get(types.AccountIDSessionKey, &accountID) {
		return nil, nil, nil
	}
	err := s.db.WithContext(ctx).Where("account_id = ? and url = ?", accountID, url).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal([]byte(token.Data), &struct {
		Config *oauth2.Config `json:"config,omitempty"`
		Token  *oauth2.Token  `json:"token,omitempty"`
	}{
		Config: &oauth2Config,
		Token:  &oauth2Token,
	})
	return &oauth2Config, &oauth2Token, err
}

func (s *Store) SetTokenConfig(ctx context.Context, url string, oauth2Config *oauth2.Config, oauth2token *oauth2.Token) error {
	var (
		accountID string
		token     Token
	)
	session := mcp.SessionFromContext(ctx)
	if !session.Get(types.AccountIDSessionKey, &accountID) {
		return fmt.Errorf("account ID not found in session")
	}

	err := s.db.WithContext(ctx).Where("account_id = ? and url = ?", accountID, url).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		token = Token{
			AccountID: accountID,
			URL:       url,
		}
	} else if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	tokenData, err := json.Marshal(map[string]any{
		"config": oauth2Config,
		"token":  oauth2token,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	token.Data = string(tokenData)
	if token.ID == 0 {
		return s.db.WithContext(ctx).Create(&token).Error
	}
	return s.db.WithContext(ctx).Save(&token).Error
}

func (s *Store) DeleteTokenConfig(ctx context.Context, url string) error {
	var accountID string
	session := mcp.SessionFromContext(ctx)
	if !session.Get(types.AccountIDSessionKey, &accountID) {
		return nil
	}

	return s.db.WithContext(ctx).Where("account_id = ? AND url = ?", accountID, url).Delete(&Token{}).Error
}
