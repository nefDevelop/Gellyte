#!/bin/bash

# Gellyte EXTREME Stress Test & Memory Monitor
SERVER_URL="${GELLYTE_URL:-http://localhost:8081}"
CONCURRENCY=${CONCURRENCY:-20}

get_mem() {
    echo -n "RAM en uso (HeapAlloc): "
    curl -s "$SERVER_URL/debug/pprof/heap?debug=1" | grep "# HeapAlloc =" | awk '{print $4/1024/1024 " MB"}'
}

get_goroutines() {
    echo -n "Goroutines: "
    curl -s "$SERVER_URL/debug/pprof/" | grep "goroutine" | head -n 1 | awk '{print $1}'
}

echo "=== Gellyte EXTREME Stress Test Started ==="
echo "Estado inicial:"
get_mem
get_goroutines
echo "-----------------------------------"

echo "Fase 1: Extrayendo IDs de medios (Filtro: ${1:-Ninguno})..."
# Si se pasa un argumento, filtramos por ese texto
if [ -n "$1" ]; then
    ALL_IDS=$(curl -s "$SERVER_URL/Items" | grep -i "$1" -B 1 | grep -o '"Id":"[^"]*"' | cut -d'"' -f4)
else
    ALL_IDS=$(curl -s "$SERVER_URL/Items" | grep -o '"Id":"[^"]*"' | cut -d'"' -f4)
fi
COUNT=$(echo "$ALL_IDS" | grep -v '^$' | wc -l)

if [ "$COUNT" -eq 0 ]; then
    echo "ERROR: No se han encontrado ítems en el servidor. ¿Está el escáner todavía trabajando?"
    exit 1
fi
echo "Encontrados $COUNT ítems para el test."

echo "Fase 2: Carga masiva de Metadatos ($CONCURRENCY peticiones en paralelo)..."
echo "$ALL_IDS" | xargs -I{} -P "$CONCURRENCY" curl -s -o /dev/null -w "%{http_code}\n" "$SERVER_URL/Items/{}" | grep -c "200" | xargs -I{} echo "-> {} peticiones exitosas (HTTP 200)."
echo "Estado tras Fase 2:"
get_mem
get_goroutines
echo "-----------------------------------"

echo "Fase 3: Simulación de Streaming Masivo (Static Play por 10 segundos)..."
echo "Lanzando descargas paralelas con límite de tiempo..."
echo "$ALL_IDS" | xargs -I{} -P "$CONCURRENCY" timeout 10 curl -s -o /dev/null "$SERVER_URL/Videos/{}/stream?Static=true"
echo "Estado tras Fase 3 (Streaming detenido):"
get_mem
get_goroutines
echo "-----------------------------------"

echo "Fase 4: Test de Transcodificación (FFmpeg) en paralelo (10 segundos)..."
# Solo lanzamos para los primeros 3 para no saturar el CPU del servidor
echo "$ALL_IDS" | head -n 3 | xargs -I{} -P 3 timeout 10 curl -s -o /dev/null "$SERVER_URL/Videos/{}/stream"
echo "Estado tras Fase 4 (Transcode detenido):"
get_mem
get_goroutines
echo "-----------------------------------"

echo "Fase 5: Esperando liberación de memoria (65 segundos)..."
sleep 65
echo "Estado Final (Idle de nuevo):"
get_mem
get_goroutines
echo "-----------------------------------"

echo "=== Test EXTREMO Finalizado ==="
