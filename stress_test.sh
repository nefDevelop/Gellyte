#!/bin/bash

# Gellyte EXTREME Stress Test & Memory Monitor
SERVER_URL="http://192.168.30.30:8081"

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

echo "Fase 1: Extrayendo todos los IDs de medios..."
# Versión sin jq para máxima compatibilidad
ALL_IDS=$(curl -s "$SERVER_URL/Items" | grep -o '"Id":"[^"]*"' | sed 's/"Id":"//;s/"//g')
COUNT=$(echo "$ALL_IDS" | grep -v '^$' | wc -l)

if [ "$COUNT" -eq 0 ]; then
    echo "ERROR: No se han encontrado ítems en el servidor. ¿Está el escáner todavía trabajando?"
    exit 1
fi
echo "Encontrados $COUNT ítems para el test."

echo "Fase 2: Carga masiva de Metadatos (500 peticiones en paralelo)..."
echo "$ALL_IDS" | xargs -I{} -P 20 curl -s -o /dev/null "$SERVER_URL/Items/{}"
echo "Estado tras Fase 2:"
get_mem
get_goroutines
echo "-----------------------------------"

echo "Fase 3: Simulación de Streaming Masivo (Static Play)..."
echo "Lanzando descargas paralelas de todos los ítems..."
for id in $ALL_IDS; do
    curl -s -o /dev/null "$SERVER_URL/Videos/$id/stream?Static=true" &
done
echo "Esperando 10 segundos de descarga activa..."
sleep 10
killall curl 2>/dev/null
echo "Estado tras Fase 3 (Streaming detenido):"
get_mem
get_goroutines
echo "-----------------------------------"

echo "Fase 4: Test de Transcodificación (FFmpeg) en paralelo..."
# Solo lanzamos para los primeros 3 para no saturar el CPU del servidor
echo "$ALL_IDS" | head -n 3 | xargs -I{} -P 3 curl -s -o /dev/null "$SERVER_URL/Videos/{}/stream" &
sleep 10
killall curl 2>/dev/null
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
