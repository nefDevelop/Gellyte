#!/bin/bash

# Gellyte EXTREME Stress Test & Memory Monitor
SERVER_URL="${GELLYTE_URL:-http://localhost:8081}"
CONCURRENCY=${CONCURRENCY:-20}
TIMEOUT_SEC=${TIMEOUT_SEC:-15}

get_mem() {
    echo -n "RAM App (HeapAlloc): "
    curl -s "$SERVER_URL/debug/pprof/heap?debug=1" | grep "# HeapAlloc =" | awk '{print $4/1024/1024 " MB"}'
    echo -n "RAM OS  (Usada/Total): "
    free -h | awk '/^Mem:/ {print $3 " / " $2}'
}

get_goroutines() {
    echo -n "Goroutines: "
    curl -s "$SERVER_URL/debug/pprof/goroutine?debug=1" | grep "goroutine profile: total" | awk '{print $4}'
}

echo "=== Gellyte EXTREME Stress Test Started ==="
echo "Estado inicial:"
get_mem
get_goroutines
echo "-----------------------------------"

echo "Fase 1: Extrayendo IDs de medios (Filtro: ${1:-Ninguno})..."
# Si se pasa un argumento, filtramos por ese texto
# Añadimos IncludeItemTypes=Movie,Episode para garantizar que no elegimos carpetas (Folders/Seasons)
if [ -n "$1" ]; then
    RAW_RESP=$(curl -s "$SERVER_URL/Items?IncludeItemTypes=Movie,Episode&limit=100000")
    ALL_IDS=$(echo "$RAW_RESP" | grep -i "$1" -B 1 | grep -o '"Id":"[^"]*"' | cut -d'"' -f4)
else
    RAW_RESP=$(curl -s "$SERVER_URL/Items?IncludeItemTypes=Movie,Episode&limit=100000")
    ALL_IDS=$(echo "$RAW_RESP" | grep -o '"Id":"[^"]*"' | cut -d'"' -f4)
fi
COUNT=$(echo "$ALL_IDS" | grep -v '^$' | wc -l)

if [ "$COUNT" -eq 0 ]; then
    echo "ERROR: No se han encontrado ítems en el servidor. ¿Está el escáner todavía trabajando?"
    echo "-> Respuesta cruda de la API (Primeros 300 caracteres):"
    echo "$RAW_RESP" | head -c 300
    echo -e "\n..."
    exit 1
fi
echo "Encontrados $COUNT ítems para el test."

echo "-> Escaneando inteligentemente TODO el catálogo para garantizar testeo en 4K, 1080p y SD..."
# Extraemos al instante el ID y el Ancho (Width) de cada video desde la respuesta JSON gigante
ITEM_STATS=$(echo "$RAW_RESP" | sed 's/},{/}\n{/g' | grep '"Width":' | sed -E 's/.*"Id":"([^"]+)".*"Width":([0-9]+).*/\1 \2/')

# Clasificamos usando el Width (Ancho) ya que el Height varía en películas formato cine/ultrawide
HIGH_RES_IDS=$(echo "$ITEM_STATS" | awk '$2 >= 2560 {print $1}' | sort -R | head -n 3 | tr '\n' ' ')
MED_RES_IDS=$(echo "$ITEM_STATS" | awk '$2 >= 1900 && $2 < 2560 {print $1}' | sort -R | head -n 3 | tr '\n' ' ')
LOW_RES_IDS=$(echo "$ITEM_STATS" | awk '$2 < 1900 && $2 > 0 {print $1}' | sort -R | head -n 3 | tr '\n' ' ')

SAMPLE_IDS=$(echo "$HIGH_RES_IDS $MED_RES_IDS $LOW_RES_IDS" | tr ' ' '\n' | grep -v '^$')

printf "   %-15s | %-30s | %-10s | %-10s | %-10s\n" "RESOLUCIÓN" "NOMBRE" "CONTENEDOR" "VIDEO" "AUDIO"
echo "   ----------------|--------------------------------|------------|------------|------------"

for id in $SAMPLE_IDS; do
    INFO=$(curl -s "$SERVER_URL/Items/$id/PlaybackInfo")
    CONTAINER=$(echo "$INFO" | grep -o '"Container":"[^"]*"' | head -1 | cut -d'"' -f4)
    VCODEC=$(echo "$INFO" | grep -o '"Codec":"[^"]*"' | head -1 | cut -d'"' -f4)
    ACODEC=$(echo "$INFO" | grep -o '"Codec":"[^"]*"' | sed -n '2p' | cut -d'"' -f4)
    NAME=$(echo "$INFO" | grep -o '"Name":"[^"]*"' | head -1 | cut -d'"' -f4)
    WIDTH=$(echo "$INFO" | grep -o '"Width": *[0-9]*' | head -1 | sed 's/"Width": *//')

    if [ -z "$WIDTH" ]; then continue; fi

    RES_LABEL=""
    if [ "$WIDTH" -ge 2560 ]; then
        RES_LABEL="Alta (4K+)"
    elif [ "$WIDTH" -ge 1900 ]; then
        RES_LABEL="Media (1080p)"
    else
        RES_LABEL="Baja (<=720p)"
    fi

    # Truncar el nombre a 30 caracteres para no romper la tabla
    SHORT_NAME=$(echo "$NAME" | cut -c 1-30)
    printf "   %-15s | %-30s | %-10s | %-10s | %-10s\n" "$RES_LABEL" "$SHORT_NAME" "$CONTAINER" "$VCODEC" "$ACODEC"
done

echo "Fase 2: Carga masiva de Metadatos ($CONCURRENCY peticiones en paralelo)..."
echo "$ALL_IDS" | xargs -I{} -P "$CONCURRENCY" curl -s -o /dev/null -w "%{http_code}\n" "$SERVER_URL/Items/{}" | grep -c "200" | xargs -I{} echo "-> {} peticiones exitosas (HTTP 200)."
echo "Estado tras Fase 2:"
get_mem
get_goroutines
echo "-----------------------------------"

echo "Fase 3: Simulación de Streaming Masivo (Static Play por $TIMEOUT_SEC segundos)..."
echo "Lanzando descargas estáticas paralelas (Límite: 50 ítems aleatorios)..."
STATIC_IDS=$(echo "$ALL_IDS" | sort -R | head -n 50)
echo "$STATIC_IDS" | xargs -I{} -P "$CONCURRENCY" timeout "$TIMEOUT_SEC" curl -s -o /dev/null "$SERVER_URL/Videos/{}/stream?Static=true"
echo "Estado tras Fase 3 (Streaming detenido):"
get_mem
get_goroutines
echo "-----------------------------------"

echo "Fase 4: Test de Transcodificación Escalonada por Formatos (FFmpeg)..."

run_batch() {
    local label=$1
    local ids=$2
    local params=$3

    if [ -z "$(echo "$ids" | tr -d ' ')" ]; then
        echo "-> Omitiendo tanda: $label (No se encontraron ítems en esta categoría)"
        return
    fi
    
    echo "-> Tanda $label..."
    # Extraemos hasta 3 items de esa categoría y disparamos FFmpeg
    echo "$ids" | tr ' ' '\n' | grep -v '^$' | head -n 3 | xargs -I{} -P 3 timeout "$TIMEOUT_SEC" curl -s -o /dev/null "$SERVER_URL/Videos/{}/stream?$params"
    echo "Estado tras $label:"
    get_mem
    get_goroutines
    echo "-----------------------------------"
}

run_batch "4.1: Archivos de Baja Res -> Forzando a 1Mbps (Transcode duro a h264/aac)" "$LOW_RES_IDS" "maxBitrate=1000000&VideoCodec=h264&AudioCodec=aac"
run_batch "4.2: Archivos de Media Res -> Forzando a 5Mbps (Transcode estándar h264/aac)" "$MED_RES_IDS" "maxBitrate=5000000&VideoCodec=h264&AudioCodec=aac"
run_batch "4.3: Archivos de Alta Res -> Calidad Original (Remux Video Original / Transcode Audio)" "$HIGH_RES_IDS" "VideoCodec=copy&AudioCodec=aac"

echo "Fase 5: Esperando liberación de memoria (65 segundos)..."
sleep 65
echo "Estado Final (Idle de nuevo):"
get_mem
get_goroutines
echo "-----------------------------------"

echo "=== Test EXTREMO Finalizado ==="
