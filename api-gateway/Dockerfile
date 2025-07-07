FROM rust:1.85-bookworm AS build

# Set environment variables for cargo
ENV CARGO_REGISTRIES_CRATES_IO_PROTOCOL=sparse

# Install build dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

# Create app directory
WORKDIR /app

# Copy dependency files first for better caching
COPY Cargo.toml /app/
COPY build.rs /app/
COPY proto /app/proto
COPY migrations /app/migrations

# Create a dummy main to build dependencies
RUN mkdir -p src/bin && echo 'fn main() {}' > src/main.rs && echo 'fn main() {}' > src/bin/seeder.rs

# Build dependencies only (this will be cached unless Cargo.toml changes)
RUN cargo build --release
RUN rm -f target/release/deps/api_gateway* target/release/deps/server* target/release/deps/seeder*

# Now copy the actual source code
COPY src /app/src

# Build the actual application
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libssl3 \
    && rm -rf /var/lib/apt/lists/*

# Copy the binary
COPY --from=build /app/target/release/server /app/server

# Create non-root user for security
RUN groupadd -r appuser && useradd -r -g appuser appuser
RUN chown appuser:appuser /app/server
USER appuser

EXPOSE 4000

CMD ["/app/server"]