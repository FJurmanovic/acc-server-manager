#!/usr/bin/env bash
# migrate-wine.sh — Pull the latest acc-wine image and migrate all server containers.
#
# Usage:
#   ./scripts/migrate-wine.sh -u <username> -p <password> -b <base_url> [--stop-running]
#
# Options:
#   -u, --username      API username
#   -p, --password      API password
#   -b, --base-url      Base URL of the manager (e.g. http://192.168.0.177:3000)
#       --stop-running  Also migrate currently running servers (stop → migrate → restart)
#
# Examples:
#   ./scripts/migrate-wine.sh -u admin -p secret -b http://192.168.0.177:3000
#   ./scripts/migrate-wine.sh -u admin -p secret -b http://192.168.0.177:3000 --stop-running

set -euo pipefail

# ── Argument parsing ────────────────────────────────────────────────────────────
USERNAME=""
PASSWORD=""
BASE_URL=""
STOP_RUNNING=false

usage() {
    grep '^#' "$0" | sed 's/^# \{0,1\}//' | tail -n +2
    exit 1
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        -u|--username)    USERNAME="$2";  shift 2 ;;
        -p|--password)    PASSWORD="$2";  shift 2 ;;
        -b|--base-url)    BASE_URL="$2";  shift 2 ;;
        --stop-running)   STOP_RUNNING=true; shift ;;
        -h|--help)        usage ;;
        *) echo "Unknown option: $1"; usage ;;
    esac
done

[[ -z "$USERNAME" ]] && { echo "Error: --username is required"; usage; }
[[ -z "$PASSWORD" ]] && { echo "Error: --password is required"; usage; }
[[ -z "$BASE_URL" ]] && { echo "Error: --base-url is required"; usage; }

BASE_URL="${BASE_URL%/}"              # strip trailing slash
BASE_URL="${BASE_URL/#http:\/\//https://}"  # upgrade http → https

# ── Check dependencies ──────────────────────────────────────────────────────────
for cmd in curl jq; do
    command -v "$cmd" &>/dev/null || { echo "Error: '$cmd' is required but not installed."; exit 1; }
done

# ── Login ───────────────────────────────────────────────────────────────────────
echo "Logging in as '$USERNAME'..."

LOGIN_RESPONSE=$(curl -sL --post301 --post302 -X POST "${BASE_URL}/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "$(jq -n --arg u "$USERNAME" --arg p "$PASSWORD" '{username: $u, password: $p}')")

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // empty' 2>/dev/null || true)
if [[ -z "$TOKEN" ]]; then
    echo "Error: Login failed — no token in response."
    exit 1
fi
echo "Login successful."

# ── Migrate ─────────────────────────────────────────────────────────────────────
echo "Starting image migration (stopRunning=${STOP_RUNNING})..."

MIGRATE_RESPONSE=$(curl -sL --post301 --post302 -X POST "${BASE_URL}/v1/system/migrate-image" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}" \
    -d "{\"stopRunning\":${STOP_RUNNING}}")

# ── Results ──────────────────────────────────────────────────────────────────────
SERVER_ERROR=$(echo "$MIGRATE_RESPONSE" | jq -r '.error // empty' 2>/dev/null || true)
if [[ -n "$SERVER_ERROR" ]]; then
    echo "Error: Migration failed — $SERVER_ERROR"
    echo "Full response: $MIGRATE_RESPONSE"
    exit 1
fi

IMAGE=$(echo "$MIGRATE_RESPONSE"    | jq -r '.image')
MIGRATED=$(echo "$MIGRATE_RESPONSE" | jq -r '.migrated | length')
SKIPPED=$(echo "$MIGRATE_RESPONSE"  | jq -r '.skipped  | length')
FAILED=$(echo "$MIGRATE_RESPONSE"   | jq -r '.failed   | length')

echo ""
echo "Image: ${IMAGE}"
echo "────────────────────────────────"
printf "  Migrated : %s\n" "$MIGRATED"
printf "  Skipped  : %s\n" "$SKIPPED"
printf "  Failed   : %s\n" "$FAILED"
echo "────────────────────────────────"

if [[ "$MIGRATED" -gt 0 ]]; then
    echo ""
    echo "Migrated servers:"
    echo "$MIGRATE_RESPONSE" | jq -r '.migrated[] | "  - \(.name) (wasRunning: \(.wasRunning))"'
fi

if [[ "$SKIPPED" -gt 0 ]]; then
    echo ""
    echo "Skipped (running, use --stop-running to include):"
    echo "$MIGRATE_RESPONSE" | jq -r '.skipped[] | "  - \(.name)"'
fi

if [[ "$FAILED" -gt 0 ]]; then
    echo ""
    echo "Failed:"
    echo "$MIGRATE_RESPONSE" | jq -r '.failed[] | "  - \(.name): \(.error)"'
    exit 1
fi

echo ""
echo "Done."
