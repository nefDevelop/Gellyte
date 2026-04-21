package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	
	r.GET("/Library/VirtualFolders", GetVirtualFolders)
	r.GET("/Items", GetItems)
	r.GET("/Items/:id", GetItemDetails)
	r.GET("/Items/:id/Images/:imageType", GetItemImage)
	r.GET("/Users/:id/Images/Primary", GetUserPrimaryImage)
	r.GET("/Shows/NextUp", GetNextUp)
	r.GET("/Items/Resume", GetResumeItems)
	r.GET("/Suggestions", GetSuggestions)
	
	return r
}

func TestGetVirtualFolders(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Library/VirtualFolders", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Se esperaba 200, se obtuvo %v", w.Code)
	}

	var response []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if len(response) != 2 {
		t.Errorf("Se esperaban 2 carpetas virtuales, se obtuvieron %d", len(response))
	}
}

func TestGetItems(t *testing.T) {
	setupTestDB()
	items := []models.MediaItem{
		{ID: "m1", Name: "Action Movie", Type: "Movie", Path: "/media/m1.mp4"},
		{ID: "m2", Name: "Comedy Movie", Type: "Movie", Path: "/media/m2.mp4"},
		{ID: "s1", Name: "Drama Series", Type: "Series", Path: "/media/s1"},
	}
	for _, item := range items {
		database.DB.Create(&item)
	}

	r := setupRouter()
	tests := []struct {
		name          string
		query         string
		expectedCount int
	}{
		{"All items", "", 3},
		{"Filter Movie", "?IncludeItemTypes=Movie", 2},
		{"Search Action", "?SearchTerm=Action", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/Items"+tt.query, nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Se esperaba 200, se obtuvo %v", w.Code)
			}

			var resp struct {
				Items []BaseItemDto `json:"Items"`
			}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if len(resp.Items) != tt.expectedCount {
				t.Errorf("Esperados %d, obtenidos %d", tt.expectedCount, len(resp.Items))
			}
		})
	}
}

func TestGetItemDetails(t *testing.T) {
	setupTestDB()
	movieID := "movie123"
	database.DB.Create(&models.MediaItem{
		ID:   movieID,
		Name: "Detailed Movie",
		Type: "Movie",
	})

	r := setupRouter()
	tests := []struct {
		name         string
		id           string
		expectedCode int
	}{
		{"Existing", movieID, http.StatusOK},
		{"Virtual", "12345678-1234-1234-1234-123456789012", http.StatusOK},
		{"Non-existing", "wrong-id", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/Items/"+tt.id, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Esperado %d, obtenido %d", tt.expectedCode, w.Code)
			}
		})
	}
}

func TestGetItemImage(t *testing.T) {
	setupTestDB()
	tmpDir, _ := os.MkdirTemp("", "gellyte-test-*")
	defer os.RemoveAll(tmpDir)

	posterPath := filepath.Join(tmpDir, "poster.jpg")
	os.WriteFile(posterPath, []byte("fake image"), 0644)

	id := "img-test"
	database.DB.Create(&models.MediaItem{
		ID:   id,
		Name: "Image Test",
		Path: filepath.Join(tmpDir, "movie.mp4"),
	})

	r := setupRouter()
	
	// Test Primary
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Items/"+id+"/Images/Primary", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtenido %d", w.Code)
	}

	// Test 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/Items/no-id/Images/Primary", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Esperado 404, obtenido %d", w.Code)
	}
}

func TestGetUserPrimaryImage(t *testing.T) {
	setupTestDB()
	userID := "u1"
	database.DB.Create(&models.User{ID: userID, Username: "TestUser"})

	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Users/"+userID+"/Images/Primary", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtenido %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "image/svg+xml" {
		t.Errorf("Esperado image/svg+xml, obtenido %s", w.Header().Get("Content-Type"))
	}
}

func TestGetNextUp(t *testing.T) {
	setupTestDB()
	userId := "53896590-3b41-46a4-9591-96b054a8e3f6"
	database.DB.Create(&models.MediaItem{ID: "series1", Name: "Series 1", Type: "Series"})
	database.DB.Create(&models.MediaItem{ID: "ep1", Name: "Episode 1", Type: "Episode", ParentID: "series1"})
	database.DB.Create(&models.MediaItem{ID: "ep2", Name: "Episode 2", Type: "Episode", ParentID: "series1"})
	database.DB.Create(&models.UserItemData{UserID: userId, MediaItemID: "ep1", Played: true})

	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Shows/NextUp", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtenido %d", w.Code)
	}

	var resp struct{ Items []BaseItemDto }
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Items) != 1 || resp.Items[0].Id != "ep2" {
		t.Errorf("Esperado ep2 en NextUp, obtenidos %v", resp.Items)
	}
}

func TestGetResumeItems(t *testing.T) {
	setupTestDB()
	userId := "53896590-3b41-46a4-9591-96b054a8e3f6"
	database.DB.Create(&models.MediaItem{ID: "m1", Name: "Movie 1", Type: "Movie"})
	database.DB.Create(&models.UserItemData{UserID: userId, MediaItemID: "m1", PlaybackPositionTicks: 1000, Played: false})

	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Items/Resume", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtenido %d", w.Code)
	}

	var resp struct{ Items []BaseItemDto }
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Items) != 1 {
		t.Errorf("Esperado 1 item en Resume, obtenidos %d", len(resp.Items))
	}
}

func TestGetSuggestions(t *testing.T) {
	setupTestDB()
	database.DB.Create(&models.MediaItem{ID: "m1", Name: "Movie 1", Type: "Movie"})

	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Suggestions?userId=u1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtenido %d", w.Code)
	}

	var resp struct{ Items []BaseItemDto }
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Items) != 1 {
		t.Errorf("Esperado 1 item en Suggestions, obtenidos %d", len(resp.Items))
	}
}
