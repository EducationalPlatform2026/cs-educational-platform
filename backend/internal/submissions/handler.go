package submissions

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"cs-educational-platform/backend/internal/auth"
	"cs-educational-platform/backend/internal/db"
	"cs-educational-platform/backend/internal/httputil"
)

// submissionPrivileged reports whether this role can see all submissions (not just their own).
func submissionPrivileged(role string) bool {
	return role == auth.RoleProfessor || role == auth.RoleAdmin || role == auth.RoleTeachingAssistant
}

const submissionFields = `id, exercise_id, user_id, code, language, status, score, stderr, submitted_at`

func scanSubmission(s *Submission, scan func(...any) error) error {
	return scan(&s.ID, &s.ExerciseID, &s.UserID, &s.Code, &s.Language,
		&s.Status, &s.Score, &s.Stderr, &s.SubmittedAt)
}

// SubmitHandler POST /exercises/{id}/submit  [any authenticated user]
// Stores the submission as pending. The sandbox will evaluate it asynchronously.
func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")
	userID := auth.UserIDFromCtx(r.Context())

	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Code == "" {
		httputil.Error(w, "code is required", http.StatusBadRequest)
		return
	}
	if req.Language == "" {
		httputil.Error(w, "language is required", http.StatusBadRequest)
		return
	}

	// Verify the exercise exists and fetch its course_id + published flag.
	var isPublished bool
	var courseID string
	err := db.Pool.QueryRow(r.Context(),
		`SELECT is_published, course_id FROM exercises WHERE id = $1`, exerciseID,
	).Scan(&isPublished, &courseID)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	role := auth.RoleFromCtx(r.Context())
	privileged := role == auth.RoleProfessor || role == auth.RoleAdmin

	if !isPublished && !privileged {
		httputil.Error(w, "exercise is not available for submission", http.StatusForbidden)
		return
	}

	// Students and TAs must be enrolled to submit.
	if !privileged {
		var enrolled bool
		if err := db.Pool.QueryRow(r.Context(),
			`SELECT EXISTS(SELECT 1 FROM course_enrollments WHERE course_id = $1 AND user_id = $2)`,
			courseID, userID,
		).Scan(&enrolled); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if !enrolled {
			httputil.Error(w, "you must be enrolled in this course to submit", http.StatusForbidden)
			return
		}
	}

	var s Submission
	row := db.Pool.QueryRow(r.Context(),
		`INSERT INTO submissions (exercise_id, user_id, code, language)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+submissionFields,
		exerciseID, userID, req.Code, req.Language,
	)
	if err := scanSubmission(&s, row.Scan); err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, s)
}

// ListHandler GET /exercises/{id}/submissions  [any authenticated user]
// Students and TAs must be enrolled; students see only their own submissions.
func ListHandler(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")
	userID := auth.UserIDFromCtx(r.Context())
	role := auth.RoleFromCtx(r.Context())

	// Verify the exercise exists and fetch course_id for enrollment check.
	var courseID string
	if err := db.Pool.QueryRow(r.Context(),
		`SELECT course_id FROM exercises WHERE id = $1`, exerciseID,
	).Scan(&courseID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Error(w, "exercise not found", http.StatusNotFound)
			return
		}
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	privileged := role == auth.RoleProfessor || role == auth.RoleAdmin || role == auth.RoleTeachingAssistant

	// Students and TAs must be enrolled in the course to view submissions.
	if !privileged {
		var enrolled bool
		if err := db.Pool.QueryRow(r.Context(),
			`SELECT EXISTS(SELECT 1 FROM course_enrollments WHERE course_id = $1 AND user_id = $2)`,
			courseID, userID,
		).Scan(&enrolled); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if !enrolled {
			httputil.Error(w, "you must be enrolled in this course to view submissions", http.StatusForbidden)
			return
		}
	}

	var (
		rows interface {
			Next() bool
			Scan(...any) error
			Close()
			Err() error
		}
		err error
	)

	if privileged {
		rows, err = db.Pool.Query(r.Context(),
			`SELECT `+submissionFields+` FROM submissions
			 WHERE exercise_id = $1
			 ORDER BY submitted_at DESC LIMIT 200`,
			exerciseID,
		)
	} else {
		rows, err = db.Pool.Query(r.Context(),
			`SELECT `+submissionFields+` FROM submissions
			 WHERE exercise_id = $1 AND user_id = $2
			 ORDER BY submitted_at DESC LIMIT 50`,
			exerciseID, userID,
		)
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := make([]Submission, 0)
	for rows.Next() {
		var s Submission
		if err := scanSubmission(&s, rows.Scan); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		result = append(result, s)
	}
	if rows.Err() != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetHandler GET /submissions/{id}  [any authenticated user]
// Students can only retrieve their own submissions.
func GetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := auth.UserIDFromCtx(r.Context())
	role := auth.RoleFromCtx(r.Context())

	var s Submission
	err := scanSubmission(&s, db.Pool.QueryRow(r.Context(),
		`SELECT `+submissionFields+` FROM submissions WHERE id = $1`, id,
	).Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "submission not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	privileged := role == auth.RoleProfessor || role == auth.RoleAdmin || role == auth.RoleTeachingAssistant
	if s.UserID != userID && !privileged {
		httputil.Error(w, "submission not found", http.StatusNotFound)
		return
	}

	// Fetch per-test-case results.
	rows, err := db.Pool.Query(r.Context(),
		`SELECT id, submission_id, test_case_id, status, actual_output, runtime_ms, memory_kb
		 FROM submission_results WHERE submission_id = $1 ORDER BY id`,
		id,
	)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := make([]SubmissionResult, 0)
	for rows.Next() {
		var r SubmissionResult
		if err := rows.Scan(&r.ID, &r.SubmissionID, &r.TestCaseID, &r.Status,
			&r.ActualOutput, &r.RuntimeMs, &r.MemoryKb); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		results = append(results, r)
	}
	if rows.Err() != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, SubmissionDetail{Submission: s, Results: results})
}
