#!/bin/bash
set -e

# Use a display number derived from the server port to avoid conflicts
# between containers (each server gets its own unique display).
DISPLAY_NUM=$((SERVER_PORT % 1000 + 100))
export DISPLAY=:${DISPLAY_NUM}

# Clean up any stale X lock from a previous container run.
rm -f /tmp/.X${DISPLAY_NUM}-lock /tmp/.X11-unix/X${DISPLAY_NUM} 2>/dev/null || true

Xvfb ${DISPLAY} -screen 0 1024x768x16 -nolisten tcp &
XVFB_PID=$!

cleanup() {
    echo "Stopping ACC server..."
    kill $XVFB_PID 2>/dev/null || true
}
trap cleanup SIGTERM SIGINT

mkdir -p /acc/game/log

echo "Starting ACC server on port ${SERVER_PORT}..."

cd /acc/game
exec wine /acc/game/accServer.exe
