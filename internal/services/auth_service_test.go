package services

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthTest() (AuthService, *gorm.DB) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&models.User{})
	userRepo := repository.NewUserRepository(db)
	return NewAuthService(userRepo), db
}

func TestAuthService(t *testing.T) {
	service, db := setupAuthTest()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{
		ID:       "1",
		Username: "testuser",
		Password: string(hashed),
	}
	db.Create(user)

	t.Run("Authenticate Success", func(t *testing.T) {
		u, token, err := service.Authenticate("testuser", "password123", "device1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if u.Username != "testuser" {
			t.Errorf("expected username testuser, got %s", u.Username)
		}
		if token == "" {
			t.Error("expected token, got empty string")
		}
	})

	t.Run("Authenticate Wrong Password", func(t *testing.T) {
		_, _, err := service.Authenticate("testuser", "wrong", "device1")
		if err == nil || err.Error() != "password incorrecta" {
			t.Errorf("expected password incorrecta error, got %v", err)
		}
	})

	t.Run("Authenticate User Not Found", func(t *testing.T) {
		_, _, err := service.Authenticate("unknown", "password123", "device1")
		if err == nil || err.Error() != "usuario no encontrado" {
			t.Errorf("expected usuario no encontrado error, got %v", err)
		}
	})
}
