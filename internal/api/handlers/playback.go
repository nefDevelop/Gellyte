package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/transcoder"
	"github.com/gin-gonic/gin"
)

// GetPlaybackInfo godoc
func (h *Handler) GetPlaybackInfo(c *gin.Context) {
	id := c.Param("id")

	item, err := h.PlaybackService.GetItem(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item no encontrado"})
		return
	}

	streamUrl := fmt.Sprintf("/Videos/%s/stream", item.ID)

	mediaStreams := []gin.H{}
	for _, s := range item.MediaStreams {
		mediaStreams = append(mediaStreams, gin.H{
			"Type":      s.Type,
			"Index":     s.Index,
			"Codec":     s.Codec,
			"Language":  s.Language,
			"IsDefault": s.IsDefault,
			"Width":     s.Width,
			"Height":    s.Height,
		})
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
				"SupportsTranscoding":  true,
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
func (h *Handler) StreamVideo(c *gin.Context) {
	id := c.Param("id")

	item, err := h.PlaybackService.GetItem(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video no encontrado"})
		return
	}

	if c.Query("Static") == "true" {
		mimeType := "video/" + item.Container
		if item.Container == "mkv" {
			mimeType = "video/x-matroska"
		} else if item.Container == "avi" {
			mimeType = "video/x-msvideo"
		} else if item.Container == "m4v" {
			mimeType = "video/mp4"
		}

		c.Header("Content-Type", mimeType)
		c.File(item.Path)
		return
	}

	h.TranscodeVideo(c)
}

// ReportPlaying godoc
func (h *Handler) ReportPlaying(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// ReportPlayingProgress godoc
func (h *Handler) ReportPlayingProgress(c *gin.Context) {
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
		userId = config.AppConfig.Jellyfin.AdminUUID
	}

	err := h.PlaybackService.UpdateProgress(userId, req.ItemId, req.PositionTicks, req.IsPaused)
	if err == nil {
		NotifyUserDataChanged(userId, req.ItemId)
	}

	c.Status(http.StatusNoContent)
}

// ReportPlayingStopped godoc
func (h *Handler) ReportPlayingStopped(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

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
		vCodec = config.AppConfig.Transcoder.DefaultCodec
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
func (h *Handler) TranscodeVideo(c *gin.Context) {
	id := c.Param("id")
	item, err := h.PlaybackService.GetItem(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video no encontrado"})
		return
	}

	opts := parseTranscodeOptions(c, *item)
	cmd := h.PlaybackService.StartTranscode(*item, opts)

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

	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	c.Header("Content-Type", "video/x-matroska")
	c.Header("Transfer-Encoding", "chunked")

	_, err = io.Copy(c.Writer, stdout)
	if err != nil {
		// handle err
	}
	cmd.Process.Kill()
}

// GetHlsPlaylist genera un archivo .m3u8 dinámico para el video.
func (h *Handler) GetHlsPlaylist(c *gin.Context) {
	id := c.Param("id")
	item, err := h.PlaybackService.GetItem(id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	segmentDuration := 10
	totalSeconds := item.RunTimeTicks / 10000000

	if totalSeconds == 0 {
		c.String(http.StatusBadRequest, "Duración del video desconocida")
		return
	}

	numSegments := int(totalSeconds / int64(segmentDuration))
	remainder := totalSeconds % int64(segmentDuration)
	if remainder > 0 {
		numSegments++
	}

	playlist := "#EXTM3U\n"
	playlist += "#EXT-X-VERSION:3\n"
	playlist += fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", segmentDuration)
	playlist += "#EXT-X-MEDIA-SEQUENCE:0\n"
	playlist += "#EXT-X-PLAYLIST-TYPE:VOD\n"

	for i := 0; i < numSegments; i++ {
		duration := segmentDuration
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
func (h *Handler) GetHlsSegment(c *gin.Context) {
	id := c.Param("id")
	segmentIdRaw := c.Param("segmentId")
	segmentIdx, _ := strconv.Atoi(segmentIdRaw)

	item, err := h.PlaybackService.GetItem(id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	opts := parseTranscodeOptions(c, *item)

	cmd := h.PlaybackService.GetHLSSegment(*item, segmentIdx, 10, opts)
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

	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	c.Header("Content-Type", "video/mp2t")
	io.Copy(c.Writer, stdout)
}

// GetSubtitleStream extrae subtítulos embebidos al vuelo en formato WebVTT compatible con web.
func (h *Handler) GetSubtitleStream(c *gin.Context) {
	id := c.Param("id")
	indexRaw := c.Param("index")
	index, err := strconv.Atoi(indexRaw)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	item, err := h.PlaybackService.GetItem(id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	cmd := exec.Command(config.AppConfig.Transcoder.FFmpegPath,
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
