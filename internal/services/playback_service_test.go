package services

import (
	"testing"

	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPlaybackTest() (PlaybackService, *gorm.DB) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&models.MediaItem{}, &models.UserItemData{})
	mediaRepo := repository.NewMediaRepository(db)
	userDataRepo := repository.NewUserItemDataRepository(db)
	return NewPlaybackService(mediaRepo, userDataRepo), db
}

func TestPlaybackService(t *testing.T) {
	service, db := setupPlaybackTest()

	item := &models.MediaItem{ID: "1", Name: "Test Movie"}
	db.Create(item)

	t.Run("UpdateProgress New", func(t *testing.T) {
		err := service.UpdateProgress("user1", "1", 1000, false)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		userDataRepo := repository.NewUserItemDataRepository(db)
		data, _ := userDataRepo.Get("user1", "1")
		if data.PlaybackPositionTicks != 1000 {
			t.Errorf("expected 1000 ticks, got %d", data.PlaybackPositionTicks)
		}
	})

	t.Run("UpdateProgress Update", func(t *testing.T) {
		err := service.UpdateProgress("user1", "1", 2000, false)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		userDataRepo := repository.NewUserItemDataRepository(db)
		data, _ := userDataRepo.Get("user1", "1")
		if data.PlaybackPositionTicks != 2000 {
			t.Errorf("expected 2000 ticks, got %d", data.PlaybackPositionTicks)
		}
	})
}
