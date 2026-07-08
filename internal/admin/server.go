package admin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rejzzzz/goGate/internal/healthcheck"
	"github.com/rejzzzz/goGate/internal/loadbalancer"
	"github.com/rejzzzz/goGate/internal/router"
)

type Server struct {
	httpServer *http.Server
	port       int

	router      *router.Router
	upstreamMap map[string][]*loadbalancer.Upstream
	registry    *healthcheck.Registry
	reloadChan  chan<- struct{}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// NewServer creates a new admin server
func NewServer(port int, r *router.Router, uMap map[string][]*loadbalancer.Upstream, reg *healthcheck.Registry, reloadChan chan<- struct{}) *Server {
	s := &Server{
		port:        port,
		router:      r,
		upstreamMap: uMap,
		registry:    reg,
		reloadChan:  reloadChan,
	}

	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("/admin/api/stats", corsMiddleware(s.HandleStats))
	mux.HandleFunc("/admin/api/routes", corsMiddleware(s.HandleRoutes))
	mux.HandleFunc("/admin/api/upstreams", corsMiddleware(s.HandleUpstreams))
	mux.HandleFunc("/admin/api/circuit-breakers", corsMiddleware(s.HandleCircuitBreakers))
	mux.HandleFunc("/admin/api/circuit-breakers/reset", corsMiddleware(s.HandleCircuitBreakerReset))
	mux.HandleFunc("/admin/api/config/reload", corsMiddleware(s.HandleConfigReload))
	mux.HandleFunc("/admin/health", corsMiddleware(s.HandleHealth))

	// Serve React app (assuming built files are in admin-ui/dist)
	fs := http.FileServer(http.Dir("./admin-ui/dist"))
	mux.Handle("/", fs)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return s
}

// Start begins serving the admin API and UI
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the admin server
func (s *Server) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
