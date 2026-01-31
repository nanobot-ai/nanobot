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
FROM cgr.dev/chainguard/wolfi-base:latest AS runtime

# Install bash
RUN apk add --no-cache bash

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
