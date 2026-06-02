package dashboard

import (
	"net/http"
	"time"

	"cs-educational-platform/backend/internal/auth"
	"cs-educational-platform/backend/internal/db"
	"cs-educational-platform/backend/internal/httputil"
)

// Stats is the full dashboard payload for the requesting user.
type Stats struct {
	EnrolledCourses    int               `json:"enrolled_courses"`
	ExercisesAttempted int               `json:"exercises_attempted"`
	TotalSubmissions   int               `json:"total_submissions"`
	AcceptedSubmissions int              `json:"accepted_submissions"`
	AcceptanceRate     float64           `json:"acceptance_rate"` // 0–100
	RecentSubmissions  []RecentSubmission `json:"recent_submissions"`
}

// RecentSubmission is a slim submission row enriched with exercise title.
type RecentSubmission struct {
	ID            string    `json:"id"`
	ExerciseID    string    `json:"exercise_id"`
	ExerciseTitle string    `json:"exercise_title"`
	Language      string    `json:"language"`
	Status        string    `json:"status"`
	Score         float64   `json:"score"`
	SubmittedAt   time.Time `json:"submitted_at"`
}

// Handler GET /dashboard  [any authenticated user]
// Returns aggregated statistics for the requesting user.
func Handler(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromCtx(r.Context())
	ctx := r.Context()

	var stats Stats

	// ── Enrolled courses ──────────────────────────────────────────────────────
	if err := db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM course_enrollments WHERE user_id = $1`, userID,
	).Scan(&stats.EnrolledCourses); err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// ── Submission aggregates ─────────────────────────────────────────────────
	if err := db.Pool.QueryRow(ctx,
		`SELECT
		    COUNT(*)                                              AS total,
		    COUNT(*) FILTER (WHERE status = 'accepted')          AS accepted,
		    COUNT(DISTINCT exercise_id)                          AS exercises_attempted
		 FROM submissions
		 WHERE user_id = $1`, userID,
	).Scan(&stats.TotalSubmissions, &stats.AcceptedSubmissions, &stats.ExercisesAttempted); err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if stats.TotalSubmissions > 0 {
		stats.AcceptanceRate = float64(stats.AcceptedSubmissions) / float64(stats.TotalSubmissions) * 100
	}

	// ── Recent submissions (last 10, with exercise title) ─────────────────────
	rows, err := db.Pool.Query(ctx,
		`SELECT s.id, s.exercise_id, e.title, s.language, s.status, s.score, s.submitted_at
		 FROM   submissions s
		 JOIN   exercises   e ON e.id = s.exercise_id
		 WHERE  s.user_id = $1
		 ORDER  BY s.submitted_at DESC
		 LIMIT  10`,
		userID,
	)
	if err != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	stats.RecentSubmissions = make([]RecentSubmission, 0)
	for rows.Next() {
		var rs RecentSubmission
		if err := rows.Scan(&rs.ID, &rs.ExerciseID, &rs.ExerciseTitle,
			&rs.Language, &rs.Status, &rs.Score, &rs.SubmittedAt); err != nil {
			httputil.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		stats.RecentSubmissions = append(stats.RecentSubmissions, rs)
	}
	if rows.Err() != nil {
		httputil.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, stats)
}
