package auth

// Global user_role constants (stored on users.role).
// teaching_assistant is NOT a global role — it is an enrollment-only role
// stored in course_enrollments.role.
const (
	RoleStudent   = "student"
	RoleProfessor = "professor"
	RoleAdmin     = "admin"

	// RoleEnrollmentTA is used only when reading/writing course_enrollments.role.
	RoleEnrollmentTA = "teaching_assistant"
)
