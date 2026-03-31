# ACC Server container: Wine Staging + accServer.exe
# Game files are bind-mounted from the host at /acc/game.
FROM debian:bookworm-slim

# Install Wine Staging from WineHQ (newer than Debian's packages, includes fsync support).
RUN dpkg --add-architecture i386 && \
    apt-get update && \
    apt-get install -y --no-install-recommends wget ca-certificates && \
    mkdir -pm755 /etc/apt/keyrings && \
    wget -q -O /etc/apt/keyrings/winehq-archive.key https://dl.winehq.org/wine-builds/winehq.key && \
    echo "deb [arch=amd64,i386 signed-by=/etc/apt/keyrings/winehq-archive.key] https://dl.winehq.org/wine-builds/debian/ bookworm main" \
        > /etc/apt/sources.list.d/winehq.list && \
    apt-get update && \
    apt-get install -y --install-recommends \
        winehq-staging \
        xvfb \
        tcpdump \
        && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /tmp/.X11-unix && chmod 1777 /tmp/.X11-unix

WORKDIR /root

ENV WINEARCH=win64
ENV WINEPREFIX=/root/.wine
ENV WINEDEBUG=-all
# fsync uses Linux futex2 for Wine synchronization — reduces timing jitter under load.
# Falls back to esync (eventfd) if fsync is unavailable; falls back to standard if neither.
ENV WINEFSYNC=1
ENV WINEESYNC=1

# Copy entrypoint script.
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Bind-mounted by DockerRuntime.Create: per-server game install (rw)
VOLUME ["/acc/game"]

ENV SERVER_PORT=9600

ENTRYPOINT ["/entrypoint.sh"]
