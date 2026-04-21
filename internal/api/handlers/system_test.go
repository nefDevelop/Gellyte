package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupSystemRouter() (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := setupHandler()
	
	r.GET("/System/Info/Public", h.GetPublicInfo)
	r.GET("/System/Info", h.GetSystemInfo)
	r.GET("/System/Ping", h.GetPingSystem)
	r.GET("/System/Endpoint", h.GetEndpointInfo)
	
	return r, h
}

func TestGetPingSystem(t *testing.T) {
	r, _ := setupSystemRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/System/Ping", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Se esperaba 200, se obtuvo %v", w.Code)
	}

	if w.Body.String() != "Jellyfin Server" {
		t.Errorf("Se esperaba 'Jellyfin Server', se obtuvo '%v'", w.Body.String())
	}
}

func TestGetPublicInfo(t *testing.T) {
	r, _ := setupSystemRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/System/Info/Public", nil)
	req.Host = "localhost:8081"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Se esperaba 200, se obtuvo %v", w.Code)
	}

	var response PublicSystemInfo
	json.Unmarshal(w.Body.Bytes(), &response)
	if response.ServerName != "Gellyte" {
		t.Errorf("Se esperaba 'Gellyte', se obtuvo '%s'", response.ServerName)
	}
}

func TestGetSystemInfo(t *testing.T) {
	r, _ := setupSystemRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/System/Info", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Se esperaba 200, se obtuvo %v", w.Code)
	}

	var response SystemInfo
	json.Unmarshal(w.Body.Bytes(), &response)
	if response.Version != "10.11.8" {
		t.Errorf("Versión incorrecta: %s", response.Version)
	}
}

func TestGetEndpointInfo(t *testing.T) {
	r, _ := setupSystemRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/System/Endpoint", nil)
	req.Host = "test-host"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Se esperaba 200, se obtuvo %v", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["Address"] != "http://test-host" {
		t.Errorf("Address incorrecta: %v", response["Address"])
	}
}
