# ACC Server container: Wine + accServer.exe
# Game files are mounted read-only from the shared acc-game volume.
# Per-server config files are mounted read-write from the manager's config volume.
FROM debian:bookworm-slim

# Install Wine, Xvfb (headless display for Wine), and minimal runtime deps.
RUN dpkg --add-architecture i386 && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
        wine \
        wine32 \
        wine64 \
        xvfb \
        ca-certificates \
        && \
    rm -rf /var/lib/apt/lists/*

# Create a non-root user for running the server.
RUN useradd -ms /bin/bash accserver

WORKDIR /home/accserver

# Copy entrypoint script.
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Volumes (bind-mounted by DockerRuntime.Create):
#   /acc/game  — shared game files, mounted read-only (ACC_GAME_VOLUME)
#   /acc/cfg   — per-server config dir, mounted read-write (server.GetConfigPath())
VOLUME ["/acc/game", "/acc/cfg"]

# Default server port — overridden at runtime via SERVER_PORT env var.
ENV SERVER_PORT=9600
ENV WINEDEBUG=-all
ENV DISPLAY=:99

EXPOSE ${SERVER_PORT}/tcp
EXPOSE ${SERVER_PORT}/udp

USER accserver
ENTRYPOINT ["/entrypoint.sh"]
