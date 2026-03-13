package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	obotconfig "github.com/nanobot-ai/nanobot/pkg/servers/obot"
)

const (
	defaultSearchLimit = 20
	maxSearchLimit     = 50
)

type obotClient struct {
	httpClient *http.Client
	config     obotconfig.Config
}

type skillRecord struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DisplayName  string `json:"displayName,omitempty"`
	Description  string `json:"description,omitempty"`
	RepoID       string `json:"repoID,omitempty"`
	RepoURL      string `json:"repoURL,omitempty"`
	RepoRef      string `json:"repoRef,omitempty"`
	CommitSHA    string `json:"commitSHA,omitempty"`
	RelativePath string `json:"relativePath,omitempty"`
	InstallHash  string `json:"installHash,omitempty"`
	Valid        bool   `json:"valid,omitempty"`
}

type listSkillsResponse struct {
	Items []skillRecord `json:"items"`
}

func newClient(ctx context.Context) (*obotClient, error) {
	cfg, err := obotconfig.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w — nanobot skill tools require an Obot platform connection", err)
	}

	return &obotClient{
		httpClient: http.DefaultClient,
		config:     cfg,
	}, nil
}

func (c *obotClient) SearchSkills(ctx context.Context, query, repoID string, limit int) ([]skillRecord, error) {
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}

	u, err := url.Parse(c.config.BaseURL + "/api/skills")
	if err != nil {
		return nil, fmt.Errorf("failed to build skill search URL: %w", err)
	}

	params := u.Query()
	if trimmed := strings.TrimSpace(query); trimmed != "" {
		params.Set("q", trimmed)
	}
	if trimmed := strings.TrimSpace(repoID); trimmed != "" {
		params.Set("repoID", trimmed)
	}
	params.Set("limit", strconv.Itoa(limit))
	u.RawQuery = params.Encode()

	var result listSkillsResponse
	if err := c.doJSON(ctx, http.MethodGet, u.String(), nil, &result, "search skills"); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c *obotClient) GetSkill(ctx context.Context, skillID string) (*skillRecord, error) {
	var result skillRecord
	if err := c.doJSON(ctx, http.MethodGet, c.config.BaseURL+"/api/skills/"+url.PathEscape(skillID), nil, &result, "load skill detail"); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *obotClient) DownloadSkill(ctx context.Context, skillID string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.BaseURL+"/api/skills/"+url.PathEscape(skillID)+"/download", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create skill download request: %w", err)
	}
	obotconfig.ApplyAuth(req, c.config)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download skill: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, requestError(resp, "download skill")
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 100*1024*1024+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read skill download: %w", err)
	}
	if len(data) > 100*1024*1024 {
		return nil, fmt.Errorf("skill archive exceeds maximum size of %d bytes", 100*1024*1024)
	}

	return data, nil
}

func (c *obotClient) doJSON(ctx context.Context, method, requestURL string, body io.Reader, out any, action string) error {
	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request to %s: %w", action, err)
	}
	obotconfig.ApplyAuth(req, c.config)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to %s: %w", action, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return requestError(resp, action)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode %s response: %w", action, err)
	}
	return nil
}

func requestError(resp *http.Response, action string) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return fmt.Errorf("%s failed (status %d)", action, resp.StatusCode)
	}
	return fmt.Errorf("%s failed (status %d): %s", action, resp.StatusCode, trimmed)
}
