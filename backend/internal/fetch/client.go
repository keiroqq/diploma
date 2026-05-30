package fetch

import (
	"context"
	"crypto/tls"
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
	DefaultTimeout          = 25 * time.Second
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
	base     *http.Transport
	fallback *http.Transport
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
			base:     newBaseTransport(true),
			fallback: newBaseTransport(false),
			maxBytes: maxBytes,
		},
	}
}

func (t *safeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := ValidateURL(req.URL.String()); err != nil {
		return nil, err
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil && shouldRetryWithHTTP1(req, err) && t.fallback != nil {
		resp, err = t.fallback.RoundTrip(req.Clone(req.Context()))
	}
	if err != nil {
		return nil, err
	}
	return t.wrapResponse(resp)
}

func (t *safeTransport) wrapResponse(resp *http.Response) (*http.Response, error) {
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

func shouldRetryWithHTTP1(req *http.Request, err error) bool {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "http2:") ||
		strings.Contains(message, "http/2") ||
		strings.Contains(message, "timeout awaiting response headers")
}

func newBaseTransport(forceHTTP2 bool) *http.Transport {
	dialer := &safeDialer{
		dialer: net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     forceHTTP2,
		MaxIdleConns:          32,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if !forceHTTP2 {
		transport.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	}
	return transport
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
