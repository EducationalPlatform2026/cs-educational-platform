-- Convert quiz_correct from SMALLINT (single index) to JSONB (array of indices)
-- Existing single-answer quizzes: e.g. quiz_correct = 1  →  [1]
ALTER TABLE exercises
  ALTER COLUMN quiz_correct TYPE JSONB
  USING CASE
    WHEN quiz_correct IS NULL THEN NULL
    ELSE to_jsonb(ARRAY[quiz_correct::int])
  END;

-- Flag that allows students to select more than one correct answer
ALTER TABLE exercises
  ADD COLUMN IF NOT EXISTS quiz_allow_multiple BOOLEAN NOT NULL DEFAULT false;
