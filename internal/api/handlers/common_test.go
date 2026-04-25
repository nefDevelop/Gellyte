package handlers

import (
	"sync"
	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/repository"
	"github.com/gellyte/gellyte/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	hubOnce sync.Once
	confOnce sync.Once
)

func setupTestDB() {
	// Iniciar Configuración una sola vez para los tests
	confOnce.Do(func() {
		config.InitConfig()
	})

	// Iniciar una base de datos SQLite efímera en memoria RAM
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&models.User{}, &models.MediaItem{}, &models.MediaStream{}, &models.UserItemData{})
	database.DB = db // Sobrescribir la variable global para los tests
	
	hubOnce.Do(func() {
		go GlobalHub.Run()
	})
}

func setupHandler() *Handler {
	userRepo := repository.NewUserRepository(database.DB)
	mediaRepo := repository.NewMediaRepository(database.DB)
	userDataRepo := repository.NewUserItemDataRepository(database.DB)

	authService := services.NewAuthService(userRepo)
	libraryService := services.NewLibraryService(mediaRepo, userDataRepo)
	playbackService := services.NewPlaybackService(mediaRepo, userDataRepo)

	return NewHandler(authService, libraryService, playbackService)
}
