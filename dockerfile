# Build
FROM golang:1.23-alpine3.20 AS builder
RUN apk update && \
    apk add build-base ca-certificates tzdata && \
    adduser -D -g '' appuser
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download
# Copy source code (use .dockerignore to exclude sensitive files)
COPY . .

# Build with security flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o orchestratorm8 .


# Release
FROM alpine:3.20

# Security updates and minimal runtime dependencies
RUN apk update && \
    apk --no-cache add ca-certificates tzdata && \
    apk --no-cache upgrade && \
    rm -rf /var/cache/apk/* && \
    adduser -D -g '' -s /bin/sh appuser

COPY --from=builder --chown=appuser:appuser /app/orchestratorm8 /usr/local/bin/

# Security: Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD orchestratorm8 --help || exit 1

# Metadata
LABEL maintainer="i@deifzar.me" \
    version="1.0" \
    description="ASMM8 - Hardened Orchestrator Scanner (Runtime)" \
    security.scan="required-non-root-privileges"

CMD ["orchestratorm8","--help"]