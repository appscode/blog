# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder

# Install ca-certificates needed for HTTPS connections during go mod
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy dependency files first so module download layer is cached separately
COPY go.mod go.sum ./

# Download modules with cache mount - reused across builds even if go.mod changes
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build with both module cache and build cache mounts for fast incremental builds
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o mssql-load-test .

# Final lightweight image - distroless has no shell, no package manager, minimal attack surface
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/mssql-load-test .
USER nonroot:nonroot
ENTRYPOINT ["/app/mssql-load-test"]
