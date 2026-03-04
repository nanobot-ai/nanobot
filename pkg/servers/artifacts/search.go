package artifacts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type searchArtifactsParams struct {
	Query        string `json:"query,omitempty"`
	ArtifactType string `json:"artifactType,omitempty"`
}

type searchResult struct {
	Items []searchResultItem `json:"items"`
}

type searchResultItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DisplayName   string `json:"displayName,omitempty"`
	Description   string `json:"description,omitempty"`
	ArtifactType  string `json:"artifactType"`
	AuthorEmail   string `json:"authorEmail,omitempty"`
	LatestVersion int    `json:"latestVersion"`
	Visibility    string `json:"visibility"`
}

func (s *Server) searchArtifacts(ctx context.Context, params searchArtifactsParams) (*searchResult, error) {
	cfg, err := getObotConfig(ctx)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(cfg.baseURL + "/api/published-artifacts")
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	q := u.Query()
	if params.ArtifactType != "" {
		q.Set("type", params.ArtifactType)
	}
	if params.Query != "" {
		q.Set("q", params.Query)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if cfg.authHeader != "" {
		req.Header.Set("Authorization", cfg.authHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search artifacts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("search artifacts failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result searchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	return &result, nil
}
