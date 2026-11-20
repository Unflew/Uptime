#!/usr/bin/env bash
set -Eeuo pipefail

project_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
backend_dir="$project_dir/backend"
frontend_dir="$project_dir/frontend"
backend_pid=""
frontend_pid=""
temp_dir=""

cleanup() {
  local status=$?
  trap - EXIT INT TERM
  for pid in "$frontend_pid" "$backend_pid"; do
    if [[ -n "$pid" ]]; then kill -- "-$pid" 2>/dev/null || true; fi
  done
  for pid in "$frontend_pid" "$backend_pid"; do
    if [[ -n "$pid" ]]; then wait "$pid" 2>/dev/null || true; fi
  done
  if [[ -n "$temp_dir" ]]; then rm -rf -- "$temp_dir"; fi
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

for command in go npm docker setsid; do
  if ! command -v "$command" >/dev/null 2>&1; then
    printf 'Ошибка: команда %s не найдена.\n' "$command" >&2
    exit 1
  fi
done
if ! docker compose version >/dev/null 2>&1; then
  printf 'Ошибка: требуется Docker Compose.\n' >&2
  exit 1
fi
if [[ ! -x "$frontend_dir/node_modules/.bin/next" ]]; then
  printf 'Зависимости frontend не установлены. Выполните: cd frontend && npm ci\n' >&2
  exit 1
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  if [[ "${POSTGRES_USER:-uptime_up}" != uptime_up || "${POSTGRES_PASSWORD:-uptime_up}" != uptime_up || "${POSTGRES_DB:-uptime_up}" != uptime_up ]]; then
    printf 'Для нестандартных параметров PostgreSQL задайте DATABASE_URL.\n' >&2
    exit 1
  fi
  export DATABASE_URL="postgres://uptime_up:uptime_up@127.0.0.1:${POSTGRES_PORT:-5432}/uptime_up?sslmode=disable"
fi
export JWT_SECRET="${JWT_SECRET:-local-development-only-secret-change-before-production}"
export BACKEND_URL="${BACKEND_URL:-http://127.0.0.1:8080}"

printf 'Запуск PostgreSQL...\n'
(cd "$backend_dir" && docker compose up -d --wait)
printf 'Применение миграций...\n'
(cd "$backend_dir" && go run ./cmd/migrate)
printf 'Сборка backend...\n'
temp_dir=$(mktemp -d "${TMPDIR:-/tmp}/uptime-up-dev.XXXXXXXX")
(cd "$backend_dir" && go build -o "$temp_dir/server" ./cmd/server)

(cd "$backend_dir" && exec setsid "$temp_dir/server") &
backend_pid=$!
(cd "$frontend_dir" && exec setsid npm run dev) &
frontend_pid=$!

printf '\nFrontend: http://localhost:3000\nBackend: %s\nДля остановки нажмите Ctrl+C. PostgreSQL останется запущенным.\n\n' "$BACKEND_URL"
# If either server exits, stop its companion and return a failure for unexpected exits.
set +e
wait -n "$backend_pid" "$frontend_pid"
status=$?
set -e
printf '\nОдин из серверов завершил работу; останавливаем второй.\n' >&2
if [[ "$status" -eq 0 ]]; then status=1; fi
exit "$status"
