package artifacts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type listSubjectsParams struct {
	Type  string `json:"type"`
	Query string `json:"query,omitempty"`
}

type listSubjectsResult struct {
	Type  string            `json:"type"`
	Items []listSubjectItem `json:"items"`
}

type listSubjectItem struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Username    string `json:"username,omitempty"`
	Email       string `json:"email,omitempty"`
}

func (s *Server) listSubjects(ctx context.Context, params listSubjectsParams) (*listSubjectsResult, error) {
	cfg, err := getObotConfig(ctx)
	if err != nil {
		return nil, err
	}

	subjectType := strings.ToLower(strings.TrimSpace(params.Type))
	query := strings.TrimSpace(params.Query)

	switch subjectType {
	case "user":
		items, err := listUserSubjects(ctx, cfg, query)
		if err != nil {
			return nil, err
		}
		return &listSubjectsResult{Type: subjectType, Items: items}, nil
	case "group":
		items, err := listGroupSubjects(ctx, cfg, query)
		if err != nil {
			return nil, err
		}
		return &listSubjectsResult{Type: subjectType, Items: items}, nil
	default:
		return nil, fmt.Errorf("type must be 'user' or 'group'")
	}
}

func listUserSubjects(ctx context.Context, cfg obotConfig, query string) ([]listSubjectItem, error) {
	u, err := url.Parse(cfg.baseURL + "/api/users")
	if err != nil {
		return nil, fmt.Errorf("failed to build users URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if cfg.authHeader != "" {
		req.Header.Set("Authorization", cfg.authHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("list users failed (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Items []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName,omitempty"`
			Username    string `json:"username,omitempty"`
			Email       string `json:"email,omitempty"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse users response: %w", err)
	}

	items := make([]listSubjectItem, 0, len(apiResp.Items))
	for _, item := range apiResp.Items {
		subject := listSubjectItem{
			ID:          item.ID,
			Type:        "user",
			DisplayName: item.DisplayName,
			Username:    item.Username,
			Email:       item.Email,
		}
		if subjectMatchesQuery(subject, query) {
			items = append(items, subject)
		}
	}

	return items, nil
}

func listGroupSubjects(ctx context.Context, cfg obotConfig, query string) ([]listSubjectItem, error) {
	u, err := url.Parse(cfg.baseURL + "/api/groups")
	if err != nil {
		return nil, fmt.Errorf("failed to build groups URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if cfg.authHeader != "" {
		req.Header.Set("Authorization", cfg.authHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("list groups failed (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp []struct {
		ID   string `json:"id"`
		Name string `json:"name,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse groups response: %w", err)
	}

	items := make([]listSubjectItem, 0, len(apiResp))
	for _, item := range apiResp {
		subject := listSubjectItem{
			ID:   item.ID,
			Type: "group",
			Name: item.Name,
		}
		if subjectMatchesQuery(subject, query) {
			items = append(items, subject)
		}
	}

	return items, nil
}

func subjectMatchesQuery(item listSubjectItem, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}

	fields := []string{
		item.ID,
		item.Name,
		item.DisplayName,
		item.Username,
		item.Email,
	}

	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}

	return false
}
