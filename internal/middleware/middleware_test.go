package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestMiddlewareChain(t *testing.T) {
	// Create a no-op logger for testing
	logger := zap.NewNop()

	// Create a dummy handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Create a dummy handler that returns 200 OK
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	t.Run("Normal Request", func(t *testing.T) {
		chain := Chain(okHandler, Recovery(logger), RequestID(), Logging(logger, nil))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		chain.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		if rec.Body.String() != "OK" {
			t.Errorf("expected body 'OK', got '%s'", rec.Body.String())
		}

		requestID := rec.Header().Get("X-Request-ID")
		if requestID == "" {
			t.Error("expected X-Request-ID header to be set")
		}
	})

	t.Run("Panic Recovery", func(t *testing.T) {
		chain := Chain(panicHandler, Recovery(logger), RequestID(), Logging(logger, nil))

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		rec := httptest.NewRecorder()

		chain.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}

		requestID := rec.Header().Get("X-Request-ID")
		if requestID == "" {
			t.Error("expected X-Request-ID header to be set even on panic")
		}
	})
}
