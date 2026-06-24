package courses

import (
	"context"
	"encoding/json"
	"errors"
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
// Used by CreateHandler and UpdateHandler (RETURNING courseFields — no is_enrolled column).
func scanCourse(c *Course, scan func(...any) error) error {
	return scan(&c.ID, &c.Title, &c.Description, &c.CreatedBy, &c.IsPublished, &c.CreatedAt, &c.UpdatedAt)
}

// scanCourseWithEnrollment scans courseFields + the trailing is_enrolled boolean.
// Used by ListHandler and GetHandler which append the EXISTS subquery.
func scanCourseWithEnrollment(c *Course, scan func(...any) error) error {
	return scan(&c.ID, &c.Title, &c.Description, &c.CreatedBy, &c.IsPublished, &c.CreatedAt, &c.UpdatedAt,
		&c.IsEnrolled)
}

// ── Handlers ────────────────────────────────────────────────────────────────

// ListHandler GET /courses
// Professors and admins see all courses; others see only published ones.
// The is_enrolled flag reflects whether the requesting user is enrolled in each course.
func ListHandler(w http.ResponseWriter, r *http.Request) {
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	// Use table alias `c` so the EXISTS subquery can reference it unambiguously.
	query := `SELECT ` + courseFields + `,
		EXISTS(SELECT 1 FROM course_enrollments WHERE course_id = c.id AND user_id = $1) AS is_enrolled
		FROM courses c`
	switch role {
	case auth.RoleProfessor:
		query += ` WHERE c.created_by = $1`
	case auth.RoleAdmin:
		// admins see every course, no WHERE needed
	default:
		query += ` WHERE c.is_published = true`
	}
	query += ` ORDER BY c.created_at DESC LIMIT 200`

	rows, err := db.Pool.Query(r.Context(), query, userID)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := make([]Course, 0)
	for rows.Next() {
		var c Course
		if err := scanCourseWithEnrollment(&c, rows.Scan); err != nil {
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

	// Populate is_enrolled for the requesting user.
	// Ignore any error — worst case the flag stays false.
	_ = db.Pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM course_enrollments WHERE course_id = $1 AND user_id = $2)`,
		id, userID,
	).Scan(&c.IsEnrolled)

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

// EnrollHandler POST /courses/{id}/enroll  [student]
// Students always enroll with enrollment_role='student'.
// Professors promote them to course TA via PATCH /courses/{id}/members/{userId}/role.
func EnrollHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	userID := auth.UserIDFromCtx(r.Context())

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
		 VALUES ($1, $2, 'student')
		 ON CONFLICT (user_id, course_id) DO NOTHING`,
		userID, courseID,
	)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// LeaderboardHandler GET /courses/{id}/leaderboard
// Returns enrolled students ranked by the number of accepted exercises.
func LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")

	rows, err := db.Pool.Query(r.Context(),
		`SELECT u.id, u.first_name, u.last_name,
		        COUNT(DISTINCT s.exercise_id) AS solved_count
		 FROM users u
		 JOIN course_enrollments ce ON ce.user_id = u.id
		                           AND ce.course_id = $1
		                           AND ce.role = 'student'
		 LEFT JOIN submissions s ON s.user_id = u.id
		                        AND s.status = 'accepted'
		                        AND s.exercise_id IN (
		                            SELECT id FROM exercises
		                            WHERE course_id = $1 AND is_published = true
		                        )
		 GROUP BY u.id, u.first_name, u.last_name
		 ORDER BY solved_count DESC, u.first_name ASC
		 LIMIT 100`,
		courseID,
	)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	entries := make([]LeaderboardEntry, 0)
	rank := 1
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.FirstName, &e.LastName, &e.SolvedCount); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		e.Rank = rank
		rank++
		entries = append(entries, e)
	}
	if rows.Err() != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, entries)
}

// MembersHandler GET /courses/{id}/members  [professor (owner), admin, enrolled course TA]
// Any enrolled student may call this to discover their own enrollment role.
// Access is checked inside: professor/admin see everyone; enrolled students see all members.
func MembersHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	if role != auth.RoleAdmin && role != auth.RoleProfessor {
		// Students must be enrolled in the course.
		var enrolled bool
		if err := db.Pool.QueryRow(r.Context(),
			`SELECT EXISTS(SELECT 1 FROM course_enrollments WHERE course_id = $1 AND user_id = $2)`,
			courseID, userID,
		).Scan(&enrolled); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if !enrolled {
			httputil.Error(w, "forbidden: you must be enrolled in this course", http.StatusForbidden)
			return
		}
	}

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

// UpdateMemberRoleHandler PATCH /courses/{id}/members/{userId}/role
// Only the course owner (professor) or an admin can promote/demote an enrolled member.
// Body: { "role": "student" | "teaching_assistant" }
func UpdateMemberRoleHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	targetUserID := r.PathValue("userId")
	callerID := auth.UserIDFromCtx(r.Context())
	callerRole := auth.RoleFromCtx(r.Context())

	owner, err := courseOwner(r.Context(), courseID)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "course not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if callerRole != auth.RoleAdmin && owner != callerID {
		httputil.Error(w, "only the course owner or an admin can change member roles", http.StatusForbidden)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.Role != auth.RoleStudent && body.Role != auth.RoleEnrollmentTA {
		httputil.Error(w, "role must be 'student' or 'teaching_assistant'", http.StatusBadRequest)
		return
	}

	tag, err := db.Pool.Exec(r.Context(),
		`UPDATE course_enrollments SET role = $1 WHERE course_id = $2 AND user_id = $3`,
		body.Role, courseID, targetUserID,
	)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		httputil.Error(w, "user is not enrolled in this course", http.StatusNotFound)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"role": body.Role})
}
