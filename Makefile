.PHONY: all build run dev swagger clean

# Nombre del binario
BINARY_NAME=gellyte

all: build

# Comando para instalar swag si no está instalado, generar docs y ejecutar la API
dev: swagger run

# Levantar la API directamente
run:
	@echo "=> Iniciando el servidor Gellyte (Gin API)..."
	go run ./cmd/api serve

# Generar la documentación de Swagger
swagger:
	@echo "=> Generando documentación Swagger..."
	@if ! command -v swag &> /dev/null; then \
		echo "swag no encontrado, instalando..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	swag init -g cmd/api/main.go -o docs

# Construir el binario (similar a la stage 1 del Dockerfile)
build: swagger
	@echo "=> Compilando el binario..."
	go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/api

# Limpiar archivos generados
clean:
	@echo "=> Limpiando..."
	go clean
	rm -f $(BINARY_NAME)
	rm -rf docs/
