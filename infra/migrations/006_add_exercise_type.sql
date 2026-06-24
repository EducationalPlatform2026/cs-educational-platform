-- Add exercise_type enum
DO $$ BEGIN
  CREATE TYPE exercise_type AS ENUM ('coding', 'quiz');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- 'quiz' language value for quiz submissions stored in submissions table
ALTER TYPE prog_language ADD VALUE IF NOT EXISTS 'quiz';

-- Extend exercises table
ALTER TABLE exercises
  ALTER COLUMN language DROP NOT NULL,
  ADD COLUMN IF NOT EXISTS exercise_type exercise_type NOT NULL DEFAULT 'coding',
  ADD COLUMN IF NOT EXISTS quiz_options  JSONB,
  ADD COLUMN IF NOT EXISTS quiz_correct  SMALLINT;
