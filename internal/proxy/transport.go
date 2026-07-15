package proxy

// transport.go - Optimized HTTP transport for upstream connections
//
// Responsibilities:
// - Configure http.Transport with performance-tuned settings
// - Enable connection reuse across requests (high MaxIdleConnsPerHost)
// - Set appropriate timeouts for connection establishment and idle connections
// - Enable TCP_NODELAY to minimize latency
// - Support HTTP/2 for better multiplexing
//
// Key Settings:
// - MaxIdleConnsPerHost: 100+ (allow many idle connections per upstream)
// - IdleConnTimeout: 90s (keep idle connections alive)
// - DisableCompression: false (allow gzip)
// - TLSHandshakeTimeout: 10s
// - ExpectContinueTimeout: 1s
//
// Key Functions:
// - NewTransport() *http.Transport: Create optimized transport with default settings
// - WithTimeout(timeout time.Duration) *http.Transport: Create transport with custom timeout
//
// Inputs: Timeout configurations
// Outputs: Configured http.Transport ready for ReverseProxy

import (
	"net"
	"net/http"
	"time"
)

// NewTransport creates an optimized HTTP transport for upstream connections
func NewTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second, // Fast failure for unreachable hosts
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10000,
		MaxIdleConnsPerHost:   10000,
		MaxConnsPerHost:       0, // unlimited
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second, // Time to wait for a server's response headers
		DisableKeepAlives:     false,
	}
}
