# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o orchestrator-service ./services/orchestrator-service/cmd/orchestrator-service

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/orchestrator-service .

# Run the application
ENTRYPOINT ["/app/orchestrator-service"]
