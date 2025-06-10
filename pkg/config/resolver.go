package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/types"
	"sigs.k8s.io/yaml"
)

type resource struct {
	resourceType string
	url          string
	parts        []string
	ref          string
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %w", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching %s: status code %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response from %s: %w", url, err)
	}

	return data, nil
}

func (r *resource) Load(ctx context.Context) (result types.Config, _ error) {
	data, err := r.read(ctx)
	if err != nil {
		return result, fmt.Errorf("error reading resource %s: %w", r.url, err)
	}

	obj := map[string]any{}
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return result, fmt.Errorf("error unmarshalling resource %s: %w", r.url, err)
	}

	s := getSchema()
	if err := s.Validate(obj); err != nil {
		return result, fmt.Errorf("error validating resource %s: %w", r.url, err)
	}

	if err := types.Marshal(obj, &result); err != nil {
		return result, fmt.Errorf("error marshalling resource %s: %w", r.url, err)
	}
	return
}

func (r *resource) SourceRel(source mcp.ServerSource) (mcp.ServerSource, error) {
	if source.Repo != "" || source.SubPath == "" {
		return source, nil
	}

	switch r.resourceType {
	case "http":
		return mcp.ServerSource{}, fmt.Errorf("cannot resolve relative source for config loaded from HTTP: %s", r.url)
	case "path":
		cwd, err := r.Cwd()
		if err != nil {
			return mcp.ServerSource{}, fmt.Errorf("error getting current working directory for %s: %w", r, err)
		}
		subPath := source.SubPath
		for strings.HasPrefix(subPath, "../") {
			subPath = strings.TrimPrefix(subPath, "../")
			cwd = filepath.Dir(cwd)
		}

		source.Repo = cwd
		return source, nil
	case "git":
		source.Repo = fmt.Sprintf("https://%s/%s/%s.git", r.parts[0], r.parts[1], r.parts[2])
		if source.Reference == "" && source.Tag == "" && source.Branch == "" && source.Commit == "" {
			source.Reference = r.ref
		}
		source.SubPath = strings.Join(append(r.parts[3:], source.SubPath), "/")
		return source, nil
	}

	return mcp.ServerSource{}, fmt.Errorf("unknown resource type: %s", r.resourceType)
}

func (r *resource) String() string {
	if len(r.parts) > 0 {
		return strings.Join(r.parts, "/")
	}
	return r.url
}

func (r *resource) read(ctx context.Context) ([]byte, error) {
	if r.resourceType == "http" {
		return httpGet(ctx, r.url)
	}

	if r.resourceType == "path" {
		f, err := r.fileToRead()
		if err != nil {
			return nil, err
		}
		return os.ReadFile(f)
	}

	if r.resourceType == "git" {
		return gitRead(ctx, r.parts, r.ref)
	}

	return nil, fmt.Errorf("unknown resource type: %s", r.resourceType)
}

func joinPath(url, path string) string {
	parts := strings.Split(url, "/")
	if len(parts) == 1 {
	}

	dir, file := parts[:len(parts)-1], parts[len(parts)-1]
	if strings.Contains(file, ".") {
		url = strings.Join(dir, "/")
	}

	if !strings.HasSuffix(url, "/") {
		url += "/"
	}

	if strings.HasPrefix(path, "./") {
		path = strings.TrimPrefix(path, "./")
	}
	return url + path
}

func (r *resource) Cwd() (string, error) {
	if r.resourceType != "path" {
		return ".", nil
	}

	file, err := r.fileToRead()
	if err != nil {
		return "", fmt.Errorf("error getting file to read for %s: %w", r.url, err)
	}

	return filepath.Dir(file), nil
}

func (r *resource) fileToRead() (string, error) {
	path := r.url

	s, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("error looking up %s: %w", path, err)
	}

	if s.IsDir() {
		path = filepath.Join(path, "nanobot.yaml")
	}

	return path, nil
}

func (r *resource) Rel(path string) (*resource, error) {
	switch r.resourceType {
	case "http":
		return &resource{
			resourceType: "http",
			url:          joinPath(r.url, path),
		}, nil
	case "path":
		return &resource{
			resourceType: "path",
			url:          joinPath(r.url, path),
		}, nil
	case "git":
		return &resource{
			resourceType: "git",
			parts:        append(r.parts, path),
			ref:          r.ref,
		}, nil
	}
	return nil, fmt.Errorf("unknown resource type: %s", r.resourceType)
}

func gitRead(ctx context.Context, parts []string, ref string) ([]byte, error) {
	if parts[0] != "github.com" {
		// Handle other git providers or formats
		return nil, fmt.Errorf("git read not implemented for %s", strings.Join(parts, "/"))
	}

	owner, repo := parts[1], parts[2]
	path := strings.Join(parts[3:], "/")

	if ref == "" {
		ref = "HEAD"
	}

	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, ref, path)
	return httpGet(ctx, url)
}

func resolve(name string) (*resource, error) {
	if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") {
		// Handle HTTP resources
		return &resource{
			resourceType: "http",
			url:          name,
		}, nil
	}

	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, ".") {
		// Handle local file paths
		return &resource{
			resourceType: "path",
			url:          name,
		}, nil
	}

	location, ref, _ := strings.Cut(name, "@")
	parts := strings.Split(location, "/")

	if !strings.Contains(parts[0], ".") {
		parts = append([]string{"github.com"}, parts...)
	}

	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid git resource format, must be {owner}/{repo} format: %s", name)
	}

	return &resource{
		resourceType: "git",
		parts:        parts,
		ref:          ref,
	}, nil
}
