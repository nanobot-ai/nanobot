# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 go build -o nanobot .

# Final stage
FROM cgr.dev/chainguard/wolfi-base:latest AS runtime

# Create non-root user
RUN adduser -D -s /bin/sh nanobot

# Create data directory and set ownership
RUN mkdir -p /data && chown nanobot:nanobot /data

USER nanobot

# Set common env vars
ENV NANOBOT_STATE=/data/nanobot.db

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
