# syntax=docker/dockerfile:1
FROM cgr.dev/chainguard/wolfi-base:latest

# Install ca-certificates and glibc
RUN apk add -U --no-cache ca-certificates glibc

# Copy the binary
COPY nanobot /usr/local/bin/nanobot

# Create non-root user
RUN adduser -D -s /bin/sh nanobot

# Create data directory and set ownership
RUN mkdir -p /data && chown nanobot:nanobot /data

USER nanobot

# Set environment variable for database location
ENV NANOBOT_STATE=/data/nanobot.db

# Define volume for persistent data
VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/nanobot"]
CMD ["serve"]
