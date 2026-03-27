# ACC Server container: Wine + accServer.exe
# Game files are bind-mounted from the host at /acc/game.
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
        tcpdump \
        && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /tmp/.X11-unix && chmod 1777 /tmp/.X11-unix

WORKDIR /root

ENV WINEARCH=win64
ENV WINEPREFIX=/root/.wine
ENV WINEDEBUG=-all

# Copy entrypoint script.
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Bind-mounted by DockerRuntime.Create: per-server game install (rw)
VOLUME ["/acc/game"]

ENV SERVER_PORT=9600

ENTRYPOINT ["/entrypoint.sh"]
