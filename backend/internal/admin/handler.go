package admin

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"cs-educational-platform/backend/internal/auth"
	"cs-educational-platform/backend/internal/db"
	"cs-educational-platform/backend/internal/httputil"
)

// ── Models ───────────────────────────────────────────────────────────────────

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type CourseOverview struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	CreatedBy    string `json:"created_by"`
	CreatorName  string `json:"creator_name"`
	IsPublished  bool   `json:"is_published"`
	EnrollCount  int    `json:"enroll_count"`
	ExerciseCount int   `json:"exercise_count"`
	CreatedAt    string `json:"created_at"`
}

type Stats struct {
	TotalUsers       int `json:"total_users"`
	TotalCourses     int `json:"total_courses"`
	TotalExercises   int `json:"total_exercises"`
	TotalSubmissions int `json:"total_submissions"`
	ActiveUsers      int `json:"active_users"`
	PublishedCourses int `json:"published_courses"`
}

// ── Handlers ─────────────────────────────────────────────────────────────────

// StatsHandler — GET /admin/stats
func StatsHandler(w http.ResponseWriter, r *http.Request) {
	var s Stats
	err := db.Pool.QueryRow(r.Context(), `
		SELECT
			(SELECT COUNT(*) FROM users)                          AS total_users,
			(SELECT COUNT(*) FROM courses)                        AS total_courses,
			(SELECT COUNT(*) FROM exercises)                      AS total_exercises,
			(SELECT COUNT(*) FROM submissions)                    AS total_submissions,
			(SELECT COUNT(*) FROM users    WHERE is_active = true) AS active_users,
			(SELECT COUNT(*) FROM courses  WHERE is_published = true) AS published_courses
	`).Scan(&s.TotalUsers, &s.TotalCourses, &s.TotalExercises,
		&s.TotalSubmissions, &s.ActiveUsers, &s.PublishedCourses)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, s)
}

// ListUsersHandler — GET /admin/users
func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(r.Context(), `
		SELECT id, email, first_name, last_name, role::text, is_active, created_at::text
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName,
			&u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}
	httputil.WriteJSON(w, http.StatusOK, users)
}

// UpdateRoleHandler — PATCH /admin/users/{id}/role
func UpdateRoleHandler(w http.ResponseWriter, r *http.Request) {
	targetID := r.PathValue("id")
	callerID := auth.UserIDFromCtx(r.Context())
	if targetID == callerID {
		httputil.Error(w, "cannot change your own role", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	allowed := map[string]bool{
		auth.RoleStudent: true, auth.RoleTeachingAssistant: true,
		auth.RoleProfessor: true, auth.RoleAdmin: true,
	}
	if !allowed[body.Role] {
		httputil.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	tag, err := db.Pool.Exec(r.Context(),
		`UPDATE users SET role = $1 WHERE id = $2`, body.Role, targetID)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		httputil.Error(w, "user not found", http.StatusNotFound)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"role": body.Role})
}

// ToggleActiveHandler — PATCH /admin/users/{id}/active
func ToggleActiveHandler(w http.ResponseWriter, r *http.Request) {
	targetID := r.PathValue("id")
	callerID := auth.UserIDFromCtx(r.Context())
	if targetID == callerID {
		httputil.Error(w, "cannot deactivate your own account", http.StatusBadRequest)
		return
	}

	var isActive bool
	err := db.Pool.QueryRow(r.Context(),
		`UPDATE users SET is_active = NOT is_active WHERE id = $1 RETURNING is_active`,
		targetID,
	).Scan(&isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Error(w, "user not found", http.StatusNotFound)
			return
		}
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"is_active": isActive})
}

// ListCoursesHandler — GET /admin/courses
func ListCoursesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(r.Context(), `
		SELECT
			c.id, c.title, c.created_by::text,
			u.first_name || ' ' || u.last_name AS creator_name,
			c.is_published,
			(SELECT COUNT(*) FROM course_enrollments ce WHERE ce.course_id = c.id)::int AS enroll_count,
			(SELECT COUNT(*) FROM exercises e WHERE e.course_id = c.id)::int            AS exercise_count,
			c.created_at::text
		FROM courses c
		JOIN users u ON u.id = c.created_by
		ORDER BY c.created_at DESC
	`)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := []CourseOverview{}
	for rows.Next() {
		var co CourseOverview
		if err := rows.Scan(&co.ID, &co.Title, &co.CreatedBy, &co.CreatorName,
			&co.IsPublished, &co.EnrollCount, &co.ExerciseCount, &co.CreatedAt); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		list = append(list, co)
	}
	httputil.WriteJSON(w, http.StatusOK, list)
}
