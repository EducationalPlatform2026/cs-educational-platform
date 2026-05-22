-- Migration 005: Create submissions table
-- UP

CREATE TYPE submission_status AS ENUM (
    'pending',          -- queued, not yet evaluated
    'running',          -- currently executing in sandbox
    'accepted',         -- all test cases passed
    'wrong_answer',     -- output did not match expected
    'runtime_error',    -- program crashed
    'time_limit',       -- exceeded time_limit_ms
    'memory_limit',     -- exceeded memory_limit_kb
    'compile_error'     -- code failed to compile
);

CREATE TABLE submissions (
    id          UUID              PRIMARY KEY DEFAULT uuid_generate_v4(),
    exercise_id UUID              NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    user_id     UUID              NOT NULL REFERENCES users(id)     ON DELETE CASCADE,
    code        TEXT              NOT NULL,
    language    prog_language     NOT NULL,
    status      submission_status NOT NULL DEFAULT 'pending',
    score       NUMERIC(5, 2)     NOT NULL DEFAULT 0,   -- e.g. 75.00 out of 100
    stderr      TEXT,                                    -- compiler / runtime error output
    submitted_at TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_submissions_exercise_id ON submissions(exercise_id);
CREATE INDEX idx_submissions_user_id     ON submissions(user_id);
CREATE INDEX idx_submissions_status      ON submissions(status);

-- Track per-test-case results for a submission
CREATE TABLE submission_results (
    id            UUID              PRIMARY KEY DEFAULT uuid_generate_v4(),
    submission_id UUID              NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    test_case_id  UUID              NOT NULL REFERENCES test_cases(id)  ON DELETE CASCADE,
    status        submission_status NOT NULL,
    actual_output TEXT,
    runtime_ms    INT,
    memory_kb     INT
);

CREATE INDEX idx_submission_results_submission_id ON submission_results(submission_id);
