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

func setupAuthRouter() (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := setupHandler()
	
	r.GET("/Users/Public", h.GetPublicUsers)
	r.POST("/Users/AuthenticateByName", h.AuthenticateByName)
	r.GET("/Users/Me", h.GetCurrentUser)
	r.GET("/Users/:id", h.GetUserById)
	r.GET("/Users/:id/Views", h.GetUserViews)
	r.GET("/DisplayPreferences/:id", h.GetDisplayPreferences)
	
	return r, h
}

func TestGetPublicUsers(t *testing.T) {
	setupTestDB()
	database.DB.Create(&models.User{
		ID:       "u1",
		Username: "user1",
	})

	r, _ := setupAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Users/Public", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtenido %d", w.Code)
	}

	var resp []UserDto
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 1 || resp[0].Name != "user1" {
		t.Errorf("Respuesta inesperada: %v", resp)
	}
}

func TestAuthenticateByName(t *testing.T) {
	setupTestDB()
	database.DB.Create(&models.User{
		ID:       "u1",
		Username: "user1",
		Password: "password123",
	})

	r, _ := setupAuthRouter()

	// Success
	w := httptest.NewRecorder()
	body, _ := json.Marshal(AuthRequest{Username: "user1", Pw: "password123"})
	req, _ := http.NewRequest("POST", "/Users/AuthenticateByName", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Success: Esperado 200, obtenido %d", w.Code)
	}

	// Wrong password
	w = httptest.NewRecorder()
	body, _ = json.Marshal(AuthRequest{Username: "user1", Pw: "wrong"})
	req, _ = http.NewRequest("POST", "/Users/AuthenticateByName", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Wrong PW: Esperado 401, obtenido %d", w.Code)
	}

	// Not found
	w = httptest.NewRecorder()
	body, _ = json.Marshal(AuthRequest{Username: "noone", Pw: "pw"})
	req, _ = http.NewRequest("POST", "/Users/AuthenticateByName", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Not found: Esperado 401, obtenido %d", w.Code)
	}
}

func TestGetUserById(t *testing.T) {
	setupTestDB()
	database.DB.Create(&models.User{
		ID:       "u1",
		Username: "user1",
	})

	r, _ := setupAuthRouter()
	
	// Found
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Users/u1", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Found: Esperado 200, obtenido %d", w.Code)
	}

	// Not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/Users/none", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Not found: Esperado 404, obtenido %d", w.Code)
	}
}

func TestGetUserViews(t *testing.T) {
	r, _ := setupAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Users/u1/Views", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtenido %d", w.Code)
	}

	var resp struct{ Items []interface{} }
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Items) != 2 {
		t.Errorf("Esperadas 2 vistas, obtenidas %d", len(resp.Items))
	}
}

func TestGetDisplayPreferences(t *testing.T) {
	r, _ := setupAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/DisplayPreferences/u1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtenido %d", w.Code)
	}
}
