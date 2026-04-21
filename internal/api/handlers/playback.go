package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/library"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/transcoder"
	"github.com/gin-gonic/gin"
)

// GetPlaybackInfo godoc
// @Summary Obtener información de reproducción
// @Description Determina si el archivo se puede reproducir directamente.
// @Tags Playback
// @Produce json
// @Param id path string true "ID del item"
// @Success 200 {object} map[string]interface{}
// @Router /Items/{id}/PlaybackInfo [get]
// GetPlaybackInfo godoc
func GetPlaybackInfo(c *gin.Context) {
	id := c.Param("id")

	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item no encontrado"})
		return
	}

	// Por ahora simulamos que todo es Direct Play para máxima fluidez en el MVP extendido.
	// Sin embargo, preparamos la estructura para que apps como Swiftfin no den error.

	streamUrl := fmt.Sprintf("/Videos/%s/stream", item.ID)

	// Extraer pistas de audio y subtítulos al vuelo para poblar la UI de Jellyfin
	mediaStreams := []gin.H{}
	rawMeta, err := library.GetRawMetadata(item.Path)
	if err == nil && rawMeta != nil {
		for _, s := range rawMeta.Streams {
			lang := ""
			if s.Tags != nil && s.Tags["language"] != "" {
				lang = s.Tags["language"]
			}

			streamType := "Video"
			if s.CodecType == "audio" {
				streamType = "Audio"
			} else if s.CodecType == "subtitle" {
				streamType = "Subtitle"
			}

			isDefault := false
			if s.Disposition != nil && s.Disposition["default"] == 1 {
				isDefault = true
			}

			mediaStreams = append(mediaStreams, gin.H{
				"Type":         streamType,
				"Index":        s.Index,
				"Codec":        s.CodecName,
				"Language":     lang,
				"IsDefault":    isDefault,
				"IsInterlaced": false,
				"Width":        s.Width,
				"Height":       s.Height,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"MediaSources": []gin.H{
			{
				"Protocol":             "Http",
				"Id":                   item.ID,
				"Path":                 item.Path,
				"Type":                 "Default",
				"Container":            item.Container,
				"Size":                 item.Size,
				"Name":                 item.Name,
				"IsRemote":             false,
				"SupportsDirectPlay":   true,
				"SupportsDirectStream": true,
				"SupportsTranscoding":  true, // Decimos que sí para que la app lo considere
				"TranscodingUrl":         fmt.Sprintf("/Videos/%s/main.m3u8?VideoCodec=h264&AudioCodec=aac", item.ID),
				"TranscodingSubProtocol": "hls",
				"TranscodingContainer":   "ts",
				"SupportsResume":       true,
				"DirectStreamUrl":      streamUrl,
				"RunTimeTicks":         item.RunTimeTicks,
				"Bitrate":              item.Bitrate,
				"MediaStreams":         mediaStreams,
			},
		},
		"PlaySessionId": "79b3602d385642d99723ecdbf6a4773a",
	})
}

// StreamVideo godoc
// @Summary Stream de video (Direct Play o Transcodificación)
// @Description Sirve el archivo de video. Si se solicita mediante parámetros de transcodificación, usa FFmpeg.
// @Tags Playback
// @Param id path string true "ID del item"
// @Router /Videos/{id}/stream [get]
func StreamVideo(c *gin.Context) {
	id := c.Param("id")

	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video no encontrado"})
		return
	}

	// Si el cliente pide "Static=true", servimos el archivo original (Direct Play)
	if c.Query("Static") == "true" {
		// Asegurar el tipo MIME correcto para ayudar a la compatibilidad del navegador
		mimeType := "video/" + item.Container
		if item.Container == "mkv" {
			mimeType = "video/x-matroska"
		} else if item.Container == "avi" {
			mimeType = "video/x-msvideo"
		} else if item.Container == "m4v" {
			mimeType = "video/mp4"
		}

		c.Header("Content-Type", mimeType)
		// c.File maneja automáticamente los HTTP Range Requests (adelantar/retrasar video)
		c.File(item.Path)
		return
	}

	// De lo contrario, activamos el motor de transcodificación
	TranscodeVideo(c)
}

// ReportPlaying godoc
// @Summary Reportar inicio/progreso de reproducción
// @Description Sincroniza el estado de reproducción con el servidor.
// @Tags Playback
// @Success 204 "No Content"
// @Router /Sessions/Playing [post]
func ReportPlaying(c *gin.Context) {
	// Por ahora solo aceptamos el reporte para que la app no de error.
	c.Status(http.StatusNoContent)
}

// ReportPlayingProgress godoc
// @Summary Reportar progreso de reproducción
// @Tags Playback
// @Router /Sessions/Playing/Progress [post]
func ReportPlayingProgress(c *gin.Context) {
	var req struct {
		ItemId        string `json:"ItemId"`
		PositionTicks int64  `json:"PositionTicks"`
		IsPaused      bool   `json:"IsPaused"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusNoContent)
		return
	}

	userId := c.GetString("UserID")
	if userId == "" {
		userId = "53896590-3b41-46a4-9591-96b054a8e3f6" // Fallback al admin si no hay sesión
	}

	// Buscar el item para saber su duración total
	var item models.MediaItem
	if err := database.DB.Where("id = ?", req.ItemId).First(&item).Error; err != nil {
		c.Status(http.StatusNoContent)
		return
	}

	// Calcular si ya se consideraría "visto" (ej. 90% del video)
	isPlayed := false
	if item.RunTimeTicks > 0 && float64(req.PositionTicks) > float64(item.RunTimeTicks)*0.9 {
		isPlayed = true
	}

	userData := models.UserItemData{
		UserID:      userId,
		MediaItemID: req.ItemId,
	}

	// Update or Create usando GORM
	database.DB.Where(&userData).FirstOrCreate(&userData)

	// Solo incrementar la cuenta de reproducciones la primera vez que cruza el umbral del 90%
	if isPlayed && !userData.Played {
		userData.PlayCount++
	}

	userData.PlaybackPositionTicks = req.PositionTicks
	userData.Played = isPlayed
	userData.LastPlayedDate = time.Now()

	database.DB.Save(&userData)

	// Sincronizar con otros dispositivos vía WebSocket
	NotifyUserDataChanged(userId, req.ItemId)

	c.Status(http.StatusNoContent)
}

// ReportPlayingStopped godoc
// @Summary Reportar fin de reproducción
// @Description Notifica al servidor que el usuario cerró el reproductor.
// @Tags Playback
// @Success 204 "No Content"
// @Router /Sessions/Playing/Stopped [post]
func ReportPlayingStopped(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// parseTranscodeOptions centraliza la lógica inteligente de codec para todas las transmisiones
func parseTranscodeOptions(c *gin.Context, item models.MediaItem) transcoder.TranscodeOptions {
	startTimeTicks, _ := strconv.ParseInt(c.Query("startTimeTicks"), 10, 64)
	bitrate, _ := strconv.ParseInt(c.Query("maxBitrate"), 10, 64)

	audioStreamIndex, _ := strconv.Atoi(c.Query("AudioStreamIndex"))
	if audioStreamIndex == 0 {
		audioStreamIndex, _ = strconv.Atoi(c.Query("audioStreamIndex"))
	}

	subtitleStreamIndex, _ := strconv.Atoi(c.Query("SubtitleStreamIndex"))
	if subtitleStreamIndex == 0 {
		subtitleStreamIndex, _ = strconv.Atoi(c.Query("subtitleStreamIndex"))
	}

	reqVideoCodec := c.Query("VideoCodec")
	vCodec := "copy"

	if (bitrate > 0 && item.Bitrate > 0 && item.Bitrate > bitrate) ||
		(reqVideoCodec != "" && item.VideoCodec != "" && !strings.Contains(strings.ToLower(reqVideoCodec), strings.ToLower(item.VideoCodec))) {
		vCodec = "libx264"
	}

	reqAudioCodec := c.Query("AudioCodec")
	aCodec := "copy"

	if reqAudioCodec != "" && item.AudioCodec != "" && !strings.Contains(strings.ToLower(reqAudioCodec), strings.ToLower(item.AudioCodec)) {
		aCodec = "aac"
	}

	return transcoder.TranscodeOptions{
		StartTimeTicks:      startTimeTicks,
		Bitrate:             bitrate,
		VideoCodec:          vCodec,
		AudioCodec:          aCodec,
		AudioStreamIndex:    audioStreamIndex,
		SubtitleStreamIndex: subtitleStreamIndex,
	}
}

// TranscodeVideo maneja la transcodificación en tiempo real.
func TranscodeVideo(c *gin.Context) {
	id := c.Param("id")
	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video no encontrado"})
		return
	}

	opts := parseTranscodeOptions(c, item)

	cmd := transcoder.BuildTranscodeCmd(item, opts)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[Transcoder] Error creando stdout pipe para FFmpeg: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falló el pipe de FFmpeg"})
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[Transcoder] Error iniciando FFmpeg: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falló el inicio de FFmpeg"})
		return
	}

	// Asegurarse de que el proceso de FFmpeg se termine si el cliente se desconecta prematuramente
	// o si hay un error durante el streaming.
	defer func() {
		cmd.Process.Kill()
		cmd.Wait() // Esperar a que el proceso realmente termine
	}()

	// Configurar headers para streaming
	c.Header("Content-Type", "video/x-matroska")
	c.Header("Transfer-Encoding", "chunked")

	// Transmitir datos en vivo
	_, err = io.Copy(c.Writer, stdout) // This will block until stdout is closed or an error occurs.
	// The defer function will handle cmd.Process.Kill() and cmd.Wait()

	if err != nil {
		//log.Printf("[Transcoder] Error durante el streaming: %v", err)
	}

	cmd.Process.Kill() // Asegurarse de cerrar FFmpeg al terminar
}

// GetHlsPlaylist genera un archivo .m3u8 dinámico para el video.
func GetHlsPlaylist(c *gin.Context) {
	id := c.Param("id")
	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	segmentDuration := 10 // segundos
	totalSeconds := item.RunTimeTicks / 10000000

	if totalSeconds == 0 {
		// Fallback de seguridad por si el archivo no tiene metadatos de duración válidos
		c.String(http.StatusBadRequest, "Duración del video desconocida")
		return
	}

	numSegments := int(totalSeconds / int64(segmentDuration))
	remainder := totalSeconds % int64(segmentDuration)
	if remainder > 0 {
		numSegments++ // Agregar un segmento adicional para el residuo final
	}

	playlist := "#EXTM3U\n"
	playlist += "#EXT-X-VERSION:3\n"
	playlist += fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", segmentDuration)
	playlist += "#EXT-X-MEDIA-SEQUENCE:0\n"
	playlist += "#EXT-X-PLAYLIST-TYPE:VOD\n"

	for i := 0; i < numSegments; i++ {
		duration := segmentDuration
		// El último segmento debe durar exactamente lo que resta del video, no 10 segundos fijos
		if i == numSegments-1 && remainder > 0 {
			duration = int(remainder)
		}
		playlist += fmt.Sprintf("#EXTINF:%d.0,\n", duration)
		playlist += fmt.Sprintf("hls/%d/stream.ts\n", i)
	}

	playlist += "#EXT-X-ENDLIST\n"

	c.Header("Content-Type", "application/x-mpegURL")
	c.String(http.StatusOK, playlist)
}

// GetHlsSegment transcodifica y sirve un trozo específico de 10 segundos.
func GetHlsSegment(c *gin.Context) {
	id := c.Param("id")
	segmentIdRaw := c.Param("segmentId")
	segmentIdx, _ := strconv.Atoi(segmentIdRaw)

	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	opts := parseTranscodeOptions(c, item) // Lógica Remux ahora aplicada a Apple HLS

	cmd := transcoder.BuildHLSSegmentCmd(item, segmentIdx, 10, opts)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[Transcoder] Error creando stdout pipe para FFmpeg HLS segment: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[Transcoder] Error iniciando FFmpeg para HLS segment: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Asegurarse de que el proceso de FFmpeg se termine sin importar cómo acabe la función
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	c.Header("Content-Type", "video/mp2t")
	io.Copy(c.Writer, stdout)
}

// GetSubtitleStream extrae subtítulos embebidos al vuelo en formato WebVTT compatible con web.
func GetSubtitleStream(c *gin.Context) {
	id := c.Param("id")
	indexRaw := c.Param("index")
	index, err := strconv.Atoi(indexRaw)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	cmd := exec.Command("ffmpeg",
		"-v", "error",
		"-i", item.Path,
		"-map", fmt.Sprintf("0:%d", index),
		"-f", "webvtt",
		"-",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	cmd.Start()
	defer func() { cmd.Process.Kill(); cmd.Wait() }()

	c.Header("Content-Type", "text/vtt; charset=utf-8")
	io.Copy(c.Writer, stdout)
}
