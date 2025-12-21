#!/usr/bin/env bash
set -euo pipefail

echo "[db-init] Seeding admin user..."

psql_user="${POSTGRES_USER:-postgres}"
psql_db="${POSTGRES_DB:-postgres}"
psql_base=(psql -v ON_ERROR_STOP=1 --username "$psql_user" --dbname "$psql_db")

if [[ ! -r /flag1 || ! -r /flag2 ]]; then
  echo "[db-init] Required flag files are not readable; cannot seed admin password." >&2
  exit 1
fi

admin_password="$(cat /flag1 /flag2 | sha256sum | awk '{print $1}')"

"${psql_base[@]}" --set=admin_pass="$admin_password" <<'SQL'
INSERT INTO users(username, password, is_admin, is_writer)
VALUES ('admin', :'admin_pass', TRUE, TRUE)
ON CONFLICT (username) DO UPDATE
SET password = EXCLUDED.password,
    is_admin = TRUE,
    is_writer = TRUE;
SQL

echo "[db-init] Admin user seeded."
