package httputil

import (
	"encoding/json"
	"net/http"
)

// WriteJSON marshals v to JSON and writes it with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b) //nolint:errcheck
}

// Error writes a JSON {"error": msg} response with the given status code.
func Error(w http.ResponseWriter, msg string, code int) {
	WriteJSON(w, code, map[string]string{"error": msg})
}
