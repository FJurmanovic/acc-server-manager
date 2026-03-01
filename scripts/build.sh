#!/usr/bin/env bash
set -euo pipefail

BINARY_NAME="acc-server-manager"
MIGRATE_BINARY_NAME="acc-migrate"
GOOS="windows"
GOARCH="amd64"
OUTPUT_PATH="./build"
SKIP_TESTS=false

usage() {
    cat <<EOF
Usage: $0 [options]

Options:
  --binary-name <name>        Main binary name (default: acc-server-manager)
  --migrate-binary-name <n>   Migration binary name (default: acc-migrate)
  --goos <os>                 Target OS (default: windows)
  --goarch <arch>             Target arch (default: amd64)
  --output-path <path>        Output directory (default: ./build)
  --skip-tests                Skip running tests
  -h, --help                  Show this help
EOF
}

log() { echo "[$(date '+%H:%M:%S')] $*"; }
die() { echo "[$(date '+%H:%M:%S')] ERROR: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
    case $1 in
        --binary-name)          BINARY_NAME="$2";         shift 2 ;;
        --migrate-binary-name)  MIGRATE_BINARY_NAME="$2"; shift 2 ;;
        --goos)                 GOOS="$2";                shift 2 ;;
        --goarch)               GOARCH="$2";              shift 2 ;;
        --output-path)          OUTPUT_PATH="$2";         shift 2 ;;
        --skip-tests)           SKIP_TESTS=true;          shift   ;;
        -h|--help)              usage; exit 0              ;;
        *) die "Unknown option: $1" ;;
    esac
done

command -v go >/dev/null 2>&1 || die "Go is not installed or not in PATH"

# Clean output directory
rm -rf "$OUTPUT_PATH"
mkdir -p "$OUTPUT_PATH"

# Run tests on host before setting cross-compilation environment
if [[ "$SKIP_TESTS" == false ]]; then
    log "Running tests..."
    go test ./...
fi

# CGo cross-compilation setup (set after tests so they run natively)
HOST_OS="$(go env GOHOSTOS)"
if [[ "$GOOS" == "windows" && "$HOST_OS" != "windows" ]]; then
    command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1 \
        || die "mingw-w64 not found. Install it with: sudo apt install gcc-mingw-w64-x86-64"
    export CC=x86_64-w64-mingw32-gcc
    export CGO_ENABLED=1
    log "Cross-compiling for Windows (CC=$CC)"
fi

EXT=""
[[ "$GOOS" == "windows" ]] && EXT=".exe"

export GOOS GOARCH

# Build API binary
log "Building API binary ($GOOS/$GOARCH)..."
go build -o "$OUTPUT_PATH/$BINARY_NAME$EXT" ./cmd/api

# Build migration binary
log "Building migration binary ($GOOS/$GOARCH)..."
go build -o "$OUTPUT_PATH/$MIGRATE_BINARY_NAME$EXT" ./cmd/migrate

unset GOOS GOARCH CC CGO_ENABLED

log "Build completed -> $OUTPUT_PATH"
