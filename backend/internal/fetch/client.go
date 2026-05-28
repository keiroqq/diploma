package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultTimeout          = 15 * time.Second
	DefaultMaxResponseBytes = 10 * 1024 * 1024
)

var (
	ErrUnsafeURL        = errors.New("unsafe url")
	ErrResponseTooLarge = errors.New("response too large")
)

type maxBytesReadCloser struct {
	body      io.ReadCloser
	remaining int64
}

func (r *maxBytesReadCloser) Read(p []byte) (int, error) {
	if r.remaining <= 0 {
		var probe [1]byte
		n, err := r.body.Read(probe[:])
		if n > 0 {
			return 0, ErrResponseTooLarge
		}
		return 0, err
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}

	n, err := r.body.Read(p)
	r.remaining -= int64(n)
	return n, err
}

func (r *maxBytesReadCloser) Close() error {
	return r.body.Close()
}

type safeTransport struct {
	base     http.RoundTripper
	maxBytes int64
}

func NewSafeHTTPClient(timeout time.Duration, maxBytes int64) *http.Client {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxResponseBytes
	}

	return &http.Client{
		Timeout: timeout,
		Transport: &safeTransport{
			base:     newBaseTransport(),
			maxBytes: maxBytes,
		},
	}
}

func (t *safeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := ValidateURL(req.URL.String()); err != nil {
		return nil, err
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if t.maxBytes > 0 && resp.ContentLength > t.maxBytes {
		_ = resp.Body.Close()
		return nil, ErrResponseTooLarge
	}

	resp.Body = &maxBytesReadCloser{
		body:      resp.Body,
		remaining: t.maxBytes,
	}
	return resp, nil
}

func newBaseTransport() *http.Transport {
	dialer := &safeDialer{
		dialer: net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}

	return &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          32,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func ValidateURL(value string) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnsafeURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%w: url must use http or https", ErrUnsafeURL)
	}
	if parsed.Host == "" {
		return fmt.Errorf("%w: url host is required", ErrUnsafeURL)
	}
	host := strings.TrimSuffix(parsed.Hostname(), ".")
	if host == "" {
		return fmt.Errorf("%w: url host is required", ErrUnsafeURL)
	}
	if strings.EqualFold(host, "localhost") {
		return fmt.Errorf("%w: localhost is not allowed", ErrUnsafeURL)
	}
	if addr, err := netip.ParseAddr(host); err == nil && !isSafeAddr(addr) {
		return fmt.Errorf("%w: private addresses are not allowed", ErrUnsafeURL)
	}
	return nil
}

type safeDialer struct {
	dialer net.Dialer
}

func (d *safeDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	host = strings.TrimSuffix(host, ".")

	addrs, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("%w: host has no addresses", ErrUnsafeURL)
	}

	for _, addr := range addrs {
		if !isSafeAddr(addr) {
			return nil, fmt.Errorf("%w: private addresses are not allowed", ErrUnsafeURL)
		}
	}

	return d.dialer.DialContext(ctx, network, net.JoinHostPort(addrs[0].String(), port))
}

func isSafeAddr(addr netip.Addr) bool {
	return addr.IsValid() &&
		!addr.IsLoopback() &&
		!addr.IsPrivate() &&
		!addr.IsLinkLocalUnicast() &&
		!addr.IsLinkLocalMulticast() &&
		!addr.IsMulticast() &&
		!addr.IsUnspecified()
}
