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

RUN mkdir -p /tmp/.X11-unix && chmod 1777 /tmp/.X11-unix

WORKDIR /root

# Copy entrypoint script.
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Bind-mounted by DockerRuntime.Create: per-server game install (rw)
VOLUME ["/acc/game"]

# Default server port — overridden at runtime via SERVER_PORT env var.
ENV SERVER_PORT=9600
ENV WINEDEBUG=-all

EXPOSE ${SERVER_PORT}/tcp
EXPOSE ${SERVER_PORT}/udp

ENTRYPOINT ["/entrypoint.sh"]
