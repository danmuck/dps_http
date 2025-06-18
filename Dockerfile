# syntax=docker/dockerfile:1

### 1. Build the Go binary ###
FROM golang:1.24-alpine AS builder

# Install git (if you pull modules via git) and ca-certificates
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your source code
COPY . .

# Build a statically-linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o server ./cmd/server

### 2. Create a tiny final image ###
FROM scratch

# Copy CA certs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary
COPY --from=builder /app/server /server

# Run as non-root for safety (optional)
USER 65532:65532

ENTRYPOINT ["/server"]