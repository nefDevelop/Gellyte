package repository

import (
	"testing"

	"github.com/gellyte/gellyte/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMediaTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&models.MediaItem{}, &models.MediaStream{})
	return db
}

func TestMediaRepository(t *testing.T) {
	db := setupMediaTestDB()
	repo := NewMediaRepository(db)

	items := []models.MediaItem{
		{ID: "1", Name: "Movie A", Type: "Movie"},
		{ID: "2", Name: "Series B", Type: "Series"},
		{ID: "3", Name: "Episode 1", Type: "Episode", ParentID: "2"},
		{ID: "4", Name: "Episode 2", Type: "Episode", ParentID: "2"},
	}

	for _, item := range items {
		db.Create(&item)
	}

	t.Run("GetByID", func(t *testing.T) {
		found, err := repo.GetByID("1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if found.Name != "Movie A" {
			t.Errorf("expected Movie A, got %s", found.Name)
		}
	})

	t.Run("GetByParentID", func(t *testing.T) {
		children, err := repo.GetByParentID("2")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(children) != 2 {
			t.Errorf("expected 2 children, got %d", len(children))
		}
	})

	t.Run("Search", func(t *testing.T) {
		results, total, err := repo.Search("Movie", nil, 0, 10)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if results[0].Name != "Movie A" {
			t.Errorf("expected Movie A, got %s", results[0].Name)
		}
	})

	t.Run("GetByIDs", func(t *testing.T) {
		results, err := repo.GetByIDs([]string{"1", "2"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("GetByType", func(t *testing.T) {
		results, err := repo.GetByType("Series", 10)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
	})

	t.Run("GetNextEpisode", func(t *testing.T) {
		next, err := repo.GetNextEpisode("2", "Episode 1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if next.Name != "Episode 2" {
			t.Errorf("expected Episode 2, got %s", next.Name)
		}
	})

	t.Run("Count", func(t *testing.T) {
		count, err := repo.Count("Episode")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if count != 2 {
			t.Errorf("expected 2, got %d", count)
		}
	})
}
