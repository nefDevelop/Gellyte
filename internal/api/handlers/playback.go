package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gellyte/gellyte/internal/database"
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
				"SupportsResume":       true,
				"DirectStreamUrl":      streamUrl,
				"RunTimeTicks":         item.RunTimeTicks,
				"Bitrate":              item.Bitrate,
				"MediaStreams": []gin.H{
					{
						"Type":         "Video",
						"Codec":        item.VideoCodec,
						"IsInterlaced": false,
						"Width":        item.Width,
						"Height":       item.Height,
						"BitRate":      item.Bitrate,
					},
					{
						"Type":      "Audio",
						"Codec":     item.AudioCodec,
						"IsDefault": true,
					},
				},
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

	userData.PlaybackPositionTicks = req.PositionTicks
	userData.Played = isPlayed
	userData.LastPlayedDate = time.Now()
	if isPlayed {
		userData.PlayCount++
	}

	database.DB.Save(&userData)

	// Sincronizar con otros dispositivos vía WebSocket
	NotifyUserDataChanged(userId, req.ItemId)

	c.Status(http.StatusNoContent)
}

// TranscodeVideo maneja la transcodificación en tiempo real.
func TranscodeVideo(c *gin.Context) {
	id := c.Param("id")
	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video no encontrado"})
		return
	}

	// Extraer opciones de la URL (usadas por Jellyfin)
	startTimeTicks, _ := strconv.ParseInt(c.Query("startTimeTicks"), 10, 64)
	bitrate, _ := strconv.ParseInt(c.Query("maxBitrate"), 10, 64)

	opts := transcoder.TranscodeOptions{
		StartTimeTicks: startTimeTicks,
		Bitrate:        bitrate,
		VideoCodec:     "libx264",
		AudioCodec:     "aac",
	}

	cmd := transcoder.BuildTranscodeCmd(item, opts)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falló el pipe de FFmpeg"})
		return
	}

	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falló el inicio de FFmpeg"})
		return
	}

	// Configurar headers para streaming
	c.Header("Content-Type", "video/x-matroska")
	c.Header("Transfer-Encoding", "chunked")

	// Transmitir datos en vivo
	_, err = io.Copy(c.Writer, stdout)
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
	numSegments := int(totalSeconds / int64(segmentDuration))

	playlist := "#EXTM3U\n"
	playlist += "#EXT-X-VERSION:3\n"
	playlist += fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", segmentDuration)
	playlist += "#EXT-X-MEDIA-SEQUENCE:0\n"
	playlist += "#EXT-X-PLAYLIST-TYPE:VOD\n"

	for i := 0; i < numSegments; i++ {
		playlist += fmt.Sprintf("#EXTINF:%d.0,\n", segmentDuration)
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

	opts := transcoder.TranscodeOptions{
		VideoCodec: "libx264",
		AudioCodec: "aac",
	}

	cmd := transcoder.BuildHLSSegmentCmd(item, segmentIdx, 10, opts)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if err := cmd.Start(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "video/mp2t")
	io.Copy(c.Writer, stdout)
	cmd.Wait()
}
