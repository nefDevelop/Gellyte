# --- Stage 1: Build the Go binary ---
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o gellyte ./cmd/api/main.go

# --- Stage 2: Final image ---
FROM alpine:3.19

# Install runtime dependencies: FFmpeg and SSL certificates
RUN apk add --no-cache ffmpeg ca-certificates tzdata

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/gellyte .

# Create media and config directories
RUN mkdir -p /app/media/peliculas /app/media/series /app/config

# Expose the port
EXPOSE 8081

# Command to run the application
CMD ["./gellyte"]
