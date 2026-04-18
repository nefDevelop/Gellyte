package database

import (
	"log"

	"github.com/gellyte/gellyte/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("gellyte.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Error conectando a la base de datos: ", err)
	}

	err = DB.AutoMigrate(&models.User{}, &models.MediaItem{})
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
