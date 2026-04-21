package main

import (
	"log"
	"net/http"

	_ "github.com/gellyte/gellyte/docs"
	"github.com/gellyte/gellyte/internal/api/discovery"
	"github.com/gellyte/gellyte/internal/api/handlers"
	"github.com/gellyte/gellyte/internal/api/middleware"
	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/library"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Gellyte API
// @version 1.0
// @description Servidor compatible con Jellyfin escrito en Go.
// @host localhost:8080
// @BasePath /
func main() {
	database.InitDB()

	// Iniciar el Hub de WebSockets
	go handlers.GlobalHub.Run()

	library.OnLibraryChanged = handlers.NotifyLibraryChanged

	go func() {
		ssdp := discovery.SSDPServer{Port: 8081, ServerID: handlers.ServerUUID}
		ssdp.Start()
	}()

	go library.WatchFolder("./media/peliculas", "movies")
	go library.WatchFolder("./media/series", "series")

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.ResponseLoggerMiddleware())
	r.Use(middleware.EmbyAuthMiddleware())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// --- RUTAS COMPATIBLES (CON ALIAS) ---

	// Sistema
	r.GET("/System/Info/Public", handlers.GetPublicInfo)
	r.GET("/system/info/public", handlers.GetPublicInfo)
	r.GET("/emby/System/Info/Public", handlers.GetPublicInfo)
	r.GET("/System/Info", handlers.GetSystemInfo)
	r.GET("/system/info", handlers.GetSystemInfo)
	r.GET("/emby/System/Info", handlers.GetSystemInfo)
	r.GET("/System/Ping", handlers.GetPingSystem)
	r.GET("/system/ping", handlers.GetPingSystem)
	r.GET("/Moonfin/Ping", handlers.GetPingSystem)
	r.GET("/moonfin/ping", handlers.GetPingSystem)

	// Usuarios y Auth
	r.GET("/Users/Public", handlers.GetPublicUsers)
	r.GET("/users/public", handlers.GetPublicUsers)
	r.GET("/emby/Users/Public", handlers.GetPublicUsers)
	r.GET("/Users", handlers.GetPublicUsers)
	r.GET("/users", handlers.GetPublicUsers)
	r.GET("/emby/Users", handlers.GetPublicUsers)

	r.POST("/Users/AuthenticateByName", handlers.AuthenticateByName)
	r.POST("/users/authenticatebyname", handlers.AuthenticateByName)
	r.POST("/emby/Users/AuthenticateByName", handlers.AuthenticateByName)
	r.POST("/Users/:id", handlers.GetUserById)
	r.GET("/Users/:id", handlers.GetUserById)
	r.GET("/users/:id", handlers.GetUserById)
	r.GET("/Users/Me", handlers.GetCurrentUser)
	r.GET("/users/me", handlers.GetCurrentUser)

	// Vistas y Preferencias
	r.GET("/Users/:id/Views", handlers.GetUserViews)
	r.GET("/users/:id/views", handlers.GetUserViews)
	r.GET("/DisplayPreferences/usersettings", handlers.GetDisplayPreferences)
	r.GET("/displaypreferences/usersettings", handlers.GetDisplayPreferences)

	// Biblioteca
	r.GET("/Library/VirtualFolders", handlers.GetVirtualFolders)
	r.GET("/library/virtualfolders", handlers.GetVirtualFolders)
	r.GET("/Library/MediaFolders", handlers.GetMediaFolders)
	r.GET("/library/mediafolders", handlers.GetMediaFolders)
	r.GET("/Library/PhysicalPaths", handlers.GetPhysicalPaths)
	r.GET("/library/physicalpaths", handlers.GetPhysicalPaths)
	r.GET("/Items", handlers.GetItems)
	r.GET("/items", handlers.GetItems)
	r.GET("/Items/Counts", handlers.GetItemsCounts)
	r.GET("/items/counts", handlers.GetItemsCounts)
	r.GET("/Items/Filters", handlers.GetItemsFilters)
	r.GET("/items/filters", handlers.GetItemsFilters)
	r.GET("/Items/Filters2", handlers.GetItemsFilters)
	r.GET("/items/filters2", handlers.GetItemsFilters)
	r.GET("/Items/Root", handlers.GetItemsRoot)
	r.GET("/items/root", handlers.GetItemsRoot)
	r.GET("/Items/:id", handlers.GetItemDetails)
	r.GET("/items/:id", handlers.GetItemDetails)

	// Reproducción
	r.GET("/Items/:id/PlaybackInfo", handlers.GetPlaybackInfo)
	r.GET("/Videos/:id/stream", handlers.StreamVideo)
	r.POST("/Items/:id/PlaybackInfo", handlers.GetPlaybackInfo) // El cliente puede enviar POST para PlaybackInfo
	r.GET("/Items/:id/Images/:imageType", handlers.GetItemImage)
	r.GET("/items/:id/images/:imageType", handlers.GetItemImage)
	r.POST("/Sessions/Playing", handlers.ReportPlaying)
	r.POST("/Sessions/Playing/Progress", handlers.ReportPlayingProgress)
	r.POST("/Sessions/Playing/Stopped", handlers.ReportPlayingStopped)
	r.GET("/Sessions", handlers.GetSessions)
	r.GET("/sessions", handlers.GetSessions)
	r.GET("/emby/Sessions", handlers.GetSessions)
	r.GET("/Videos/:id/:mediaSourceId/Subtitles/:index/0/Stream.vtt", handlers.GetSubtitleStream)
	r.GET("/Videos/:id/:mediaSourceId/Subtitles/:index/Stream.vtt", handlers.GetSubtitleStream)

	// Otros
	r.GET("/Videos/:id/main.m3u8", handlers.GetHlsPlaylist)
	r.GET("/Videos/:id/hls/:segmentId/stream.ts", handlers.GetHlsSegment)

	r.NoRoute(func(c *gin.Context) {
		////log.Printf("[404] No encontrado: %s %s", c.Request.Method, c.Request.URL.Path)
		c.JSON(404, gin.H{"error": "Endpoint not implemented", "path": c.Request.URL.Path})
	})
	r.Match([]string{"GET", "HEAD"}, "/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Gellyte API Server is running", "version": "10.8.13"})
	})
	r.GET("/favicon.ico", func(c *gin.Context) { c.Status(204) })
	r.GET("/QuickConnect/Enabled", handlers.GetQuickConnectEnabled)
	r.GET("/quickconnect/enabled", handlers.GetQuickConnectEnabled)
	r.GET("/Branding/Configuration", handlers.GetBrandingConfiguration)
	r.GET("/branding/configuration", handlers.GetBrandingConfiguration)
	r.GET("/System/Endpoint", handlers.GetEndpointInfo)
	r.GET("/Playback/BitrateTest", handlers.GetBitrateTest)
	r.POST("/Sessions/Capabilities", handlers.PostCapabilities)
	r.POST("/sessions/capabilities", handlers.PostCapabilities)
	r.POST("/Sessions/Capabilities/Full", handlers.PostCapabilities)
	// Health check endpoints
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/Health", func(c *gin.Context) { c.Status(http.StatusOK) })      // Common casing variation
	r.GET("/emby/Health", func(c *gin.Context) { c.Status(http.StatusOK) }) // Emby specific health check
	r.GET("/socket", handlers.GetDummySocket)
	r.GET("/emby/socket", handlers.GetDummySocket)
	r.GET("/UserViews", handlers.GetUserViews)
	r.GET("/userviews", handlers.GetUserViews)
	r.GET("/UserViews/GroupingOptions", handlers.GetGroupingOptions)
	r.GET("/userviews/groupingoptions", handlers.GetGroupingOptions)
	r.GET("/Shows/NextUp", handlers.GetNextUp)
	r.GET("/shows/nextup", handlers.GetNextUp)
	r.GET("/Shows/:id/Episodes", handlers.GetShowEpisodes)
	r.GET("/shows/:id/episodes", handlers.GetShowEpisodes)
	r.GET("/Shows/:id/Seasons", handlers.GetShowSeasons)
	r.GET("/shows/:id/seasons", handlers.GetShowSeasons)
	r.GET("/UserItems/Resume", handlers.GetResumeItems)
	r.GET("/Items/:id/SpecialFeatures", handlers.GetSpecialFeatures) // Nuevo handler
	r.GET("/items/:id/specialfeatures", handlers.GetSpecialFeatures) // Nuevo handler
	r.GET("/Items/:id/Ancestors", handlers.GetAncestors)             // Nuevo handler
	r.GET("/items/:id/ancestors", handlers.GetAncestors)             // Nuevo handler
	r.GET("/Items/:id/Similar", handlers.GetSimilarItems)            // Nuevo handler
	r.GET("/items/:id/similar", handlers.GetSimilarItems)            // Nuevo handler
	r.GET("/MediaSegments/:id", handlers.GetMediaSegments)           // Nuevo handler para MediaSegments
	r.GET("/mediasegments/:id", handlers.GetMediaSegments)           // Alias para MediaSegments
	r.GET("/Users/:id/Images/Primary", handlers.GetUserPrimaryImage) // Nuevo handler
	r.GET("/users/:id/images/primary", handlers.GetUserPrimaryImage) // Nuevo handler
	r.GET("/useritems/resume", handlers.GetResumeItems)
	r.GET("/Items/Latest", handlers.GetLatestItems)
	r.GET("/items/latest", handlers.GetLatestItems)
	r.GET("/Items/Suggestions", handlers.GetSuggestions)
	r.GET("/items/suggestions", handlers.GetSuggestions)
	r.GET("/api/ws/dashboard", handlers.GetDummySocket) // Handle WebSocket dashboard requests
	r.GET("/Streamyfin/config", func(c *gin.Context) { c.JSON(200, gin.H{}) })

	log.Println("Gellyte server starting on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}
