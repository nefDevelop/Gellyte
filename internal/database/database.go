package database

import (
	"log"

	"github.com/gellyte/gellyte/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("gellyte.db?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatal("Error conectando a la base de datos: ", err)
	}

	err = DB.AutoMigrate(&models.User{}, &models.MediaItem{}, &models.UserItemData{})
	if err != nil {
		log.Fatal("Error en la migración de base de datos: ", err)
	}

	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		admin := models.User{
			ID:       "53896590-3b41-46a4-9591-96b054a8e3f6", // UUID fijo para admin con guiones
			Username: "admin",
			Password: "admin",
			IsAdmin:  true,
		}
		DB.Create(&admin)
		log.Println("Usuario administrador 'admin' creado con UUID fijo.")
	}
}
