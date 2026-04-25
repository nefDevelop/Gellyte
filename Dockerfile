# --- Stage 1: Build the Go binary ---
FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o gellyte ./cmd/api

# --- Stage 2: Final image ---
FROM alpine:3.19

# Install runtime dependencies: FFmpeg, SSL certificates and VAAPI drivers
RUN apk add --no-cache ffmpeg ca-certificates tzdata libva-intel-driver libva-mesa-driver mesa-va-gallium

# Create a non-root user 'gellyte' with UID 1000
RUN adduser -D -u 1000 gellyte

WORKDIR /app

# Copy the binary from the builder stage and set ownership
COPY --from=builder --chown=gellyte:gellyte /app/gellyte .

# Create media and config directories and set ownership
RUN mkdir -p /app/media/peliculas /app/media/series /app/config && \
    chown -R gellyte:gellyte /app/media /app/config

# Switch to the non-root user
USER gellyte

# Expose the port
EXPOSE 8081

# Command to run the application
CMD ["./gellyte"]
