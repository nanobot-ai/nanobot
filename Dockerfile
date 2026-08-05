# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 go build -o nanobot .

# Final stage
FROM cgr.dev/chainguard/wolfi-base:latest AS runtime

# Install bash, git, common utilities, uv, and poppler-utils (provides pdftoppm
# and pdfinfo, used by the built-in read tool to render PDF pages as images)
RUN apk update && apk upgrade --no-cache && apk add --no-cache \
    bash \
    git \
    curl \
    "libnghttp2-14>=1.70.0-r0" \
    wget \
    jq \
    gzip \
    xz \
    coreutils \
    findutils \
    grep \
    sed \
    gawk \
    ripgrep \
    uv \
    poppler-utils

# Install mcp-cli
ARG TARGETARCH
RUN ARCH=$(if [ "${TARGETARCH}" = "amd64" ]; then echo "x64"; else echo "${TARGETARCH}"; fi) && \
    BINARY="mcp-cli-linux-${ARCH}" && \
    wget -O release.json https://api.github.com/repos/obot-platform/mcp-cli/releases/latest && \
    VERSION=$(jq -r '.tag_name' release.json) && \
    CHECKSUMS_SHA256=$(jq -r '.assets[] | select(.name == "checksums.txt") | .digest | sub("^sha256:"; "")' release.json) && \
    BINARY_SHA256=$(jq -r --arg name "${BINARY}" '.assets[] | select(.name == $name) | .digest | sub("^sha256:"; "")' release.json) && \
    test -n "${VERSION}" && test -n "${CHECKSUMS_SHA256}" && test -n "${BINARY_SHA256}" && \
    wget -O checksums.txt "https://github.com/obot-platform/mcp-cli/releases/download/${VERSION}/checksums.txt" && \
    printf '%s  checksums.txt\n' "${CHECKSUMS_SHA256}" | sha256sum -c - && \
    wget "https://github.com/obot-platform/mcp-cli/releases/download/${VERSION}/${BINARY}" && \
    printf '%s  %s\n' "${BINARY_SHA256}" "${BINARY}" | sha256sum -c - && \
    sha256sum --ignore-missing -c checksums.txt && \
    mv "${BINARY}" /usr/bin/mcp-cli && \
    rm release.json checksums.txt && \
    chmod +x /usr/bin/mcp-cli

COPY scripts/prefetch-tiktoken-encodings.sh /tmp/prefetch-tiktoken-encodings.sh
ENV TIKTOKEN_CACHE_DIR=/var/cache/tiktoken
RUN chmod +x /tmp/prefetch-tiktoken-encodings.sh && \
    /tmp/prefetch-tiktoken-encodings.sh && \
    rm /tmp/prefetch-tiktoken-encodings.sh

# Create non-root user with home directory
RUN adduser -D -h /home/nanobot -s /bin/bash nanobot

# Create data and config directories with proper ownership
RUN mkdir -p /data /home/nanobot/.nanobot && \
    chown -R nanobot:nanobot /data /home/nanobot

USER nanobot
WORKDIR /home/nanobot

# Set common env vars
ENV HOME=/home/nanobot
ENV NANOBOT_STATE=/data/nanobot.db
ENV NANOBOT_RUN_LISTEN_ADDRESS=0.0.0.0:8080

EXPOSE 8080

# Define volume for persistent data
VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/nanobot"]
CMD ["run"]

# Release image
FROM runtime AS release

COPY nanobot /usr/local/bin/nanobot

# Dev image
FROM runtime AS dev

# Copy the binary from builder
COPY --from=builder /build/nanobot /usr/local/bin/nanobot
