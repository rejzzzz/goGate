package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type user struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var users = []user{
	{ID: "1", Name: "Alice"},
	{ID: "2", Name: "Bob"},
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	delayMs, _ := strconv.Atoi(os.Getenv("DELAY_MS"))
	if delayMs == 0 {
		delayMs = 50
	}

	errorRate, _ := strconv.ParseFloat(os.Getenv("ERROR_RATE"), 64)
	if errorRate == 0 {
		errorRate = 0.05
	}

	usersHandler, userByIDHandler := makeHandlers(delayMs, errorRate)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/users/", userByIDHandler)
	mux.HandleFunc("/users", usersHandler)

	http.ListenAndServe(":"+port, mux)
}

// makeHandlers returns /users and /users/:id handlers that inject delay and
// random errors based on the supplied configuration.
func makeHandlers(delayMs int, errorRate float64) (http.HandlerFunc, http.HandlerFunc) {
	listHandler := func(w http.ResponseWriter, r *http.Request) {
		addDelay(delayMs)
		if shouldReturnError(errorRate) {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		writeJSON(w, http.StatusOK, users)
	}

	byIDHandler := func(w http.ResponseWriter, r *http.Request) {
		addDelay(delayMs)
		if shouldReturnError(errorRate) {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/users/")
		for _, u := range users {
			if u.ID == id {
				writeJSON(w, http.StatusOK, u)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	}

	return listHandler, byIDHandler
}

// handleHealth responds immediately with no delay and no random errors so the
// health checker never incorrectly marks this upstream unhealthy.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func addDelay(delayMs int) {
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
}

func shouldReturnError(errorRate float64) bool {
	return rand.Float64() < errorRate
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
