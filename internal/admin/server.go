package admin

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	port       int
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
func NewServer(port int, gateway interface{}) *Server {
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("/admin/api/stats", corsMiddleware(HandleStats))
	mux.HandleFunc("/admin/api/routes", corsMiddleware(HandleRoutes))
	mux.HandleFunc("/admin/api/upstreams", corsMiddleware(HandleUpstreams))
	mux.HandleFunc("/admin/api/circuit-breakers", corsMiddleware(HandleCircuitBreakers))
	mux.HandleFunc("/admin/api/circuit-breakers/reset", corsMiddleware(HandleCircuitBreakerReset))
	mux.HandleFunc("/admin/api/config/reload", corsMiddleware(HandleConfigReload))
	mux.HandleFunc("/admin/health", corsMiddleware(HandleHealth))

	// Serve React app (assuming built files are in admin-ui/dist)
	fs := http.FileServer(http.Dir("./admin-ui/dist"))
	mux.Handle("/", fs)

	return &Server{
		port: port,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}
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
