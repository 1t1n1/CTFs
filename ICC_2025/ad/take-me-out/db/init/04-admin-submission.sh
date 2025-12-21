#!/usr/bin/env bash
set -euo pipefail

psql_user="${POSTGRES_USER:-postgres}"
psql_db="${POSTGRES_DB:-postgres}"
psql_base=(psql -v ON_ERROR_STOP=1 --username "$psql_user" --dbname "$psql_db")

flag_value=$(tr -d '\r' < /flag1)
flag_value="${flag_value//$'\n'/}"

code=$(printf 'print("Yes" if input().strip() == "%s" else "No")' "$flag_value")
code_literal=$(printf "%s" "$code" | sed "s/'/''/g")

"${psql_base[@]}" --set=code_literal="$code_literal" <<'SQL'
WITH admin_row AS (
  SELECT id FROM users WHERE username = 'admin'
),
challenge_row AS (
  SELECT name FROM challenges WHERE name = 'Flag check'
)
INSERT INTO submissions(user_id, challenge, language, code, result, created_at, execution_time_ms, fail_case_index, last_output, expected_output)
SELECT
  admin_row.id,
  challenge_row.name,
  'python',
  :'code_literal',
  'Success',
  NOW(),
  10,
  -1,
  'Yes',
  'Yes'
FROM admin_row, challenge_row
WHERE NOT EXISTS (
  SELECT 1 FROM submissions
  WHERE user_id = admin_row.id AND challenge = challenge_row.name
);
SQL

echo "[db-init] Admin submission seeded."
