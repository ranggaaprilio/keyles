# Multi-arch Dockerfile: builds natively on both linux/amd64 and linux/arm64.
# Builder uses BUILDPLATFORM for fast native compilation; the Go cross-compiler
# targets the requested OS/arch via TARGETOS/TARGETARCH build args.
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

ARG TARGETOS TARGETARCH

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (cross-compile for target platform)
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s" \
    -o /app/bin/server \
    ./cmd/server

# Final stage - minimal runtime image
FROM alpine:3.21

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/server /app/server

# Copy config files if needed
# COPY --from=builder /app/config /app/config

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/server"]
