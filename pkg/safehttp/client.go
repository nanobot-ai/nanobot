package safehttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultTimeout = 5 * time.Second

// NewClient returns an HTTP client that can block selected local address ranges.
func NewClient(blockLoopback, blockPrivateIP, blockLinkLocal bool) *http.Client {
	return NewClientWithTimeout(blockLoopback, blockPrivateIP, blockLinkLocal, defaultTimeout)
}

// NewClientWithAllowList returns an HTTP client that can block selected local address ranges,
// except for matching hosts in the allow list.
func NewClientWithAllowList(blockLoopback, blockPrivateIP, blockLinkLocal bool, allowList []string) *http.Client {
	return NewClientWithAllowListAndTimeout(blockLoopback, blockPrivateIP, blockLinkLocal, allowList, defaultTimeout)
}

// NewClientWithTimeout returns an HTTP client that can block selected local address ranges with a custom timeout.
func NewClientWithTimeout(blockLoopback, blockPrivateIP, blockLinkLocal bool, timeout time.Duration) *http.Client {
	return NewClientWithAllowListAndTimeout(blockLoopback, blockPrivateIP, blockLinkLocal, nil, timeout)
}

// NewClientWithAllowListAndTimeout returns an HTTP client that can block selected local address ranges
// with a custom timeout, except for matching hosts in the allow list.
func NewClientWithAllowListAndTimeout(blockLoopback, blockPrivateIP, blockLinkLocal bool, allowList []string, timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &safeDialer{
		dialer:         &net.Dialer{},
		resolver:       net.DefaultResolver,
		blockLoopback:  blockLoopback,
		blockPrivateIP: blockPrivateIP,
		blockLinkLocal: blockLinkLocal,
		allowList:      parseAllowList(allowList),
	}
	transport.DialContext = dialer.DialContext

	return &http.Client{
		Timeout:   timeout,
		Transport: checkingTransport{base: transport, dialer: dialer},
	}
}

type checkingTransport struct {
	base   http.RoundTripper
	dialer *safeDialer
}

func (t checkingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if host := req.URL.Hostname(); host != "" {
		if _, err := t.dialer.checkHost(req.Context(), host, portForURL(req.URL)); err != nil {
			return nil, err
		}
	}
	return t.base.RoundTrip(req)
}

type safeDialer struct {
	dialer   *net.Dialer
	resolver *net.Resolver

	blockLoopback  bool
	blockPrivateIP bool
	blockLinkLocal bool
	allowList      []allowListEntry
}

func (d *safeDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	ips, err := d.checkHost(ctx, host, port)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, ip := range ips {
		conn, err := d.dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("address %s resolved to no IP addresses", host)
}

func (d *safeDialer) checkHost(ctx context.Context, host, port string) ([]net.IP, error) {
	ips, err := d.lookup(ctx, host)
	if err != nil {
		return nil, err
	}
	if d.isAllowed(host, port) {
		return ips, nil
	}

	for _, ip := range ips {
		if reason := d.blockedReason(ip); reason != "" {
			return nil, fmt.Errorf("address %s resolves to blocked %s IP %s", host, reason, ip)
		}
	}
	return ips, nil
}

func (d *safeDialer) isAllowed(host, port string) bool {
	host = normalizeHost(host)
	for _, entry := range d.allowList {
		if entry.port != "" && entry.port != port {
			continue
		}
		if entry.suffix {
			if strings.HasSuffix(host, entry.host) && host != entry.host {
				return true
			}
			continue
		}
		if host == entry.host {
			return true
		}
	}
	return false
}

func (d *safeDialer) lookup(ctx context.Context, host string) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}

	addrs, err := d.resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("address %s resolved to no IP addresses", host)
	}

	ips := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		ips = append(ips, addr.IP)
	}
	return ips, nil
}

func (d *safeDialer) blockedReason(ip net.IP) string {
	if d.blockLoopback && ip.IsLoopback() {
		return "loopback"
	}
	if d.blockPrivateIP && ip.IsPrivate() {
		return "private"
	}
	if d.blockLinkLocal && (ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()) {
		return "link-local"
	}
	return ""
}

type allowListEntry struct {
	host   string
	port   string
	suffix bool
}

func parseAllowList(entries []string) []allowListEntry {
	result := make([]allowListEntry, 0, len(entries))
	for _, entry := range entries {
		parsed, ok := parseAllowListEntry(entry)
		if ok {
			result = append(result, parsed)
		}
	}
	return result
}

func parseAllowListEntry(entry string) (allowListEntry, bool) {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return allowListEntry{}, false
	}

	host, port := splitAllowListHostPort(entry)
	host = normalizeHost(host)
	if host == "" {
		return allowListEntry{}, false
	}

	host, suffix := strings.CutPrefix(host, "*")
	if suffix {
		if host == "" || !strings.HasPrefix(host, ".") || strings.Contains(host, "*") {
			return allowListEntry{}, false
		}
	} else if strings.Contains(host, "*") {
		return allowListEntry{}, false
	}

	return allowListEntry{
		host:   host,
		port:   port,
		suffix: suffix,
	}, true
}

func splitAllowListHostPort(entry string) (string, string) {
	if host, port, err := net.SplitHostPort(entry); err == nil {
		return host, port
	}
	if strings.Count(entry, ":") == 1 {
		host, port, ok := strings.Cut(entry, ":")
		if ok && host != "" && port != "" {
			return host, port
		}
	}
	return entry, ""
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	host = strings.TrimSuffix(host, ".")
	return host
}

func portForURL(u *url.URL) string {
	if port := u.Port(); port != "" {
		return port
	}
	switch u.Scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	default:
		return ""
	}
}
