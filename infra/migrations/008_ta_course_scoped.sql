-- TA is no longer a global account type.
-- Existing teaching_assistant users become students; their course-level
-- enrollment role (course_enrollments.role) is unchanged.
UPDATE users SET role = 'student' WHERE role = 'teaching_assistant';

CREATE TYPE user_role_new AS ENUM ('student', 'professor', 'admin');
ALTER TABLE users
    ALTER COLUMN role TYPE user_role_new
    USING role::text::user_role_new;
DROP TYPE user_role;
ALTER TYPE user_role_new RENAME TO user_role;
