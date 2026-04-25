package database

import (
	"fmt"
	"log"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000", config.AppConfig.Database.Path)
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatal("Error conectando a la base de datos: ", err)
	}

	err = DB.AutoMigrate(&models.User{}, &models.MediaItem{}, &models.MediaStream{}, &models.UserItemData{})
	if err != nil {
		log.Fatal("Error en la migración de base de datos: ", err)
	}

	// Optimización para SQLite: Solo un escritor a la vez
	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		admin := models.User{
			ID:       config.AppConfig.Jellyfin.AdminUUID,
			Username: "admin",
			Password: string(hashedPassword),
			IsAdmin:  true,
		}
		DB.Create(&admin)
		log.Println("Usuario administrador 'admin' creado con UUID desde configuración.")
	}
}
