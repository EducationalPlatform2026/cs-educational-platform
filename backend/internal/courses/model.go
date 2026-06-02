package courses

import "time"

// Course mirrors the courses table plus the caller-specific is_enrolled flag.
type Course struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	CreatedBy   string    `json:"created_by"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsEnrolled  bool      `json:"is_enrolled"` // true if the requesting user is enrolled
}

// Member represents a user enrolled in a course.
type Member struct {
	UserID     string    `json:"user_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Role       string    `json:"role"`
	EnrolledAt time.Time `json:"enrolled_at"`
}

// ── Request types ────────────────────────────────────────────────────────────

type createRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	IsPublished bool    `json:"is_published"`
}

// updateRequest uses pointers so omitted fields are left unchanged (COALESCE).
type updateRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	IsPublished *bool   `json:"is_published"`
}

type enrollRequest struct {
	Role string `json:"role"` // auth.RoleStudent or auth.RoleTeachingAssistant; defaults to RoleStudent
}
