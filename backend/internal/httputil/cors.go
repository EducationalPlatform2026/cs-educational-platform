package httputil

import (
	"net/http"
	"os"
	"strings"
)

// CORS wraps a handler with an origin allowlist. Allowed origins are read from
// the CORS_ORIGIN environment variable (comma-separated). Defaults to
// http://localhost:5173 for local development. Only matching origins receive
// CORS headers — unknown origins get no Access-Control-* headers and the
// browser blocks the response, preventing reflected-origin + credentials attacks.
func CORS(next http.Handler) http.Handler {
	raw := os.Getenv("CORS_ORIGIN")
	if raw == "" {
		raw = "http://localhost:5173"
	}
	allowed := make(map[string]struct{})
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowed[o] = struct{}{}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
