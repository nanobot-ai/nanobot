# ============================================
# Builder Stage - Build agentfs
# ============================================
FROM ubuntu:24.04 AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    build-essential \
    pkg-config \
    libfuse3-dev \
    libunwind-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Rust toolchain
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"

# Clone and build agentfs
WORKDIR /build
RUN git clone --branch main https://github.com/tursodatabase/agentfs && cd agentfs
WORKDIR /build/agentfs/cli
RUN cargo build --release

# ============================================
# Runtime Stage - Clean runtime environment
# ============================================
FROM ubuntu:24.04

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ripgrep \
    git \
    curl \
    wget \
    ca-certificates \
    gnupg \
    fuse3 \
    libfuse3-3 \
    libunwind8 \
    sudo \
    jq \
    xz-utils \
    build-essential \
    python-is-python3 \
    python3 \
    python3-pip \
    python3-venv \
    vim \
    nano \
    less \
    file \
    unzip \
    zip \
    tree \
    htop \
    netcat-openbsd \
    iputils-ping \
    dnsutils \
    && rm -rf /var/lib/apt/lists/*

# Install latest Node.js (using NodeSource)
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - && \
    apt-get install -y nodejs && \
    rm -rf /var/lib/apt/lists/*

# Install uv (fast Python package installer)
RUN curl -LsSf https://astral.sh/uv/install.sh | sh && \
    cd $HOME/.local/bin && \
    mv uv uvx /usr/local/bin

# Create sandbox user with UID 10000
RUN groupadd -g 10000 sandbox && \
    useradd -u 10000 -g 10000 -m -s /bin/bash sandbox && \
    echo "sandbox ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

# Copy agentfs binary from builder
COPY --from=builder /build/agentfs/cli/target/release/agentfs /usr/local/bin/agentfs

# Copy sandbox utility scripts
COPY --chmod=0755 sb_read sb_write sb_execute sb_readdir entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh

# Enable user_allow_other in FUSE config
RUN echo "user_allow_other" >> /etc/fuse.conf

# Create necessary directories for sandbox user
RUN mkdir -p /.data /workspace && \
    chown -R 10000:10000 /.data /workspace

# Set working directory
WORKDIR /workspace

# Switch to sandbox user
USER 10000:10000

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
