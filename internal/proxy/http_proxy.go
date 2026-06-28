package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// HTTPProxy is responsible for forwarding requests to upstream servers
type HTTPProxy struct {
	transport http.RoundTripper
}

// NewHTTPProxy creates a new HTTP reverse proxy
func NewHTTPProxy(transport http.RoundTripper) *HTTPProxy {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &HTTPProxy{
		transport: transport,
	}
}

// ServeHTTP forwards the request to the upstream URL, optionally stripping a prefix
func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request, upstreamURL string, stripPrefix string) {
	target, err := url.Parse(upstreamURL)
	if err != nil {
		http.Error(w, "Bad Gateway: Invalid Upstream URL", http.StatusBadGateway)
		return
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host

			// Strip the prefix if configured
			if stripPrefix != "" && strings.HasPrefix(req.URL.Path, stripPrefix) {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, stripPrefix)
				// Ensure path still starts with a slash
				if !strings.HasPrefix(req.URL.Path, "/") {
					req.URL.Path = "/" + req.URL.Path
				}
			}

			// Add custom gateway headers
			req.Header.Set("X-Gateway-Version", "1.0.0")
			
			// Clean up Host header to match target
			req.Host = target.Host
		},
		Transport: p.transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Set("X-Gateway-Error", err.Error())
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
	}

	proxy.ServeHTTP(w, r)
}
