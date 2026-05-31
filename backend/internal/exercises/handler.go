package exercises

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"cs-educational-platform/backend/internal/auth"
	"cs-educational-platform/backend/internal/db"
	"cs-educational-platform/backend/internal/httputil"
)

const exerciseFields = `id, course_id, created_by, title, description, instructions,
	difficulty, language, template_code, time_limit_ms, memory_limit_kb, is_published,
	created_at, updated_at`

// exerciseOwner returns the created_by UUID. Returns pgx.ErrNoRows if not found.
func exerciseOwner(ctx context.Context, id string) (string, error) {
	var owner string
	err := db.Pool.QueryRow(ctx, `SELECT created_by FROM exercises WHERE id = $1`, id).Scan(&owner)
	return owner, err
}

// scanExercise fills an Exercise from the 14 standard columns.
func scanExercise(e *Exercise, scan func(...any) error) error {
	return scan(
		&e.ID, &e.CourseID, &e.CreatedBy, &e.Title, &e.Description, &e.Instructions,
		&e.Difficulty, &e.Language, &e.TemplateCode, &e.TimeLimitMs, &e.MemoryLimitKb,
		&e.IsPublished, &e.CreatedAt, &e.UpdatedAt,
	)
}

// canManageExercise reports whether this role can create/edit exercises in a course.
func canManageExercise(role string) bool {
	return role == auth.RoleProfessor || role == auth.RoleAdmin
}

// ── Exercise handlers ────────────────────────────────────────────────────────

// ListHandler GET /courses/{id}/exercises
// Professors and admins see all; others only see published.
func ListHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())

	// Verify the course exists first.
	var exists bool
	if err := db.Pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1)`, courseID,
	).Scan(&exists); err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		httputil.Error(w, "course not found", http.StatusNotFound)
		return
	}

	query := `SELECT ` + exerciseFields + ` FROM exercises WHERE course_id = $1`
	if !canManageExercise(role) {
		query += ` AND is_published = true`
	}
	query += ` ORDER BY created_at ASC LIMIT 200`

	rows, err := db.Pool.Query(r.Context(), query, courseID)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := make([]Exercise, 0)
	for rows.Next() {
		var e Exercise
		if err := scanExercise(&e, rows.Scan); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		result = append(result, e)
	}
	if rows.Err() != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// CreateHandler POST /courses/{id}/exercises  [professor, admin]
func CreateHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	userID := auth.UserIDFromCtx(r.Context())

	var req createExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		httputil.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if req.Instructions == "" {
		httputil.Error(w, "instructions is required", http.StatusBadRequest)
		return
	}
	if req.Language == "" {
		httputil.Error(w, "language is required", http.StatusBadRequest)
		return
	}
	if req.Difficulty == "" {
		req.Difficulty = "medium"
	}

	// Default limits if not provided.
	timeLimitMs := 2000
	if req.TimeLimitMs != nil {
		timeLimitMs = *req.TimeLimitMs
	}
	memoryLimitKb := 65536
	if req.MemoryLimitKb != nil {
		memoryLimitKb = *req.MemoryLimitKb
	}

	var e Exercise
	row := db.Pool.QueryRow(r.Context(),
		`INSERT INTO exercises
		   (course_id, created_by, title, description, instructions,
		    difficulty, language, template_code, time_limit_ms, memory_limit_kb, is_published)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		 RETURNING `+exerciseFields,
		courseID, userID, req.Title, req.Description, req.Instructions,
		req.Difficulty, req.Language, req.TemplateCode, timeLimitMs, memoryLimitKb, req.IsPublished,
	)
	if err := scanExercise(&e, row.Scan); err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, e)
}

// GetHandler GET /exercises/{id}
// Students only see published exercises.
func GetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	var e Exercise
	err := scanExercise(&e, db.Pool.QueryRow(r.Context(),
		`SELECT `+exerciseFields+` FROM exercises WHERE id = $1`, id,
	).Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	canSeeUnpublished := canManageExercise(role) || e.CreatedBy == userID
	if !e.IsPublished && !canSeeUnpublished {
		httputil.Error(w, "exercise not found", http.StatusNotFound)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, e)
}

// UpdateHandler PUT /exercises/{id}  [professor (owner), admin]
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	var req updateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	owner, err := exerciseOwner(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if owner != userID && role != auth.RoleAdmin {
		httputil.Error(w, "forbidden: only the exercise creator or an admin can update this exercise", http.StatusForbidden)
		return
	}

	var e Exercise
	row := db.Pool.QueryRow(r.Context(),
		`UPDATE exercises SET
		     title          = COALESCE($1,  title),
		     description    = COALESCE($2,  description),
		     instructions   = COALESCE($3,  instructions),
		     difficulty     = COALESCE($4,  difficulty),
		     language       = COALESCE($5,  language),
		     template_code  = COALESCE($6,  template_code),
		     time_limit_ms  = COALESCE($7,  time_limit_ms),
		     memory_limit_kb = COALESCE($8, memory_limit_kb),
		     is_published   = COALESCE($9,  is_published)
		 WHERE id = $10
		 RETURNING `+exerciseFields,
		req.Title, req.Description, req.Instructions, req.Difficulty, req.Language,
		req.TemplateCode, req.TimeLimitMs, req.MemoryLimitKb, req.IsPublished, id,
	)
	if err = scanExercise(&e, row.Scan); errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "exercise not found", http.StatusNotFound)
		return
	} else if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, e)
}

// DeleteHandler DELETE /exercises/{id}  [professor (owner), admin]
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	owner, err := exerciseOwner(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if owner != userID && role != auth.RoleAdmin {
		httputil.Error(w, "forbidden: only the exercise creator or an admin can delete this exercise", http.StatusForbidden)
		return
	}

	if _, err = db.Pool.Exec(r.Context(), `DELETE FROM exercises WHERE id = $1`, id); err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Test case handlers ────────────────────────────────────────────────────────

// ListTestCasesHandler GET /exercises/{id}/test-cases
// Students receive hidden test cases with input/expected_output stripped.
func ListTestCasesHandler(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())

	// Verify exercise exists.
	var exists bool
	if err := db.Pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM exercises WHERE id = $1)`, exerciseID,
	).Scan(&exists); err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		httputil.Error(w, "exercise not found", http.StatusNotFound)
		return
	}

	rows, err := db.Pool.Query(r.Context(),
		`SELECT id, exercise_id, input, expected_output, is_hidden, ordinal, created_at
		 FROM test_cases WHERE exercise_id = $1 ORDER BY ordinal ASC, created_at ASC`,
		exerciseID,
	)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	privileged := canManageExercise(role) || role == auth.RoleTeachingAssistant

	// Use any to hold either TestCase or HiddenTestCase.
	result := make([]any, 0)
	for rows.Next() {
		var tc TestCase
		if err := rows.Scan(&tc.ID, &tc.ExerciseID, &tc.Input, &tc.ExpectedOutput,
			&tc.IsHidden, &tc.Ordinal, &tc.CreatedAt); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if tc.IsHidden && !privileged {
			result = append(result, HiddenTestCase{
				ID:         tc.ID,
				ExerciseID: tc.ExerciseID,
				IsHidden:   true,
				Ordinal:    tc.Ordinal,
				CreatedAt:  tc.CreatedAt,
			})
		} else {
			result = append(result, tc)
		}
	}
	if rows.Err() != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// CreateTestCaseHandler POST /exercises/{id}/test-cases  [professor, admin]
func CreateTestCaseHandler(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")

	var req createTestCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.ExpectedOutput == "" {
		httputil.Error(w, "expected_output is required", http.StatusBadRequest)
		return
	}

	var tc TestCase
	err := db.Pool.QueryRow(r.Context(),
		`INSERT INTO test_cases (exercise_id, input, expected_output, is_hidden, ordinal)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, exercise_id, input, expected_output, is_hidden, ordinal, created_at`,
		exerciseID, req.Input, req.ExpectedOutput, req.IsHidden, req.Ordinal,
	).Scan(&tc.ID, &tc.ExerciseID, &tc.Input, &tc.ExpectedOutput, &tc.IsHidden, &tc.Ordinal, &tc.CreatedAt)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, tc)
}

// UpdateTestCaseHandler PUT /test-cases/{id}  [professor, admin]
func UpdateTestCaseHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req updateTestCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	var tc TestCase
	err := db.Pool.QueryRow(r.Context(),
		`UPDATE test_cases SET
		     input           = COALESCE($1, input),
		     expected_output = COALESCE($2, expected_output),
		     is_hidden       = COALESCE($3, is_hidden),
		     ordinal         = COALESCE($4, ordinal)
		 WHERE id = $5
		 RETURNING id, exercise_id, input, expected_output, is_hidden, ordinal, created_at`,
		req.Input, req.ExpectedOutput, req.IsHidden, req.Ordinal, id,
	).Scan(&tc.ID, &tc.ExerciseID, &tc.Input, &tc.ExpectedOutput, &tc.IsHidden, &tc.Ordinal, &tc.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "test case not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, tc)
}

// DeleteTestCaseHandler DELETE /test-cases/{id}  [professor, admin]
func DeleteTestCaseHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	tag, err := db.Pool.Exec(r.Context(), `DELETE FROM test_cases WHERE id = $1`, id)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		httputil.Error(w, "test case not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
