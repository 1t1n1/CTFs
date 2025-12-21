-- Schema objects
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  username TEXT UNIQUE NOT NULL,
  password TEXT NOT NULL,
  is_admin BOOLEAN NOT NULL DEFAULT FALSE,
  is_writer BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS challenges (
  id SERIAL UNIQUE NOT NULL,
  name TEXT PRIMARY KEY,
  description TEXT,
  created_by INT REFERENCES users(id) ON DELETE SET NULL,
  input TEXT,
  output TEXT,
  points INT NOT NULL DEFAULT 100,
  is_public BOOLEAN NOT NULL DEFAULT FALSE
);

-- Legacy test_cases migration (split into judge_cases and sample_cases)
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_name = 'legacy_test_cases'
  ) THEN
    EXECUTE 'DROP TABLE legacy_test_cases';
  END IF;
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_name = 'test_cases'
  ) THEN
    EXECUTE 'ALTER TABLE test_cases RENAME TO legacy_test_cases';
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS judge_cases (
  id SERIAL PRIMARY KEY,
  challenge TEXT REFERENCES challenges(name) ON DELETE CASCADE,
  idx INT NOT NULL DEFAULT 0,
  input TEXT,
  output TEXT,
  UNIQUE (challenge, idx)
);

CREATE TABLE IF NOT EXISTS sample_cases (
  id SERIAL PRIMARY KEY,
  challenge TEXT REFERENCES challenges(name) ON DELETE CASCADE,
  idx INT NOT NULL DEFAULT 0,
  input TEXT,
  output TEXT,
  UNIQUE (challenge, idx)
);

-- Helper to clear judge cases without granting delete privileges to the web role
CREATE OR REPLACE FUNCTION purge_judge_cases(challenge_name TEXT)
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
BEGIN
  DELETE FROM judge_cases WHERE challenge = challenge_name;
END;
$$;

REVOKE ALL ON FUNCTION purge_judge_cases(TEXT) FROM PUBLIC;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_name = 'legacy_test_cases'
  ) THEN
    EXECUTE 'INSERT INTO sample_cases(challenge, idx, input, output)
             SELECT challenge, ROW_NUMBER() OVER (PARTITION BY challenge ORDER BY idx) - 1, input, output
             FROM legacy_test_cases WHERE is_sample = TRUE
             ON CONFLICT (challenge, idx) DO NOTHING';
    EXECUTE 'INSERT INTO judge_cases(challenge, idx, input, output)
             SELECT challenge, ROW_NUMBER() OVER (PARTITION BY challenge ORDER BY idx) - 1, input, output
             FROM legacy_test_cases WHERE is_sample = FALSE
             ON CONFLICT (challenge, idx) DO NOTHING';
    EXECUTE 'DROP TABLE legacy_test_cases';
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS submissions (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id),
  challenge TEXT REFERENCES challenges(name),
  language TEXT,
  code TEXT,
  result TEXT,
  created_at TIMESTAMP,
  execution_time_ms INT NOT NULL DEFAULT 0,
  fail_case_index INT NOT NULL DEFAULT -1,
  last_output TEXT,
  expected_output TEXT
);

CREATE TABLE IF NOT EXISTS solves (
  user_id INT REFERENCES users(id),
  challenge TEXT REFERENCES challenges(name),
  created_at TIMESTAMP,
  PRIMARY KEY(user_id, challenge)
);

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'challenges' AND column_name = 'id'
  ) THEN
    BEGIN
      EXECUTE 'ALTER TABLE challenges ALTER COLUMN id SET DEFAULT nextval(''challenges_id_seq'')';
    EXCEPTION WHEN undefined_column THEN
      EXECUTE 'CREATE SEQUENCE IF NOT EXISTS challenges_id_seq OWNED BY challenges.id';
      EXECUTE 'ALTER TABLE challenges ALTER COLUMN id SET DEFAULT nextval(''challenges_id_seq'')';
    END;
    EXECUTE 'UPDATE challenges SET id = nextval(''challenges_id_seq'') WHERE id IS NULL';
    EXECUTE 'ALTER TABLE challenges ALTER COLUMN id SET NOT NULL';
  END IF;
END$$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'challenges' AND column_name = 'statement'
  ) THEN
    IF EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_name = 'challenges' AND column_name = 'description'
    ) THEN
      EXECUTE 'UPDATE challenges SET description = statement WHERE description IS NULL';
      EXECUTE 'ALTER TABLE challenges DROP COLUMN statement';
    ELSE
      EXECUTE 'ALTER TABLE challenges RENAME COLUMN statement TO description';
    END IF;
  END IF;
END$$;

ALTER TABLE challenges
  ADD COLUMN IF NOT EXISTS created_by INT REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE challenges
  ADD COLUMN IF NOT EXISTS is_public BOOLEAN DEFAULT FALSE;

ALTER TABLE challenges
  ALTER COLUMN is_public SET NOT NULL;

UPDATE challenges
  SET is_public = TRUE
  WHERE created_by IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_challenges_id ON challenges(id);

-- Privileges (database name assumed 'postgres' per compose)
GRANT CONNECT ON DATABASE postgres TO "app_web", "app_runner";
GRANT USAGE ON SCHEMA public TO "app_web", "app_runner";

-- Web: insert-only for challenge data, but allow listing and scoreboard
GRANT INSERT, UPDATE ON challenges TO "app_web";
GRANT SELECT, INSERT, UPDATE, DELETE ON sample_cases TO "app_web";
GRANT INSERT ON judge_cases TO "app_web";
GRANT EXECUTE ON FUNCTION purge_judge_cases(TEXT) TO "app_web";
GRANT SELECT (id, name, description, points, created_by, is_public) ON TABLE challenges TO "app_web";

-- Web app needs full access to its own tables
GRANT SELECT, INSERT, UPDATE, DELETE ON users, submissions, solves TO "app_web";
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO "app_web";

-- Runner: needs to seed and refresh built-in challenges
GRANT SELECT, INSERT, UPDATE ON challenges TO "app_runner";
GRANT SELECT, INSERT, UPDATE, DELETE ON sample_cases TO "app_runner";
GRANT SELECT, INSERT, UPDATE, DELETE ON judge_cases TO "app_runner";
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO "app_runner";

-- Indexes for performance
-- submissions listing by user/challenge (FIFO by created_at)
CREATE INDEX IF NOT EXISTS idx_submissions_user_created ON submissions(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_chal_created ON submissions(challenge, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_chal_user_created ON submissions(challenge, user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_result_created ON submissions(result, created_at ASC);

-- sample/judge case access by challenge and index
CREATE INDEX IF NOT EXISTS idx_sample_cases_chal_idx ON sample_cases(challenge, idx);
CREATE INDEX IF NOT EXISTS idx_judge_cases_chal_idx ON judge_cases(challenge, idx);
