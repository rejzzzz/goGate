package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
	"strconv"
)

type order struct {
	ID   string `json:"id"`
	Item string `json:"item"`
	Qty  int    `json:"qty"`
}

var orders = []order{
	{ID: "1", Item: "Widget", Qty: 5},
	{ID: "2", Item: "Gadget", Qty: 2},
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	delayMs, _ := strconv.Atoi(os.Getenv("DELAY_MS"))

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/orders/", makeOrderByIDHandler(delayMs))
	mux.HandleFunc("/orders", makeOrdersHandler(delayMs))

	http.ListenAndServe(":"+port, mux)
}

func makeOrdersHandler(delayMs int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
		writeJSON(w, http.StatusOK, orders)
	}
}

func makeOrderByIDHandler(delayMs int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
		id := strings.TrimPrefix(r.URL.Path, "/orders/")
		for _, o := range orders {
			if o.ID == id {
				writeJSON(w, http.StatusOK, o)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
