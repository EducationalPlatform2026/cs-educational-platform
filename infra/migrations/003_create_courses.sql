-- Migration 003: Create courses and enrollments tables
-- UP

CREATE TABLE courses (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    created_by  UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    is_published BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_courses_created_by ON courses(created_by);

CREATE TRIGGER trg_courses_updated_at
    BEFORE UPDATE ON courses
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Junction table: which students/TAs are enrolled in which course
CREATE TYPE enrollment_role AS ENUM ('student', 'teaching_assistant');

CREATE TABLE course_enrollments (
    user_id     UUID             NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    course_id   UUID             NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    role        enrollment_role  NOT NULL DEFAULT 'student',
    enrolled_at TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, course_id)
);

CREATE INDEX idx_enrollments_course_id ON course_enrollments(course_id);
CREATE INDEX idx_enrollments_user_id   ON course_enrollments(user_id);
