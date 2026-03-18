# Multi-stage build for the acc-server-manager Go backend.
# CGO_ENABLED=1 is required for mattn/go-sqlite3 (uses cgo).
# The binary is compiled for Linux (Docker mode default).

# ─── Builder ──────────────────────────────────────────────────────────────────
FROM golang:1.25-bookworm AS builder

WORKDIR /build

# Install build dependencies for CGO + sqlite3.
RUN apt-get update && apt-get install -y --no-install-recommends \
        gcc \
        libc6-dev \
        && rm -rf /var/lib/apt/lists/*

# Download Go modules first (layer cache).
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build.
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /acc-server-manager ./cmd/api

# ─── Runtime ──────────────────────────────────────────────────────────────────
FROM debian:bookworm-slim AS runtime

# Install runtime dependencies and SteamCMD.
# steamcmd is not in the standard Debian repo — download directly from Valve.
RUN dpkg --add-architecture i386 && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        lib32gcc-s1 \
        && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /usr/games && \
    curl -sSL --retry 3 --retry-delay 5 \
        -o /tmp/steamcmd.tar.gz \
        "https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz" && \
    tar -xzf /tmp/steamcmd.tar.gz -C /usr/games && \
    rm /tmp/steamcmd.tar.gz && \
    chmod +x /usr/games/steamcmd.sh && \
    ln -sf /usr/games/steamcmd.sh /usr/games/steamcmd

WORKDIR /app

COPY --from=builder /acc-server-manager /app/acc-server-manager

# Persistent data directories.
RUN mkdir -p /data/servers /data/db

ENV PLATFORM=docker
ENV PORT=3000
ENV DB_NAME=/data/db/acc.db
ENV STEAMCMD_LINUX_PATH=/usr/games/steamcmd
ENV ACC_SERVERS_PATH=/data/servers

EXPOSE 3000

ENTRYPOINT ["/app/acc-server-manager"]
