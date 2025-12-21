#!/usr/bin/env bash
set -euo pipefail

echo "[db-entrypoint] Preparing Postgres superuser password..."

postgres_pw="$(tr -dc 'A-Za-z0-9' </dev/urandom | head -c 64 || true)"
if [[ -z "$postgres_pw" ]]; then
  postgres_pw="$(date +%s%N)"
fi
export POSTGRES_PASSWORD="$postgres_pw"

echo "[db-entrypoint] Generated random superuser password."

exec /usr/local/bin/docker-entrypoint.sh "$@"
