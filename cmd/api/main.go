package main

import (
	"log"

	_ "github.com/gellyte/gellyte/docs"
	"github.com/gellyte/gellyte/internal/api/handlers"
	"github.com/gellyte/gellyte/internal/api/middleware"
	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/library"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// @title Gellyte API
// @version 1.0
// @description Servidor compatible con Jellyfin escrito en Go.
// @host localhost:8080
// @BasePath /
func main() {
	database.InitDB()
	seedDatabase()
	go library.WatchFolder("./media/peliculas")

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.EmbyAuthMiddleware())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// --- RUTAS COMPATIBLES (CON ALIAS) ---

	// Sistema
	systemGroup := r.Group("/")
	{
		systemGroup.GET("/System/Info/Public", handlers.GetPublicInfo)
		systemGroup.GET("/emby/System/Info/Public", handlers.GetPublicInfo)
		systemGroup.GET("/System/Info", handlers.GetSystemInfo)
		systemGroup.GET("/emby/System/Info", handlers.GetSystemInfo)
	}

	// Usuarios y Auth
	authGroup := r.Group("/")
	{
		authGroup.GET("/Users/Public", handlers.GetPublicUsers)
		authGroup.GET("/emby/Users/Public", handlers.GetPublicUsers)
		authGroup.GET("/Users", handlers.GetPublicUsers)
		authGroup.GET("/emby/Users", handlers.GetPublicUsers)

		authGroup.POST("/Users/AuthenticateByName", handlers.AuthenticateByName)
		authGroup.POST("/Users/authenticatebyname", handlers.AuthenticateByName)
		authGroup.POST("/emby/Users/AuthenticateByName", handlers.AuthenticateByName)
		authGroup.POST("/emby/Users/authenticatebyname", handlers.AuthenticateByName)

		authGroup.GET("/Users/Me", handlers.GetCurrentUser)
		authGroup.GET("/Users/:id", handlers.GetUserById)
		authGroup.GET("/emby/Users/:id", handlers.GetUserById)
	}

	// Vistas y Preferencias
	viewGroup := r.Group("/")
	{
		viewGroup.GET("/Users/:id/Views", handlers.GetUserViews)
		viewGroup.GET("/emby/Users/:id/Views", handlers.GetUserViews)
		viewGroup.GET("/DisplayPreferences/usersettings", handlers.GetDisplayPreferences)
		viewGroup.GET("/Users/:id/DisplayPreferences", handlers.GetDisplayPreferences)
	}

	// Biblioteca
	libraryGroup := r.Group("/")
	{
		libraryGroup.GET("/Library/VirtualFolders", handlers.GetVirtualFolders)
		libraryGroup.GET("/emby/Library/VirtualFolders", handlers.GetVirtualFolders)
		libraryGroup.GET("/Items", handlers.GetItems)
		libraryGroup.GET("/emby/Items", handlers.GetItems)
	}

	// Reproducción
	playbackGroup := r.Group("/")
	{
		playbackGroup.GET("/Items/:id/PlaybackInfo", handlers.GetPlaybackInfo)
		playbackGroup.GET("/Videos/:id/stream", handlers.StreamVideo)
		playbackGroup.POST("/Sessions/Playing", handlers.ReportPlaying)
		playbackGroup.POST("/Sessions/Playing/Progress", handlers.ReportPlayingProgress)
		playbackGroup.POST("/Sessions/Playing/Stopped", handlers.ReportPlaying)
	}

	// Otros
	r.NoRoute(func(c *gin.Context) {
		log.Printf("[404] No encontrado: %s %s", c.Request.Method, c.Request.URL.Path)
		c.JSON(404, gin.H{"error": "Endpoint not implemented", "path": c.Request.URL.Path})
	})
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Gellyte API Server is running", "version": "10.11.8"})
	})
	r.GET("/favicon.ico", func(c *gin.Context) { c.Status(204) })
	r.GET("/QuickConnect/Enabled", handlers.GetQuickConnectEnabled)
	r.GET("/Branding/Configuration", handlers.GetBrandingConfiguration)
	r.GET("/System/Endpoint", handlers.GetEndpointInfo)
	r.GET("/Playback/BitrateTest", handlers.GetBitrateTest)
	r.POST("/Sessions/Capabilities/Full", handlers.PostCapabilities)
	r.GET("/socket", handlers.GetDummySocket)
	r.GET("/emby/socket", handlers.GetDummySocket)

	log.Println("Gellyte server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}

func seedDatabase() {
	// Comprueba si el usuario admin existe
	var user models.User
	err := database.DB.Where("username = ?", "admin").First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// El usuario Admin no existe, lo creamos
			log.Println("Usuario admin no encontrado, creando uno...")
			adminUser := models.User{
				ID:       handlers.AdminUUID,
				Username: "admin",
				Password: "admin",
				IsAdmin:  true,
			}
			if err := database.DB.Create(&adminUser).Error; err != nil {
				log.Fatalf("Fallo al crear el usuario admin: %v", err)
			}
			log.Println("Usuario admin creado con éxito.")
		} else {
			log.Fatalf("Fallo al comprobar la existencia del usuario admin: %v", err)
		}
	}
}
