package exercises

import "time"

// Exercise mirrors the exercises table.
type Exercise struct {
	ID            string    `json:"id"`
	CourseID      string    `json:"course_id"`
	CreatedBy     string    `json:"created_by"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	Instructions  string    `json:"instructions"`
	Difficulty    string    `json:"difficulty"`
	Language      *string   `json:"language,omitempty"`
	TemplateCode  *string   `json:"template_code,omitempty"`
	TimeLimitMs   int       `json:"time_limit_ms"`
	MemoryLimitKb int       `json:"memory_limit_kb"`
	IsPublished       bool      `json:"is_published"`
	ExerciseType      string    `json:"exercise_type"`
	QuizOptions       []string  `json:"quiz_options,omitempty"`
	QuizCorrect       []int     `json:"quiz_correct,omitempty"`
	QuizAllowMultiple bool      `json:"quiz_allow_multiple"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TestCase mirrors the test_cases table.
// Input and ExpectedOutput are omitted for students when is_hidden = true
// (the handler strips them before responding).
type TestCase struct {
	ID             string    `json:"id"`
	ExerciseID     string    `json:"exercise_id"`
	Input          string    `json:"input"`
	ExpectedOutput string    `json:"expected_output"`
	IsHidden       bool      `json:"is_hidden"`
	Ordinal        int       `json:"ordinal"`
	CreatedAt      time.Time `json:"created_at"`
}

// HiddenTestCase is returned to students in place of TestCase when is_hidden = true.
type HiddenTestCase struct {
	ID         string    `json:"id"`
	ExerciseID string    `json:"exercise_id"`
	IsHidden   bool      `json:"is_hidden"`
	Ordinal    int       `json:"ordinal"`
	CreatedAt  time.Time `json:"created_at"`
}

// ── Request types ────────────────────────────────────────────────────────────

type createExerciseRequest struct {
	Title         string   `json:"title"`
	Description   *string  `json:"description"`
	Instructions  string   `json:"instructions"`
	Difficulty    string   `json:"difficulty"`
	ExerciseType  string   `json:"exercise_type"`
	Language      string   `json:"language"`
	TemplateCode  *string  `json:"template_code"`
	TimeLimitMs   *int     `json:"time_limit_ms"`
	MemoryLimitKb *int     `json:"memory_limit_kb"`
	IsPublished       bool     `json:"is_published"`
	QuizOptions       []string `json:"quiz_options"`
	QuizCorrect       []int    `json:"quiz_correct"`
	QuizAllowMultiple bool     `json:"quiz_allow_multiple"`
}

// updateExerciseRequest uses pointers so omitted fields are left unchanged (COALESCE).
type updateExerciseRequest struct {
	Title         *string  `json:"title"`
	Description   *string  `json:"description"`
	Instructions  *string  `json:"instructions"`
	Difficulty    *string  `json:"difficulty"`
	ExerciseType  *string  `json:"exercise_type"`
	Language      *string  `json:"language"`
	TemplateCode  *string  `json:"template_code"`
	TimeLimitMs   *int     `json:"time_limit_ms"`
	MemoryLimitKb *int     `json:"memory_limit_kb"`
	IsPublished       *bool    `json:"is_published"`
	QuizOptions       []string `json:"quiz_options"`
	QuizCorrect       []int    `json:"quiz_correct"`
	QuizAllowMultiple *bool    `json:"quiz_allow_multiple"`
}

type createTestCaseRequest struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	IsHidden       bool   `json:"is_hidden"`
	Ordinal        int    `json:"ordinal"`
}

// updateTestCaseRequest uses pointers so omitted fields are left unchanged.
type updateTestCaseRequest struct {
	Input          *string `json:"input"`
	ExpectedOutput *string `json:"expected_output"`
	IsHidden       *bool   `json:"is_hidden"`
	Ordinal        *int    `json:"ordinal"`
}
