package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

// chromeTransport returns a round-tripper that presents a Chrome TLS fingerprint
// via utls and correctly speaks HTTP/2 (which LinkedIn requires — the browser
// HAR shows :authority/:method/:path pseudo-headers). Go's default http.Transport
// won't upgrade to HTTP/2 with a custom DialTLSContext because it doesn't
// recognise utls.UConn as *tls.Conn, so we handle the two protocols explicitly.
func chromeTransport() http.RoundTripper {
	return &utlsRoundTripper{}
}

type utlsRoundTripper struct {
	mu   sync.Mutex
	h2   *http2.Transport // lazily initialised
	h1   *http.Transport  // lazily initialised
	once sync.Once
}

func (rt *utlsRoundTripper) init() {
	rt.once.Do(func() {
		rt.h2 = &http2.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				return utlsDial(ctx, network, addr)
			},
		}
		rt.h1 = &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return utlsDial(ctx, network, addr)
			},
			ForceAttemptHTTP2: false,
		}
	})
}

func (rt *utlsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.init()
	// LinkedIn serves everything over HTTP/2; prefer h2, fall back to h1 on error.
	resp, err := rt.h2.RoundTrip(req)
	if err != nil && isH2Unsupported(err) {
		return rt.h1.RoundTrip(req)
	}
	return resp, err
}

// utlsDial creates a utls connection impersonating Chrome.
func utlsDial(ctx context.Context, network, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	tlsConn := utls.UClient(conn, &utls.Config{
		ServerName: host,
	}, utls.HelloChrome_Auto)

	if err := tlsConn.HandshakeContext(ctx); err != nil {
		conn.Close()
		return nil, err
	}

	return tlsConn, nil
}

func isH2Unsupported(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "http2: unsupported scheme") ||
		strings.Contains(msg, "http2:") ||
		strings.Contains(msg, "unexpected ALPN")
}

// isLocalhostURL returns true for localhost/127.0.0.1 URLs (test mock servers).
func isLocalhostURL(baseURL string) bool {
	return strings.Contains(baseURL, "127.0.0.1") || strings.Contains(baseURL, "localhost")
}

// tlsConfigForAddr returns a *tls.Config for the given host.
func tlsConfigForAddr(host string) *tls.Config {
	return &tls.Config{
		ServerName: host,
	}
}

// Ensure unused import doesn't cause issues.
var _ = fmt.Sprintf
