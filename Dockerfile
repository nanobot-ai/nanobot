FROM cgr.dev/chainguard/wolfi-base:latest
RUN apk add --no-cache -U iproute2 npm uv nodejs python-3.13 socat && \
    mkdir /mcp && \
    chmod 777 /mcp && \
    sed -i -e '/globalignorefile/d' -e '/python/d' /usr/lib/node_modules/npm/npmrc
COPY --chmod=755 <<"EOF" /usr/bin/proxy
#!/bin/bash
set -e -x

# Default TARGET_HOST to gateway IP if not set
if [ -z "$TARGET_HOST" ]; then
    TARGET_HOST=$(ip route show default | awk '/default/ {print $3}')
fi

# Create directories for certs
mkdir -p /mcp/certs
cd /mcp/certs

# Write certificates from environment variables
cat > ca.crt <<<"$CA_CERT"
cat > client.crt <<<"$CLIENT_CERT"
cat > client.key <<<"$CLIENT_KEY"

# Start socat with TLS client certificate
exec socat \
  TCP-LISTEN:${LISTEN_PORT:-8443},fork,reuseaddr \
  OPENSSL:${TARGET_HOST}:${TARGET_PORT:-8080},cert=/mcp/certs/client.crt,key=/mcp/certs/client.key,cafile=/mcp/certs/ca.crt,verify=0
EOF
ENV HOME=/mcp
WORKDIR /mcp
VOLUME /mcp
