package obotmcp

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

// LookupObotConnectedServerIDs resolves mcp_server_id and catalog_entry_id from obot_list_connected_mcp_servers.
// It first tries to match any normalized connect URL from connectURLs to Obot's connect_url; if none match,
// it matches by display name or alias against displayNames (e.g. config key "gmail", name "Gmail", or shortName).
// Obot often exposes a different connect_url (proxy) than Nanobot's configured MCP url, so the name pass is required.
func LookupObotConnectedServerIDs(ctx context.Context, connectURLs []string, displayNames []string) (mcpServerID, catalogEntryID string) {
	servers, err := (obotConnectedServerLister{}).ConnectedMCPServers(ctx)
	if err != nil || len(servers) == 0 {
		return "", ""
	}

	normTargets := normalizeURLList(connectURLs)
	names := dedupeNonEmpty(displayNames)

	// Pass 1: exact normalized URL match
	for _, cs := range servers {
		cu, err := normalizeConnectURL(cs.ConnectURL)
		if err != nil || cu == "" {
			continue
		}
		for _, t := range normTargets {
			if t != "" && cu == t {
				return mcpServerIDFromConnected(cs), catalogEntryIDFromConnected(cs)
			}
		}
	}

	// Pass 1b: same host (and port) as connect URL — scheme/path may differ between Obot and the client.
	candidateHosts := connectURLHostList(connectURLs)
	for _, cs := range servers {
		obHost, ok := connectURLHostKey(cs.ConnectURL)
		if !ok || obHost == "" {
			continue
		}
		for _, h := range candidateHosts {
			if h != "" && h == obHost {
				return mcpServerIDFromConnected(cs), catalogEntryIDFromConnected(cs)
			}
		}
	}

	// Pass 2: name / alias (and config key) match — Obot connect_url often differs from client BaseURL
	for _, cs := range servers {
		if matchDisplayNames(cs, names) {
			return mcpServerIDFromConnected(cs), catalogEntryIDFromConnected(cs)
		}
	}

	return "", ""
}

func normalizeURLList(urls []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, raw := range urls {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		u, err := normalizeConnectURL(raw)
		if err != nil || u == "" {
			continue
		}
		if _, dup := seen[u]; dup {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

func dedupeNonEmpty(names []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		if _, dup := seen[n]; dup {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

func matchDisplayNames(cs ConnectedServer, displayNames []string) bool {
	for _, dn := range displayNames {
		if strings.EqualFold(strings.TrimSpace(cs.Name), dn) {
			return true
		}
		if a := strings.TrimSpace(cs.Alias); a != "" && strings.EqualFold(a, dn) {
			return true
		}
	}
	return false
}

// mcpServerIDFromConnected prefers explicit mcp_server_id; otherwise uses Obot's top-level id field.
func mcpServerIDFromConnected(cs ConnectedServer) string {
	s := idToString(cs.MCPServerID)
	if s != "" {
		return s
	}
	return strings.TrimSpace(cs.ID)
}

func catalogEntryIDFromConnected(cs ConnectedServer) string {
	return idToString(cs.CatalogEntryID)
}

// LookupConnectedServerIDs matches a single connect URL and display name (legacy helper).
func LookupConnectedServerIDs(ctx context.Context, connectURL, displayName string) (mcpServerID, catalogEntryID string) {
	connectURL = strings.TrimSpace(connectURL)
	var urls []string
	if connectURL != "" {
		urls = []string{connectURL}
	}
	dn := strings.TrimSpace(displayName)
	var names []string
	if dn != "" {
		names = []string{dn}
	}
	return LookupObotConnectedServerIDs(ctx, urls, names)
}

func idToString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case float64:
		return strconv.FormatInt(int64(x), 10)
	default:
		return strings.TrimSpace(fmt.Sprint(x))
	}
}

func normalizeConnectURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("empty MCP base URL")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", err
	}
	u.Fragment = ""
	u.RawQuery = ""
	return strings.TrimSuffix(strings.ToLower(u.String()), "/"), nil
}

// connectURLHostKey returns host[:port] lowercased for comparing MCP endpoints when full URLs differ.
func connectURLHostKey(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", false
	}
	return strings.ToLower(u.Host), true
}

func connectURLHostList(urls []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, raw := range urls {
		h, ok := connectURLHostKey(raw)
		if !ok || h == "" {
			continue
		}
		if _, dup := seen[h]; dup {
			continue
		}
		seen[h] = struct{}{}
		out = append(out, h)
	}
	return out
}

// LookupConnectedServerIDsForServer resolves IDs using mcp.Server BaseURL and display name
// (name, shortName, then empty).
func LookupConnectedServerIDsForServer(ctx context.Context, srv mcp.Server) (mcpServerID, catalogEntryID string) {
	name := strings.TrimSpace(complete.First(srv.Name, srv.ShortName, ""))
	return LookupConnectedServerIDs(ctx, srv.BaseURL, name)
}

// LookupConnectedServerIDsFromCandidates tries each connect URL with a single display name (legacy).
func LookupConnectedServerIDsFromCandidates(ctx context.Context, connectURLs []string, displayName string) (mcpServerID, catalogEntryID string) {
	displayName = strings.TrimSpace(displayName)
	var names []string
	if displayName != "" {
		names = []string{displayName}
	}
	return LookupObotConnectedServerIDs(ctx, connectURLs, names)
}
