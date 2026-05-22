-- Migration 004: Create exercises and test_cases tables
-- UP

CREATE TYPE difficulty_level AS ENUM ('easy', 'medium', 'hard');

CREATE TYPE prog_language AS ENUM ('python', 'go', 'java', 'c', 'cpp', 'javascript');

CREATE TABLE exercises (
    id            UUID             PRIMARY KEY DEFAULT uuid_generate_v4(),
    course_id     UUID             NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    created_by    UUID             NOT NULL REFERENCES users(id)   ON DELETE RESTRICT,
    title         VARCHAR(255)     NOT NULL,
    description   TEXT,
    instructions  TEXT             NOT NULL,
    difficulty    difficulty_level NOT NULL DEFAULT 'medium',
    language      prog_language    NOT NULL,
    template_code TEXT,                          -- optional starter code shown to students
    time_limit_ms INT              NOT NULL DEFAULT 2000,   -- execution time limit
    memory_limit_kb INT            NOT NULL DEFAULT 65536,  -- 64 MB default
    is_published  BOOLEAN          NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exercises_course_id ON exercises(course_id);
CREATE INDEX idx_exercises_created_by ON exercises(created_by);

CREATE TRIGGER trg_exercises_updated_at
    BEFORE UPDATE ON exercises
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Test cases used by the sandbox to evaluate submissions
CREATE TABLE test_cases (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    exercise_id     UUID        NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    input           TEXT        NOT NULL DEFAULT '',
    expected_output TEXT        NOT NULL,
    is_hidden       BOOLEAN     NOT NULL DEFAULT FALSE,  -- hidden = not shown to students
    ordinal         INT         NOT NULL DEFAULT 0,      -- display/run order
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_cases_exercise_id ON test_cases(exercise_id);
