package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
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
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/v1/users/", handleUserByID)
	mux.HandleFunc("/api/v1/users", handleUsers)

	http.ListenAndServe(":"+port, mux)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, users)
}

func handleUserByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	for _, u := range users {
		if u.ID == id {
			writeJSON(w, http.StatusOK, u)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
