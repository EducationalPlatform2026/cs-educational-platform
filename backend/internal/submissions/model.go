package submissions

import "time"

// Submission mirrors the submissions table.
type Submission struct {
	ID          string    `json:"id"`
	ExerciseID  string    `json:"exercise_id"`
	UserID      string    `json:"user_id"`
	Code        string    `json:"code"`
	Language    string    `json:"language"`
	Status      string    `json:"status"`
	Score       float64   `json:"score"`
	Stderr      *string   `json:"stderr,omitempty"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// SubmissionResult mirrors the submission_results table.
type SubmissionResult struct {
	ID           string  `json:"id"`
	SubmissionID string  `json:"submission_id"`
	TestCaseID   string  `json:"test_case_id"`
	Status       string  `json:"status"`
	ActualOutput *string `json:"actual_output,omitempty"`
	RuntimeMs    *int    `json:"runtime_ms,omitempty"`
	MemoryKb     *int    `json:"memory_kb,omitempty"`
}

// SubmissionDetail is returned by GetHandler — submission + its per-test-case results.
type SubmissionDetail struct {
	Submission
	Results []SubmissionResult `json:"results"`
}

// ── Request types ────────────────────────────────────────────────────────────

type submitRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}
