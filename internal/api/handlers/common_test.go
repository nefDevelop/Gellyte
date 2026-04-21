package handlers

import (
	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() {
	// Iniciar una base de datos SQLite efímera en memoria RAM
	// Usamos un nombre único o simplemente :memory: sin cache=shared para asegurar aislamiento por test
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&models.User{}, &models.MediaItem{}, &models.UserItemData{})
	database.DB = db // Sobrescribir la variable global para los tests
}
