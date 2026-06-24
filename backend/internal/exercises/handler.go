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
	exercise_type, quiz_options, quiz_correct, quiz_allow_multiple, created_at, updated_at`

// exerciseCourseAndOwner returns the course_id and created_by for an exercise.
// Returns pgx.ErrNoRows if not found.
func exerciseCourseAndOwner(ctx context.Context, id string) (courseID, ownerID string, err error) {
	err = db.Pool.QueryRow(ctx,
		`SELECT course_id, created_by FROM exercises WHERE id = $1`, id,
	).Scan(&courseID, &ownerID)
	return
}

// testCaseCourseAndOwner returns the course_id and exercise created_by for a test case.
// Returns pgx.ErrNoRows if the test case does not exist.
func testCaseCourseAndOwner(ctx context.Context, testCaseID string) (courseID, ownerID string, err error) {
	err = db.Pool.QueryRow(ctx,
		`SELECT e.course_id, e.created_by FROM test_cases tc
		 JOIN exercises e ON e.id = tc.exercise_id
		 WHERE tc.id = $1`,
		testCaseID,
	).Scan(&courseID, &ownerID)
	return
}

// canModifyInCourse checks write permission for an exercise or test case in courseID.
// Admin: always. Professor: must own the course. Student: must be enrolled as course TA.
func canModifyInCourse(ctx context.Context, role, courseID, userID, ownerID string) (bool, error) {
	switch role {
	case auth.RoleAdmin:
		return true, nil
	case auth.RoleProfessor:
		return ownerID == userID, nil
	case auth.RoleStudent:
		return isEnrolledAsTA(ctx, courseID, userID)
	default:
		return false, nil
	}
}

// scanExercise fills an Exercise from the 18 standard columns.
func scanExercise(e *Exercise, scan func(...any) error) error {
	var qoBytes, qcBytes []byte
	err := scan(
		&e.ID, &e.CourseID, &e.CreatedBy, &e.Title, &e.Description, &e.Instructions,
		&e.Difficulty, &e.Language, &e.TemplateCode, &e.TimeLimitMs, &e.MemoryLimitKb,
		&e.IsPublished, &e.ExerciseType, &qoBytes, &qcBytes, &e.QuizAllowMultiple,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if len(qoBytes) > 0 {
		_ = json.Unmarshal(qoBytes, &e.QuizOptions)
	}
	if len(qcBytes) > 0 {
		_ = json.Unmarshal(qcBytes, &e.QuizCorrect)
	}
	return nil
}

// quizOptionsParam returns a string (JSON) suitable for a ::jsonb parameter,
// or nil (SQL NULL, keeps existing value) if opts is empty.
func quizOptionsParam(opts []string) interface{} {
	if len(opts) == 0 {
		return nil
	}
	b, err := json.Marshal(opts)
	if err != nil {
		return nil
	}
	return string(b)
}

// quizCorrectParam returns a JSON string for the quiz_correct ::jsonb parameter,
// or nil (SQL NULL, keeps existing value) if indices is empty.
func quizCorrectParam(indices []int) interface{} {
	if len(indices) == 0 {
		return nil
	}
	b, err := json.Marshal(indices)
	if err != nil {
		return nil
	}
	return string(b)
}

// canManageExercise reports whether this role can create/edit exercises in a course.
func canManageExercise(role string) bool {
	return role == auth.RoleProfessor || role == auth.RoleAdmin
}

// requiresEnrollment reports whether this role must be enrolled to access exercises.
// Professors and admins always have full access regardless of enrollment.
func requiresEnrollment(role string) bool {
	return role == auth.RoleStudent
}

// isEnrolled checks whether userID is enrolled in courseID (any role).
func isEnrolled(ctx context.Context, courseID, userID string) (bool, error) {
	var enrolled bool
	err := db.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM course_enrollments WHERE course_id = $1 AND user_id = $2)`,
		courseID, userID,
	).Scan(&enrolled)
	return enrolled, err
}

// isEnrolledAsTA checks whether userID is enrolled in courseID with enrollment role 'teaching_assistant'.
func isEnrolledAsTA(ctx context.Context, courseID, userID string) (bool, error) {
	var ok bool
	err := db.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM course_enrollments WHERE course_id = $1 AND user_id = $2 AND role = $3)`,
		courseID, userID, auth.RoleEnrollmentTA,
	).Scan(&ok)
	return ok, err
}

// ── Exercise handlers ────────────────────────────────────────────────────────

// ListHandler GET /courses/{id}/exercises
// Professors and admins see all; others only see published.
// Students and TAs must be enrolled in the course — otherwise 403.
func ListHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

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

	// Students and TAs must be enrolled to browse exercises.
	if requiresEnrollment(role) {
		enrolled, err := isEnrolled(r.Context(), courseID, userID)
		if err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if !enrolled {
			httputil.Error(w, "you must be enrolled in this course to view its exercises", http.StatusForbidden)
			return
		}
	}

	// Students enrolled as course TAs can see unpublished exercises.
	showAll := canManageExercise(role)
	if !showAll && role == auth.RoleStudent {
		ta, err := isEnrolledAsTA(r.Context(), courseID, userID)
		if err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		showAll = ta
	}

	query := `SELECT ` + exerciseFields + ` FROM exercises WHERE course_id = $1`
	if !showAll {
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

// CreateHandler POST /courses/{id}/exercises  [professor (course owner), enrolled TA, admin]
func CreateHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	userID := auth.UserIDFromCtx(r.Context())
	role := auth.RoleFromCtx(r.Context())

	// Verify the course exists and fetch its owner.
	var courseCreatedBy string
	if err := db.Pool.QueryRow(r.Context(),
		`SELECT created_by FROM courses WHERE id = $1`, courseID,
	).Scan(&courseCreatedBy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Error(w, "course not found", http.StatusNotFound)
			return
		}
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	// Admin always allowed; TA must be enrolled; professor must own the course.
	ok, err := canModifyInCourse(r.Context(), role, courseID, userID, courseCreatedBy)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		httputil.Error(w, "forbidden: only the course creator, an enrolled TA, or an admin can add exercises", http.StatusForbidden)
		return
	}

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
	if req.ExerciseType == "" {
		req.ExerciseType = "coding"
	}
	if req.ExerciseType == "coding" && req.Language == "" {
		httputil.Error(w, "language is required for coding exercises", http.StatusBadRequest)
		return
	}
	if req.ExerciseType == "quiz" {
		if len(req.QuizOptions) < 2 {
			httputil.Error(w, "quiz exercises require at least 2 options", http.StatusBadRequest)
			return
		}
		if len(req.QuizCorrect) == 0 {
			httputil.Error(w, "quiz exercises require at least one correct answer", http.StatusBadRequest)
			return
		}
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

	// Language is NULL for quiz exercises.
	var lang *string
	if req.ExerciseType == "coding" {
		lang = &req.Language
	}

	var e Exercise
	row := db.Pool.QueryRow(r.Context(),
		`INSERT INTO exercises
		   (course_id, created_by, title, description, instructions,
		    difficulty, language, template_code, time_limit_ms, memory_limit_kb, is_published,
		    exercise_type, quiz_options, quiz_correct, quiz_allow_multiple)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13::jsonb,$14::jsonb,$15)
		 RETURNING `+exerciseFields,
		courseID, userID, req.Title, req.Description, req.Instructions,
		req.Difficulty, lang, req.TemplateCode, timeLimitMs, memoryLimitKb, req.IsPublished,
		req.ExerciseType, quizOptionsParam(req.QuizOptions), quizCorrectParam(req.QuizCorrect), req.QuizAllowMultiple,
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

	// Students enrolled as course TA can see unpublished exercises.
	canSeeUnpublished := canManageExercise(role) || e.CreatedBy == userID
	if !e.IsPublished && !canSeeUnpublished && role == auth.RoleStudent {
		ta, err := isEnrolledAsTA(r.Context(), e.CourseID, userID)
		if err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		canSeeUnpublished = ta
	}
	if !e.IsPublished && !canSeeUnpublished {
		httputil.Error(w, "exercise not found", http.StatusNotFound)
		return
	}

	// Students and TAs must be enrolled in the exercise's course.
	if requiresEnrollment(role) {
		enrolled, err := isEnrolled(r.Context(), e.CourseID, userID)
		if err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if !enrolled {
			httputil.Error(w, "you must be enrolled in this course to view this exercise", http.StatusForbidden)
			return
		}
	}

	httputil.WriteJSON(w, http.StatusOK, e)
}

// UpdateHandler PUT /exercises/{id}  [professor (owner), enrolled TA, admin]
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	var req updateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	courseID, owner, err := exerciseCourseAndOwner(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	ok, err := canModifyInCourse(r.Context(), role, courseID, userID, owner)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		httputil.Error(w, "forbidden: only the exercise creator, an enrolled TA, or an admin can update this exercise", http.StatusForbidden)
		return
	}

	var e Exercise
	row := db.Pool.QueryRow(r.Context(),
		`UPDATE exercises SET
		     title               = COALESCE($1,  title),
		     description         = COALESCE($2,  description),
		     instructions        = COALESCE($3,  instructions),
		     difficulty          = COALESCE($4,  difficulty),
		     exercise_type       = COALESCE($5,  exercise_type),
		     language            = COALESCE($6,  language),
		     template_code       = COALESCE($7,  template_code),
		     time_limit_ms       = COALESCE($8,  time_limit_ms),
		     memory_limit_kb     = COALESCE($9,  memory_limit_kb),
		     is_published        = COALESCE($10, is_published),
		     quiz_options        = COALESCE($11::jsonb, quiz_options),
		     quiz_correct        = COALESCE($12::jsonb, quiz_correct),
		     quiz_allow_multiple = COALESCE($13, quiz_allow_multiple)
		 WHERE id = $14
		 RETURNING `+exerciseFields,
		req.Title, req.Description, req.Instructions, req.Difficulty, req.ExerciseType,
		req.Language, req.TemplateCode, req.TimeLimitMs, req.MemoryLimitKb, req.IsPublished,
		quizOptionsParam(req.QuizOptions), quizCorrectParam(req.QuizCorrect),
		req.QuizAllowMultiple, id,
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

// DeleteHandler DELETE /exercises/{id}  [professor (owner), enrolled TA, admin]
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	courseID, owner, err := exerciseCourseAndOwner(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	ok, err := canModifyInCourse(r.Context(), role, courseID, userID, owner)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		httputil.Error(w, "forbidden: only the exercise creator, an enrolled TA, or an admin can delete this exercise", http.StatusForbidden)
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
// Students and TAs must be enrolled in the exercise's course.
func ListTestCasesHandler(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	// Fetch exercise to confirm it exists and to get course_id for enrollment check.
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

	// Students and TAs must be enrolled in the course.
	if requiresEnrollment(role) {
		enrolled, err := isEnrolled(r.Context(), courseID, userID)
		if err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if !enrolled {
			httputil.Error(w, "you must be enrolled in this course to view test cases", http.StatusForbidden)
			return
		}
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

	privileged := canManageExercise(role)
	if !privileged && role == auth.RoleStudent {
		ta, err := isEnrolledAsTA(r.Context(), courseID, userID)
		if err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		privileged = ta
	}

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

// CreateTestCaseHandler POST /exercises/{id}/test-cases  [professor (owner), enrolled TA, admin]
func CreateTestCaseHandler(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	courseID, owner, err := exerciseCourseAndOwner(r.Context(), exerciseID)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	ok, err := canModifyInCourse(r.Context(), role, courseID, userID, owner)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		httputil.Error(w, "forbidden: only the exercise creator, an enrolled TA, or an admin can add test cases", http.StatusForbidden)
		return
	}

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
	err = db.Pool.QueryRow(r.Context(),
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

// UpdateTestCaseHandler PUT /test-cases/{id}  [professor (owner), enrolled TA, admin]
func UpdateTestCaseHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	courseID, owner, err := testCaseCourseAndOwner(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "test case not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	ok, err := canModifyInCourse(r.Context(), role, courseID, userID, owner)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		httputil.Error(w, "forbidden: only the exercise creator, an enrolled TA, or an admin can modify test cases", http.StatusForbidden)
		return
	}

	var req updateTestCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	var tc TestCase
	err = db.Pool.QueryRow(r.Context(),
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

// DeleteTestCaseHandler DELETE /test-cases/{id}  [professor (owner), enrolled TA, admin]
func DeleteTestCaseHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())

	courseID, owner, err := testCaseCourseAndOwner(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, "test case not found", http.StatusNotFound)
		return
	}
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	ok, err := canModifyInCourse(r.Context(), role, courseID, userID, owner)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		httputil.Error(w, "forbidden: only the exercise creator, an enrolled TA, or an admin can delete test cases", http.StatusForbidden)
		return
	}

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
