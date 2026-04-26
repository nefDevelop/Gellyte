package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gin-gonic/gin"
)

func setupPlaybackRouter() (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := setupHandler()
	
	r.GET("/Items/:id/PlaybackInfo", h.Playback.GetPlaybackInfo)
	r.POST("/Sessions/Playing/Progress", h.Playback.ReportPlayingProgress)
	r.GET("/Videos/:id/stream", h.Playback.StreamVideo)
	
	return r, h
}

func TestParseTranscodeOptions(t *testing.T) {
	item := models.MediaItem{
		VideoCodec: "hevc",
		AudioCodec: "eac3",
		Bitrate:    5000000,
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedVCodec string
		expectedACodec string
	}{
		{"Direct Play", "?VideoCodec=hevc,h264&AudioCodec=eac3,aac", "copy", "copy"},
		{"Transcode Video", "?VideoCodec=h264&AudioCodec=eac3", "libx264", "copy"},
		{"Transcode Audio", "?VideoCodec=hevc,h264&AudioCodec=aac,mp3", "copy", "aac"},
		{"Bitrate Limit", "?VideoCodec=hevc&AudioCodec=eac3&maxBitrate=2000000", "libx264", "copy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/dummy"+tt.queryParams, nil)

			opts := parseTranscodeOptions(c, item)

			if opts.VideoCodec != tt.expectedVCodec {
				t.Errorf("%s: VideoCodec esperado %s, obtenido %s", tt.name, tt.expectedVCodec, opts.VideoCodec)
			}
			if opts.AudioCodec != tt.expectedACodec {
				t.Errorf("%s: AudioCodec esperado %s, obtenido %s", tt.name, tt.expectedACodec, opts.AudioCodec)
			}
		})
	}
}

func TestGetPlaybackInfo(t *testing.T) {
	setupTestDB()
	database.DB.Create(&models.MediaItem{ID: "m1", Name: "Movie 1"})

	r, _ := setupPlaybackRouter()
	
	// Success
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Items/m1/PlaybackInfo", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtenido %d", w.Code)
	}

	// Not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/Items/none/PlaybackInfo", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Esperado 404, obtenido %d", w.Code)
	}
}

func TestReportPlayingProgress(t *testing.T) {
	setupTestDB()
	database.DB.Create(&models.MediaItem{ID: "m1", Name: "Movie 1", RunTimeTicks: 10000})

	r, _ := setupPlaybackRouter()
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]interface{}{
		"ItemId":        "m1",
		"PositionTicks": 5000,
		"IsPaused":      false,
	})
	req, _ := http.NewRequest("POST", "/Sessions/Playing/Progress", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Esperado 204, obtenido %d", w.Code)
	}

	// Verify DB update
	var ud models.UserItemData
	database.DB.Where("media_item_id = ?", "m1").First(&ud)
	if ud.PlaybackPositionTicks != 5000 {
		t.Errorf("Posición no guardada correctamente: %d", ud.PlaybackPositionTicks)
	}
}

func TestStreamVideo_NotFound(t *testing.T) {
	setupTestDB()
	r, _ := setupPlaybackRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Videos/none/stream", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Esperado 404, obtenido %d", w.Code)
	}
}
