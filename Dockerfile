# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.25-alpine AS builder

# Install Node.js and pnpm for UI build
RUN apk add --no-cache nodejs npm && npm install -g pnpm

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build UI and binary
RUN CI=true CGO_ENABLED=0 go generate ./... && go build -o nanobot .

# Final stage
FROM ubuntu:24.04 AS runtime

# Install bash, git, common utilities
RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    git \
    curl \
    wget \
    jq \
    gzip \
    xz-utils \
    coreutils \
    findutils \
    grep \
    sed \
    gawk \
    ripgrep \
    ca-certificates \
    xvfb \
    x11vnc \
    fluxbox \
    websockify \
    && rm -rf /var/lib/apt/lists/*

# Install Playwright and browsers as root to shared location
ENV PLAYWRIGHT_BROWSERS_PATH=/usr/local/share/ms-playwright
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 \
    python3-venv \
    && python3 -m venv /opt/playwright-venv \
    && /opt/playwright-venv/bin/pip install playwright \
    && /opt/playwright-venv/bin/playwright install chromium --with-deps \
    && rm -rf /var/lib/apt/lists/*

# Install noVNC for web-based VNC access
RUN git clone --depth 1 https://github.com/novnc/noVNC.git /usr/share/novnc && \
    ln -s /usr/share/novnc/vnc.html /usr/share/novnc/index.html

# Copy VNC startup scripts
COPY scripts/start-vnc.sh /usr/local/bin/start-vnc.sh
COPY scripts/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/start-vnc.sh /usr/local/bin/docker-entrypoint.sh

# Set display for browser-use to use
ENV DISPLAY=:99

# Create non-root user with home directory
RUN useradd -m -d /home/nanobot -s /bin/bash nanobot

# Create data and config directories with proper ownership
RUN mkdir -p /data /home/nanobot/.nanobot && \
    chown -R nanobot:nanobot /data /home/nanobot

USER nanobot
WORKDIR /home/nanobot

# Set common env vars
ENV HOME=/home/nanobot

# Install uv and browser-use as nanobot user
RUN curl -LsSf https://astral.sh/uv/install.sh | sh && \
    /home/nanobot/.local/bin/uv tool install browser-use

# Add uv tools to PATH
ENV PATH="/home/nanobot/.local/bin:$PATH"
ENV NANOBOT_STATE=/data/nanobot.db
ENV NANOBOT_RUN_LISTEN_ADDRESS=0.0.0.0:8080

EXPOSE 8080
EXPOSE 6080

# Define volume for persistent data
VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["run"]

# Release image
FROM runtime AS release

COPY nanobot /usr/local/bin/nanobot

# Dev image
FROM runtime AS dev

# Copy the binary from builder
COPY --from=builder /build/nanobot /usr/local/bin/nanobot
