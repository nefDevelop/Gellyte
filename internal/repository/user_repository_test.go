package repository

import (
	"testing"

	"github.com/gellyte/gellyte/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&models.User{})
	return db
}

func TestUserRepository(t *testing.T) {
	db := setupUserTestDB()
	repo := NewUserRepository(db)

	user := &models.User{
		ID:       "1",
		Username: "testuser",
		Password: "password123",
	}

	db.Create(user)

	t.Run("GetByUsername", func(t *testing.T) {
		found, err := repo.GetByUsername("testuser")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if found.Username != "testuser" {
			t.Errorf("expected username testuser, got %s", found.Username)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		found, err := repo.GetByID("1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if found.ID != "1" {
			t.Errorf("expected ID 1, got %s", found.ID)
		}
	})

	t.Run("ListAll", func(t *testing.T) {
		users, err := repo.ListAll()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(users) != 1 {
			t.Errorf("expected 1 user, got %d", len(users))
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := repo.GetByUsername("nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent user, got nil")
		}
	})
}
