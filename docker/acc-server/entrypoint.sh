#!/bin/bash
set -e

DISPLAY_NUM=$((SERVER_PORT % 1000 + 100))
export DISPLAY=:${DISPLAY_NUM}

rm -f /tmp/.X${DISPLAY_NUM}-lock /tmp/.X11-unix/X${DISPLAY_NUM} 2>/dev/null || true

Xvfb ${DISPLAY} -screen 0 1024x768x16 -nolisten tcp &
XVFB_PID=$!

TCPDUMP_PID=""

cleanup() {
    echo "Stopping ACC server..."
    [ -n "${TCPDUMP_PID}" ] && kill "${TCPDUMP_PID}" 2>/dev/null || true
    kill "${XVFB_PID}" 2>/dev/null || true
}
trap cleanup SIGTERM SIGINT EXIT

# Initialize Wine prefix (no-op if already done)
wineboot -u 2>/dev/null

mkdir -p /acc/game/log

# ─── Debug mode ───────────────────────────────────────────────────────────────
# Set DEBUG_MODE=1 on the container to enable extra diagnostics:
#
#   docker run -e DEBUG_MODE=1 ...
#   or in docker-compose:
#     environment:
#       DEBUG_MODE: "1"
#
# Produces:
#   /acc/game/log/tcpdump_<port>.pcap  — raw packet capture (open in Wireshark)
#   /acc/game/log/wine_debug.log       — Wine debug output
#   Stderr: Wine version, WINEDEBUG channel, prefix state
#
# You can also override which Wine debug channels are enabled:
#   WINEDEBUG_CHANNELS="+loaddll,+seh,+relay"  (relay is very verbose)
# ─────────────────────────────────────────────────────────────────────────────
if [ "${DEBUG_MODE:-0}" = "1" ]; then
    echo "[debug] Debug mode enabled for port ${SERVER_PORT}"
    echo "[debug] Wine: $(wine --version 2>/dev/null || echo unknown)"
    echo "[debug] WINEPREFIX: ${WINEPREFIX}"

    # Configure Wine debug channels (default: loader + exception + module info)
    export WINEDEBUG="${WINEDEBUG_CHANNELS:-+loaddll,+module,+seh,+ntdll}"
    echo "[debug] WINEDEBUG=${WINEDEBUG}"

    # Start packet capture on the server port — captures both TCP and UDP.
    # The pcap file records every byte Kunos sends/receives so you can inspect
    # the 2-byte "pakSizeMax" packet and anything that follows.
    #
    # NOTE: requires CAP_NET_RAW capability on the container:
    #   docker-compose: cap_add: [NET_RAW]
    #   docker run: --cap-add=NET_RAW
    # Without it tcpdump will be skipped (no crash, just no capture).
    CAPTURE_FILE="/acc/game/log/tcpdump_${SERVER_PORT}.pcap"
    echo "[debug] Starting packet capture → ${CAPTURE_FILE}"
    echo "[debug] (requires cap_add: [NET_RAW] in docker-compose — will exit silently if missing)"
    tcpdump -i any -nn -s 0 -w "${CAPTURE_FILE}" \
        "port ${SERVER_PORT}" 2>/acc/game/log/tcpdump.err &
    TCPDUMP_PID=$!

    # Redirect Wine debug output to a log file (Wine writes to stderr).
    WINE_LOG="/acc/game/log/wine_debug.log"
    echo "[debug] Wine debug log → ${WINE_LOG}"
    echo "Starting ACC server on port ${SERVER_PORT} (DEBUG)..."
    cd /acc/game
    exec wine /acc/game/accServer.exe 2>"${WINE_LOG}"
else
    # Production: suppress all Wine debug output.
    export WINEDEBUG=-all
    echo "Starting ACC server on port ${SERVER_PORT}..."
    cd /acc/game
    exec wine /acc/game/accServer.exe
fi
