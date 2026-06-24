package safehttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

const defaultTimeout = 5 * time.Second

// NewClient returns an HTTP client that can block selected local address ranges.
func NewClient(blockLoopback, blockPrivateIP, blockLinkLocal bool) *http.Client {
	return NewClientWithTimeout(blockLoopback, blockPrivateIP, blockLinkLocal, defaultTimeout)
}

// NewClientWithTimeout returns an HTTP client that can block selected local address ranges with a custom timeout.
func NewClientWithTimeout(blockLoopback, blockPrivateIP, blockLinkLocal bool, timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &safeDialer{
		dialer:         &net.Dialer{},
		resolver:       net.DefaultResolver,
		blockLoopback:  blockLoopback,
		blockPrivateIP: blockPrivateIP,
		blockLinkLocal: blockLinkLocal,
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
		if _, err := t.dialer.checkHost(req.Context(), host); err != nil {
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
}

func (d *safeDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	ips, err := d.checkHost(ctx, host)
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

func (d *safeDialer) checkHost(ctx context.Context, host string) ([]net.IP, error) {
	ips, err := d.lookup(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		if reason := d.blockedReason(ip); reason != "" {
			return nil, fmt.Errorf("address %s resolves to blocked %s IP %s", host, reason, ip)
		}
	}
	return ips, nil
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
