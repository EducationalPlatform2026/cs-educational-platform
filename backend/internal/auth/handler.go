package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"

	"cs-educational-platform/backend/internal/db"
	"cs-educational-platform/backend/internal/httputil"
)

// ── Request / Response types ────────────────────────────────────────────────

type registerRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"` // optional; defaults to RoleStudent
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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		httputil.Error(w, "email, password, first_name and last_name are required", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		httputil.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = RoleStudent
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			httputil.Error(w, "email already registered", http.StatusConflict)
			return
		}
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	token, err := GenerateToken(userID, req.Role)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, authResponse{
		Token:     token,
		UserID:    userID,
		Role:      req.Role,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		httputil.Error(w, "email and password are required", http.StatusBadRequest)
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
		// Identical message for unknown email and wrong password — prevents user enumeration.
		httputil.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !CheckPassword(req.Password, passwordHash) {
		httputil.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := GenerateToken(userID, role)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, authResponse{
		Token:     token,
		UserID:    userID,
		Role:      role,
		FirstName: firstName,
		LastName:  lastName,
	})
}
