#!/usr/bin/env bash
set -euo pipefail

echo "[db-init] Seeding Flag check challenge..."

if [[ ! -r /flag1 ]]; then
  echo "[db-init] /flag1 not readable; cannot seed challenge." >&2
  exit 1
fi

psql_user="${POSTGRES_USER:-postgres}"
psql_db="${POSTGRES_DB:-postgres}"
psql_base=(psql -v ON_ERROR_STOP=1 --username "$psql_user" --dbname "$psql_db")

flag_value=$(tr -d '\r' < /flag1)
flag_value="${flag_value//$'\n'/}"
dummy_input="ICC{dummy}"

description='Read the flag from stdin and print "Yes" if it matches the real flag, otherwise print "No".'

"${psql_base[@]}" --set=flag_value="$flag_value" --set=dummy_input="$dummy_input" --set=description="$description" <<'SQL'
WITH admin_user AS (
  SELECT id FROM users WHERE username = 'admin'
)
INSERT INTO challenges(name, description, created_by, input, output, points, is_public)
VALUES (
  'Flag check',
  :'description',
  (SELECT id FROM admin_user),
  '',
  '',
  100,
  TRUE
)
ON CONFLICT (name) DO UPDATE
  SET description = EXCLUDED.description,
      created_by = COALESCE(EXCLUDED.created_by, challenges.created_by),
      is_public = TRUE;

delete from sample_cases where challenge = 'Flag check';
INSERT INTO sample_cases(challenge, idx, input, output)
VALUES ('Flag check', 0, :'dummy_input', 'No')
ON CONFLICT (challenge, idx) DO UPDATE
  SET input = EXCLUDED.input,
      output = EXCLUDED.output;

SELECT purge_judge_cases('Flag check');

INSERT INTO judge_cases(challenge, idx, input, output) VALUES
  ('Flag check', 0, :'flag_value', 'Yes'),
  ('Flag check', 1, 'totally_wrong_flag', 'No')
ON CONFLICT (challenge, idx) DO UPDATE
  SET input = EXCLUDED.input,
      output = EXCLUDED.output;
SQL

echo "[db-init] Flag check challenge seeded."
