package artifacts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type setArtifactSubjectsParams struct {
	ID       string            `json:"id"`
	Version  *int              `json:"version,omitempty"`
	Subjects []artifactSubject `json:"subjects,omitempty"`
}

type artifactSubject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type setArtifactSubjectsResult struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Subjects []artifactSubject `json:"subjects,omitempty"`
	Message  string            `json:"message"`
}

func (s *Server) setArtifactSubjects(ctx context.Context, params setArtifactSubjectsParams) (*setArtifactSubjectsResult, error) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, fmt.Errorf("id is required")
	}

	subjects, err := normalizeArtifactSubjects(params.Subjects)
	if err != nil {
		return nil, err
	}

	cfg, err := getObotConfig(ctx)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]any{
		"version":  params.Version,
		"subjects": subjects,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, cfg.baseURL+"/api/published-artifacts/"+params.ID, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.authHeader != "" {
		req.Header.Set("Authorization", cfg.authHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update artifact subjects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("update artifact subjects failed (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		ID       string            `json:"id"`
		Name     string            `json:"name"`
		Subjects []artifactSubject `json:"subjects,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &setArtifactSubjectsResult{
		ID:       apiResp.ID,
		Name:     apiResp.Name,
		Subjects: apiResp.Subjects,
		Message:  artifactSubjectsMessage(apiResp.Name, params.Version, apiResp.Subjects),
	}, nil
}

func normalizeArtifactSubjects(subjects []artifactSubject) ([]artifactSubject, error) {
	if len(subjects) == 0 {
		return nil, nil
	}

	result := make([]artifactSubject, 0, len(subjects))
	seen := make(map[artifactSubject]struct{}, len(subjects))

	for _, subject := range subjects {
		normalized := artifactSubject{
			Type: strings.ToLower(strings.TrimSpace(subject.Type)),
			ID:   strings.TrimSpace(subject.ID),
		}

		switch normalized.Type {
		case "user", "group":
			if normalized.ID == "" {
				return nil, fmt.Errorf("%s subject id is required", normalized.Type)
			}
		case "selector":
			if normalized.ID != "*" {
				return nil, fmt.Errorf("selector subject id must be '*'")
			}
		default:
			return nil, fmt.Errorf("invalid subject type: %q", subject.Type)
		}

		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	if len(result) > 1 {
		for _, subject := range result {
			if subject.Type == "selector" && subject.ID == "*" {
				return nil, fmt.Errorf("selector '*' must be the only subject")
			}
		}
	}

	return result, nil
}

func artifactSubjectsMessage(name string, version *int, subjects []artifactSubject) string {
	target := name
	if version != nil {
		target = fmt.Sprintf("%s v%d", name, *version)
	}
	switch {
	case len(subjects) == 0:
		return fmt.Sprintf("Updated sharing for %s. It is now owner-only.", target)
	case len(subjects) == 1 && subjects[0].Type == "selector" && subjects[0].ID == "*":
		return fmt.Sprintf("Updated sharing for %s. It is now shared with all Obot users.", target)
	default:
		return fmt.Sprintf("Updated sharing for %s with %d subject(s).", target, len(subjects))
	}
}
