# Etapa 1: Compilación
FROM golang:1.22-alpine AS builder

# Instalar dependencias de compilación para SQLite
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copiar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar con flags -s -w (Stripped binary)
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o gellyte cmd/api/main.go

# Etapa 2: Imagen Final
FROM alpine:latest

# Instalar dependencias necesarias (FFmpeg para Fase 2)
RUN apk add --no-cache ffmpeg ca-certificates tzdata

WORKDIR /root/

# Copiar binario y docs
COPY --from=builder /app/gellyte .
COPY --from=builder /app/docs ./docs

# Crear estructura de carpetas
RUN mkdir -p media/peliculas

EXPOSE 8080

# Ejecutar
CMD ["./gellyte"]
