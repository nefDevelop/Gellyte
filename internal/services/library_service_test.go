package services

import (
	"testing"

	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupLibraryTest() (LibraryService, *gorm.DB) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&models.MediaItem{}, &models.MediaStream{}, &models.UserItemData{})
	mediaRepo := repository.NewMediaRepository(db)
	userDataRepo := repository.NewUserItemDataRepository(db)
	return NewLibraryService(mediaRepo, userDataRepo), db
}

func TestLibraryService(t *testing.T) {
	service, db := setupLibraryTest()

	items := []models.MediaItem{
		{ID: "1", Name: "Movie A", Type: "Movie", Path: "/media/movieA.mp4"},
		{ID: "2", Name: "Series B", Type: "Series", Path: "/media/seriesB"},
		{ID: "3", Name: "Episode 1", Type: "Episode", ParentID: "2", Path: "/media/seriesB/e1.mp4"},
	}
	for i := range items {
		err := db.Create(&items[i]).Error
		if err != nil {
			t.Fatalf("failed to create item: %v", err)
		}
	}

	t.Run("GetItems", func(t *testing.T) {
		res, total, err := service.GetItems(GetItemsParams{SearchTerm: "Movie", Limit: 10})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if total == 0 {
			var count int64
			db.Model(&models.MediaItem{}).Count(&count)
			t.Errorf("total is 0, database count is %d", count)
		}
		if len(res) == 0 {
			t.Fatal("res is empty")
		}
		if res[0].Name != "Movie A" {
			t.Errorf("unexpected results: name %s", res[0].Name)
		}
	})

	t.Run("GetResumeItems", func(t *testing.T) {
		db.Create(&models.UserItemData{
			UserID:               "user1",
			MediaItemID:          "1",
			PlaybackPositionTicks: 500,
			Played:               false,
		})
		res, err := service.GetResumeItems("user1", 10)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(res) != 1 || res[0].ID != "1" {
			t.Errorf("expected item 1, got %v", res)
		}
	})
}
