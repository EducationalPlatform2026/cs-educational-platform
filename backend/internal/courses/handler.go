package courses

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"cs-educational-platform/backend/internal/auth"
	"cs-educational-platform/backend/internal/db"
	"cs-educational-platform/backend/internal/httputil"
)

// courseFields is the shared SELECT / RETURNING column list for the courses table.
const courseFields = `id, title, description, created_by, is_published, created_at, updated_at`

// courseOwner returns pgx.ErrNoRows if the course does not exist.
func courseOwner(ctx context.Context, id string) (string, error) {
	var owner string
	err := db.Pool.QueryRow(ctx, `SELECT created_by FROM courses WHERE id = $1`, id).Scan(&owner)
	return owner, err
}

// scanCourse works with both pgx.Row.Scan and pgx.Rows.Scan.
func scanCourse(c *Course, scan func(...any) error) error {
	return scan(&c.ID, &c.Title, &c.Description, &c.CreatedBy, &c.IsPublished, &c.CreatedAt, &c.UpdatedAt)
}

// ── Handlers ────────────────────────────────────────────────────────────────

// ListHandler GET /courses
// Professors and admins see all courses; others see only published ones.
func ListHandler(w http.ResponseWriter, r *http.Request) {
	role := auth.RoleFromCtx(r.Context())

	query := `SELECT ` + courseFields + ` FROM courses`
	if role != auth.RoleProfessor && role != auth.RoleAdmin {
		query += ` WHERE is_published = true`
	}
	query += ` ORDER BY created_at DESC LIMIT 200`

	rows, err := db.Pool.Query(r.Context(), query)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := make([]Course, 0)
	for rows.Next() {
		var c Course
		if err := scanCourse(&c, rows.Scan); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		result = append(result, c)
	}
	if rows.Err() != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// CreateHandler POST /courses  [professor, admin]
func CreateHandler(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		httputil.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	var c Course
	row := db.Pool.QueryRow(r.Context(),
		`INSERT INTO courses (title, description, created_by, is_published)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+courseFields,
		req.Title, req.Description, auth.UserIDFromCtx(r.Context()), req.IsPublished,
	)
	if err := scanCourse(&c, row.Scan); err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, c)
}

// GetHandler GET /courses/{id}
// Unpublished courses return 404 for students and TAs (creator/professor/admin can still access).
func GetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	var c Course
	err := scanCourse(&c, db.Pool.QueryRow(r.Context(),
		`SELECT `+courseFields+` FROM courses WHERE id = $1`, id,
	).Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "course not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// POST /courses is gated to professor/admin, so c.CreatedBy will always be a
	// privileged role — but guard explicitly for correctness if that changes.
	canSeeUnpublished := role == auth.RoleProfessor || role == auth.RoleAdmin || c.CreatedBy == userID
	if !c.IsPublished && !canSeeUnpublished {
		httputil.Error(w, "course not found", http.StatusNotFound)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, c)
}

// UpdateHandler PUT /courses/{id}  [professor (owner), admin]
// Only provided fields are updated; omitted fields are left unchanged via COALESCE.
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	owner, err := courseOwner(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "course not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if owner != userID && role != auth.RoleAdmin {
		httputil.Error(w, "forbidden: only the course creator or an admin can update this course", http.StatusForbidden)
		return
	}

	var c Course
	row := db.Pool.QueryRow(r.Context(),
		`UPDATE courses SET
		     title        = COALESCE($1, title),
		     description  = COALESCE($2, description),
		     is_published = COALESCE($3, is_published)
		 WHERE id = $4
		 RETURNING `+courseFields,
		req.Title, req.Description, req.IsPublished, id,
	)
	if err = scanCourse(&c, row.Scan); errors.Is(err, pgx.ErrNoRows) {
		// Course was deleted between the ownership check and the UPDATE.
		httputil.Error(w, "course not found", http.StatusNotFound)
		return
	} else if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, c)
}

// DeleteHandler DELETE /courses/{id}  [professor (owner), admin]
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	owner, err := courseOwner(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "course not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if owner != userID && role != auth.RoleAdmin {
		httputil.Error(w, "forbidden: only the course creator or an admin can delete this course", http.StatusForbidden)
		return
	}

	if _, err = db.Pool.Exec(r.Context(), `DELETE FROM courses WHERE id = $1`, id); err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EnrollHandler POST /courses/{id}/enroll  [student, teaching_assistant]
func EnrollHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	userID := auth.UserIDFromCtx(r.Context())

	var req enrollRequest
	// Body is optional — a missing body enrolls as student by default.
	// Only reject genuinely malformed JSON, not an empty body (io.EOF).
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = auth.RoleStudent
	}
	if req.Role != auth.RoleStudent && req.Role != auth.RoleTeachingAssistant {
		httputil.Error(w, "role must be student or teaching_assistant", http.StatusBadRequest)
		return
	}

	var isPublished bool
	err := db.Pool.QueryRow(r.Context(),
		`SELECT is_published FROM courses WHERE id = $1`, courseID,
	).Scan(&isPublished)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "course not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !isPublished {
		httputil.Error(w, "course is not available for enrollment", http.StatusForbidden)
		return
	}

	_, err = db.Pool.Exec(r.Context(),
		`INSERT INTO course_enrollments (user_id, course_id, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, course_id) DO NOTHING`,
		userID, courseID, req.Role,
	)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MembersHandler GET /courses/{id}/members  [professor, teaching_assistant, admin]
// Uses a single LEFT JOIN to detect course-not-found vs empty-enrollment in one query.
func MembersHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")

	rows, err := db.Pool.Query(r.Context(),
		`SELECT u.id, u.first_name, u.last_name, ce.role, ce.enrolled_at
		 FROM courses c
		 LEFT JOIN course_enrollments ce ON ce.course_id = c.id
		 LEFT JOIN users u ON u.id = ce.user_id
		 WHERE c.id = $1
		 ORDER BY ce.enrolled_at ASC NULLS LAST
		 LIMIT 500`,
		courseID,
	)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	members := make([]Member, 0)
	courseFound := false

	for rows.Next() {
		courseFound = true

		// Nullable — LEFT JOIN returns a single row of NULLs when course has no members.
		var (
			userID     *string
			firstName  *string
			lastName   *string
			role       *string
			enrolledAt *time.Time
		)
		if err := rows.Scan(&userID, &firstName, &lastName, &role, &enrolledAt); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if userID == nil {
			continue // course exists but has no enrolled members
		}
		members = append(members, Member{
			UserID:     *userID,
			FirstName:  *firstName,
			LastName:   *lastName,
			Role:       *role,
			EnrolledAt: *enrolledAt,
		})
	}
	if rows.Err() != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if !courseFound {
		httputil.Error(w, "course not found", http.StatusNotFound)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, members)
}
