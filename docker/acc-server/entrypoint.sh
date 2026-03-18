#!/bin/bash
set -e

Xvfb :99 -screen 0 1024x768x16 &
XVFB_PID=$!

cleanup() {
    echo "Stopping ACC server..."
    kill $XVFB_PID 2>/dev/null || true
}
trap cleanup SIGTERM SIGINT

echo "Starting ACC server on port ${SERVER_PORT}..."

exec wine /acc/game/server/accServer.exe \
    -serverCfgPath /acc/cfg \
    -port "${SERVER_PORT}"
