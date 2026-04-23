package repository

import (
	"testing"

	"github.com/gellyte/gellyte/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserDataTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&models.UserItemData{})
	return db
}

func TestUserItemDataRepository(t *testing.T) {
	db := setupUserDataTestDB()
	repo := NewUserItemDataRepository(db)

	data := &models.UserItemData{
		UserID:               "user1",
		MediaItemID:          "item1",
		PlaybackPositionTicks: 1000,
		Played:               false,
	}

	db.Create(data)

	t.Run("Get", func(t *testing.T) {
		found, err := repo.Get("user1", "item1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if found.PlaybackPositionTicks != 1000 {
			t.Errorf("expected 1000 ticks, got %d", found.PlaybackPositionTicks)
		}
	})

	t.Run("Upsert", func(t *testing.T) {
		data.PlaybackPositionTicks = 2000
		err := repo.Upsert(data)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		found, _ := repo.Get("user1", "item1")
		if found.PlaybackPositionTicks != 2000 {
			t.Errorf("expected 2000 ticks, got %d", found.PlaybackPositionTicks)
		}
	})

	t.Run("GetResume", func(t *testing.T) {
		results, err := repo.GetResume("user1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
	})

	t.Run("GetPlayed", func(t *testing.T) {
		data.Played = true
		repo.Upsert(data)
		results, err := repo.GetPlayed("user1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
	})
}
