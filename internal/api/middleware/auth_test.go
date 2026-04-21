package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseEmbyHeader(t *testing.T) {
	// Simulamos el string exacto que enviaría la app de Android TV
	header := `MediaBrowser Client="Android TV", Device="Sony Bravia", DeviceId="abc123xyz", Version="0.16.8", Token="supersecrettoken"`

	auth := parseEmbyHeader(header)

	if auth.Client != "Android TV" {
		t.Errorf("Se esperaba Client='Android TV', se obtuvo '%s'", auth.Client)
	}
	if auth.Device != "Sony Bravia" {
		t.Errorf("Se esperaba Device='Sony Bravia', se obtuvo '%s'", auth.Device)
	}
	if auth.DeviceId != "abc123xyz" {
		t.Errorf("Se esperaba DeviceId='abc123xyz', se obtuvo '%s'", auth.DeviceId)
	}
	if auth.Token != "supersecrettoken" {
		t.Errorf("Se esperaba Token='supersecrettoken', se obtuvo '%s'", auth.Token)
	}
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.OPTIONS("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Se esperaba código 204 para OPTIONS, se obtuvo %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Falta el header CORS Access-Control-Allow-Origin")
	}
}
