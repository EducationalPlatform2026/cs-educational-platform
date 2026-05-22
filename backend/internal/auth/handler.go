package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"cs-educational-platform/backend/internal/db"
)

// ── Request / Response types ────────────────────────────────────────────────

type registerRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"` // optional; defaults to "student"
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token     string `json:"token"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// ── Handlers ────────────────────────────────────────────────────────────────

// RegisterHandler godoc
// POST /auth/register
// Body: { email, password, first_name, last_name, role? }
// Returns 201 + JWT on success.
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		httpError(w, "email, password, first_name and last_name are required", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		httpError(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = "student"
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		httpError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var userID string
	err = db.Pool.QueryRow(r.Context(),
		`INSERT INTO users (email, password_hash, first_name, last_name, role)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		req.Email, hash, req.FirstName, req.LastName, req.Role,
	).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			httpError(w, "email already registered", http.StatusConflict)
			return
		}
		httpError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	token, err := GenerateToken(userID, req.Role)
	if err != nil {
		httpError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		Token:     token,
		UserID:    userID,
		Role:      req.Role,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
}

// LoginHandler godoc
// POST /auth/login
// Body: { email, password }
// Returns 200 + JWT on success.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		httpError(w, "email and password are required", http.StatusBadRequest)
		return
	}

	var userID, passwordHash, role, firstName, lastName string
	err := db.Pool.QueryRow(r.Context(),
		`SELECT id, password_hash, role, first_name, last_name
		 FROM users
		 WHERE email = $1 AND is_active = true`,
		req.Email,
	).Scan(&userID, &passwordHash, &role, &firstName, &lastName)
	if err != nil {
		// Return same message for unknown email or wrong password (prevents user enumeration)
		httpError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !CheckPassword(req.Password, passwordHash) {
		httpError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := GenerateToken(userID, role)
	if err != nil {
		httpError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		Token:     token,
		UserID:    userID,
		Role:      role,
		FirstName: firstName,
		LastName:  lastName,
	})
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func httpError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}
