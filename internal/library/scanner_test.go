package library

import (
	"testing"
	"github.com/gellyte/gellyte/internal/models"
)

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Archivo MP4 normal", "/media/movies/Iron Man.mp4", true},
		{"Archivo MKV normal", "/media/series/show.mkv", true},
		{"Extensión en mayúsculas", "/media/movies/Thor.MP4", true},
		{"Extensión mixta", "/media/movies/Hulk.MkV", true},
		{"Archivo de subtítulos SRT", "/media/movies/Iron Man.srt", false},
		{"Archivo de imagen JPG", "/media/movies/poster.jpg", false},
		{"Archivo sin extensión", "/media/movies/Iron Man", false},
		{"Ruta con puntos en el nombre", "/media/movies/Mr. Robot.S01E01.mp4", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVideoFile(tt.path)
			if result != tt.expected {
				t.Errorf("isVideoFile(%q) = %v; se esperaba %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestProcessDirectory(t *testing.T) {
	db := setupTestDB()
	
	libRoot := "/media/series"
	libType := "series"

	t.Run("Create Series Folder", func(t *testing.T) {
		path := "/media/series/Breaking Bad"
		processDirectory(path, libType, libRoot)

		var item models.MediaItem
		err := db.Where("path = ?", path).First(&item).Error
		if err != nil {
			t.Fatalf("expected item to be created, got error: %v", err)
		}
		if item.Type != "Series" {
			t.Errorf("expected Type Series, got %s", item.Type)
		}
		if item.Name != "Breaking Bad" {
			t.Errorf("expected Name Breaking Bad, got %s", item.Name)
		}
	})

	t.Run("Create Season Folder", func(t *testing.T) {
		parentPath := "/media/series/Breaking Bad"
		path := "/media/series/Breaking Bad/Season 1"
		processDirectory(path, libType, libRoot)

		var item models.MediaItem
		err := db.Where("path = ?", path).First(&item).Error
		if err != nil {
			t.Fatalf("expected item to be created, got error: %v", err)
		}
		if item.Type != "Season" {
			t.Errorf("expected Type Season, got %s", item.Type)
		}
		
		var parent models.MediaItem
		db.Where("path = ?", parentPath).First(&parent)
		if item.ParentID != parent.ID {
			t.Errorf("expected ParentID %s, got %s", parent.ID, item.ParentID)
		}
	})
}

func TestProcessFile(t *testing.T) {
	db := setupTestDB()
	
	libType := "movies"

	t.Run("Create Movie File", func(t *testing.T) {
		path := "/media/movies/Inception.mp4"
		processFile(path, libType)

		var item models.MediaItem
		err := db.Where("path = ?", path).First(&item).Error
		if err != nil {
			t.Fatalf("expected item to be created, got error: %v", err)
		}
		if item.Type != "Movie" {
			t.Errorf("expected Type Movie, got %s", item.Type)
		}
		if item.Name != "Inception" {
			t.Errorf("expected Name Inception, got %s", item.Name)
		}
	})
}
