# 1. UI Build Stage
FROM node:18-alpine AS ui-builder

WORKDIR /app/admin-ui
# Copy package files first for better caching
COPY admin-ui/package.json admin-ui/pnpm-lock.yaml* ./
RUN npm install -g pnpm && pnpm install

# Copy the rest of the UI source code and build it
COPY admin-ui/ ./
RUN pnpm run build

# 2. Go Backend Build Stage
FROM golang:1.22-alpine AS go-builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o gateway ./cmd/gateway

# 3. Final Stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary from go-builder
COPY --from=go-builder /build/gateway .

# Copy config files
COPY configs/gateway.yaml ./configs/

# Copy the built Admin UI from the ui-builder stage
COPY --from=ui-builder /app/admin-ui/dist ./admin-ui/dist

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Run the gateway
CMD ["./gateway"]
