// Package coursestats provides the GET /courses/{id}/stats endpoint.
// Access: course creator (professor), enrolled TAs, and admins.
package coursestats

import (
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"cs-educational-platform/backend/internal/auth"
	"cs-educational-platform/backend/internal/db"
	"cs-educational-platform/backend/internal/httputil"
)

// ── Response types ───────────────────────────────────────────────────────────

type StaffMember struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type StudentStat struct {
	UserID              string     `json:"user_id"`
	FirstName           string     `json:"first_name"`
	LastName            string     `json:"last_name"`
	Email               string     `json:"email"`
	EnrolledAt          time.Time  `json:"enrolled_at"`
	ExercisesAttempted  int        `json:"exercises_attempted"`
	ExercisesSolved     int        `json:"exercises_solved"`
	TotalSubmissions    int        `json:"total_submissions"`
	AcceptedSubmissions int        `json:"accepted_submissions"`
	AcceptanceRate      float64    `json:"acceptance_rate"` // 0–100
	LastSubmissionAt    *time.Time `json:"last_submission_at,omitempty"`
}

type ExerciseStat struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Difficulty        string  `json:"difficulty"`
	StudentsAttempted int     `json:"students_attempted"`
	StudentsSolved    int     `json:"students_solved"`
	SolveRate         float64 `json:"solve_rate"` // 0–100, among enrolled students
	TotalSubmissions  int     `json:"total_submissions"`
}

type Summary struct {
	TotalStudents          int     `json:"total_students"`
	TotalTeachingAssts     int     `json:"total_teaching_assistants"`
	TotalExercises         int     `json:"total_exercises"`
	TotalSubmissions       int     `json:"total_submissions"`
	AcceptedSubmissions    int     `json:"accepted_submissions"`
	CourseAcceptanceRate   float64 `json:"course_acceptance_rate"` // 0–100
	StudentsWithSubmission int     `json:"students_with_submission"`
}

type CourseStats struct {
	Professor         StaffMember    `json:"professor"`
	TeachingAssistants []StaffMember `json:"teaching_assistants"`
	Summary           Summary        `json:"summary"`
	Students          []StudentStat  `json:"students"`
	Exercises         []ExerciseStat `json:"exercises"`
}

// ── Handler ──────────────────────────────────────────────────────────────────

// Handler GET /courses/{id}/stats
// Accessible by: course creator, enrolled TAs, admins.
func Handler(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	role := auth.RoleFromCtx(r.Context())
	userID := auth.UserIDFromCtx(r.Context())
	ctx := r.Context()

	// ── Access control ────────────────────────────────────────────────────────
	// Admins always pass. Professors must own the course. TAs must be enrolled.
	if role != auth.RoleAdmin {
		var createdBy string
		err := db.Pool.QueryRow(ctx,
			`SELECT created_by FROM courses WHERE id = $1`, courseID,
		).Scan(&createdBy)
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Error(w, "course not found", http.StatusNotFound)
			return
		}
		if err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if role == auth.RoleProfessor {
			if createdBy != userID {
				httputil.Error(w, "forbidden: only the course owner can view stats", http.StatusForbidden)
				return
			}
		} else if role == auth.RoleTeachingAssistant {
			var enrolled bool
			if err := db.Pool.QueryRow(ctx,
				`SELECT EXISTS(SELECT 1 FROM course_enrollments WHERE course_id = $1 AND user_id = $2)`,
				courseID, userID,
			).Scan(&enrolled); err != nil || !enrolled {
				httputil.Error(w, "forbidden: not enrolled in this course", http.StatusForbidden)
				return
			}
		} else {
			httputil.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	var stats CourseStats

	// ── Professor (course creator) ────────────────────────────────────────────
	if err := db.Pool.QueryRow(ctx,
		`SELECT u.id, u.first_name, u.last_name, u.email
		 FROM users u JOIN courses c ON c.created_by = u.id
		 WHERE c.id = $1`, courseID,
	).Scan(&stats.Professor.UserID, &stats.Professor.FirstName,
		&stats.Professor.LastName, &stats.Professor.Email); err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// ── Teaching assistants ───────────────────────────────────────────────────
	taRows, err := db.Pool.Query(ctx,
		`SELECT u.id, u.first_name, u.last_name, u.email
		 FROM users u
		 JOIN course_enrollments ce ON ce.user_id = u.id
		 WHERE ce.course_id = $1 AND ce.role = 'teaching_assistant'
		 ORDER BY u.last_name, u.first_name`, courseID,
	)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer taRows.Close()
	stats.TeachingAssistants = make([]StaffMember, 0)
	for taRows.Next() {
		var ta StaffMember
		if err := taRows.Scan(&ta.UserID, &ta.FirstName, &ta.LastName, &ta.Email); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		stats.TeachingAssistants = append(stats.TeachingAssistants, ta)
	}
	if taRows.Err() != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// ── Student roster with per-student aggregates ────────────────────────────
	studentRows, err := db.Pool.Query(ctx,
		`SELECT
		     u.id, u.first_name, u.last_name, u.email, ce.enrolled_at,
		     COUNT(DISTINCT s.exercise_id)                                        AS exercises_attempted,
		     COUNT(DISTINCT s.exercise_id) FILTER (WHERE s.status = 'accepted')  AS exercises_solved,
		     COUNT(s.id)                                                          AS total_submissions,
		     COUNT(s.id)                   FILTER (WHERE s.status = 'accepted')  AS accepted_submissions,
		     MAX(s.submitted_at)                                                  AS last_submission_at
		 FROM course_enrollments ce
		 JOIN users u ON u.id = ce.user_id
		 LEFT JOIN submissions s
		     ON s.user_id = ce.user_id
		    AND s.exercise_id IN (SELECT id FROM exercises WHERE course_id = $1)
		 WHERE ce.course_id = $1 AND ce.role = 'student'
		 GROUP BY u.id, u.first_name, u.last_name, u.email, ce.enrolled_at
		 ORDER BY exercises_solved DESC, accepted_submissions DESC, u.last_name`,
		courseID,
	)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer studentRows.Close()

	stats.Students = make([]StudentStat, 0)
	for studentRows.Next() {
		var s StudentStat
		if err := studentRows.Scan(
			&s.UserID, &s.FirstName, &s.LastName, &s.Email, &s.EnrolledAt,
			&s.ExercisesAttempted, &s.ExercisesSolved,
			&s.TotalSubmissions, &s.AcceptedSubmissions,
			&s.LastSubmissionAt,
		); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if s.TotalSubmissions > 0 {
			s.AcceptanceRate = float64(s.AcceptedSubmissions) / float64(s.TotalSubmissions) * 100
		}
		stats.Students = append(stats.Students, s)
	}
	if studentRows.Err() != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// ── Per-exercise stats ────────────────────────────────────────────────────
	totalStudents := len(stats.Students)
	exRows, err := db.Pool.Query(ctx,
		`SELECT
		     e.id, e.title, e.difficulty,
		     COUNT(DISTINCT s.user_id)                                          AS students_attempted,
		     COUNT(DISTINCT s.user_id) FILTER (WHERE s.status = 'accepted')    AS students_solved,
		     COUNT(s.id)                                                        AS total_submissions
		 FROM exercises e
		 LEFT JOIN submissions s ON s.exercise_id = e.id
		 WHERE e.course_id = $1 AND e.is_published = true
		 GROUP BY e.id, e.title, e.difficulty
		 ORDER BY e.created_at ASC`,
		courseID,
	)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer exRows.Close()

	stats.Exercises = make([]ExerciseStat, 0)
	for exRows.Next() {
		var ex ExerciseStat
		if err := exRows.Scan(&ex.ID, &ex.Title, &ex.Difficulty,
			&ex.StudentsAttempted, &ex.StudentsSolved, &ex.TotalSubmissions); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if totalStudents > 0 {
			ex.SolveRate = float64(ex.StudentsSolved) / float64(totalStudents) * 100
		}
		stats.Exercises = append(stats.Exercises, ex)
	}
	if exRows.Err() != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// ── Course-level summary ──────────────────────────────────────────────────
	stats.Summary.TotalStudents = totalStudents
	stats.Summary.TotalTeachingAssts = len(stats.TeachingAssistants)
	stats.Summary.TotalExercises = len(stats.Exercises)

	for _, s := range stats.Students {
		stats.Summary.TotalSubmissions += s.TotalSubmissions
		stats.Summary.AcceptedSubmissions += s.AcceptedSubmissions
		if s.TotalSubmissions > 0 {
			stats.Summary.StudentsWithSubmission++
		}
	}
	if stats.Summary.TotalSubmissions > 0 {
		stats.Summary.CourseAcceptanceRate = float64(stats.Summary.AcceptedSubmissions) /
			float64(stats.Summary.TotalSubmissions) * 100
	}

	httputil.WriteJSON(w, http.StatusOK, stats)
}
