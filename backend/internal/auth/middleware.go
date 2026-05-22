package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	ContextKeyUserID contextKey = "user_id"
	ContextKeyRole   contextKey = "role"
)

// Middleware validates the Bearer JWT on every request.
// Attaches user_id and role to the request context on success.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "missing or malformed Authorization header", http.StatusUnauthorized)
			return
		}

		claims, err := ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns a middleware that only allows requests whose token
// carries one of the specified roles. Must be chained after Middleware.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(ContextKeyRole).(string)
			if _, ok := allowed[role]; !ok {
				http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// UserIDFromCtx is a helper to retrieve the authenticated user's ID from a context.
func UserIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(ContextKeyUserID).(string)
	return id
}

// RoleFromCtx is a helper to retrieve the authenticated user's role from a context.
func RoleFromCtx(ctx context.Context) string {
	role, _ := ctx.Value(ContextKeyRole).(string)
	return role
}
