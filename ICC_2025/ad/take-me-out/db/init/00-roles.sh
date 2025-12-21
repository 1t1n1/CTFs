#!/usr/bin/env bash
set -euo pipefail

echo "[db-init] Creating roles..."

if [[ ! -r /flag1 ]]; then
  echo "[db-init] /flag1 not readable; cannot derive web role password." >&2
  exit 1
fi
if [[ ! -r /flag2 ]]; then
  echo "[db-init] /flag2 not readable; cannot derive runner role password." >&2
  exit 1
fi

web_pass="$(sha256sum /flag1 | awk '{print $1}')"
runner_pass="$(sha256sum /flag2 | awk '{print $1}')"

psql_user="${POSTGRES_USER:-postgres}"
psql_db="${POSTGRES_DB:-postgres}"
psql_base=(psql -v ON_ERROR_STOP=1 --username "$psql_user" --dbname "$psql_db")

create_or_update_role() {
  local role_name="$1"
  local role_pass="$2"
  local escaped_pass="${role_pass//\'/\'\'}"
  if ! "${psql_base[@]}" -c "CREATE ROLE \"${role_name}\" LOGIN PASSWORD '${escaped_pass}'"; then
    "${psql_base[@]}" -c "ALTER ROLE \"${role_name}\" WITH PASSWORD '${escaped_pass}' LOGIN"
  fi
}

create_or_update_role "$WEB_DB_USER" "$web_pass"
create_or_update_role "$RUNNER_DB_USER" "$runner_pass"

echo "[db-init] Roles created."
