#!/bin/bash
set -e

# Wine Staging looks for PulseAudio via XDG_RUNTIME_DIR. There is no audio
# in this container, so point it at a valid (empty) directory to silence
# the "XDG_RUNTIME_DIR is invalid" errors.
export XDG_RUNTIME_DIR=/tmp/runtime-${SERVER_PORT}
mkdir -p "${XDG_RUNTIME_DIR}"

# Use Wine's headless (null) display driver — accServer.exe has no GUI so
# there is no need for Xvfb at all. This avoids the X11 client limit and
# removes the Xvfb process entirely.
export DISPLAY=
export WINEDLLOVERRIDES="winex11.drv="

cleanup() {
    echo "Stopping ACC server..."
}
trap cleanup SIGTERM SIGINT

mkdir -p /acc/game/log

echo "Starting ACC server on port ${SERVER_PORT}..."

cd /acc/game
exec wine /acc/game/accServer.exe
