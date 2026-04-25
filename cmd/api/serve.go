package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/gellyte/gellyte/docs"
	"github.com/gellyte/gellyte/internal/api/discovery"
	"github.com/gellyte/gellyte/internal/api/handlers"
	"github.com/gellyte/gellyte/internal/api/middleware"
	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/library"
	"github.com/gellyte/gellyte/internal/repository"
	"github.com/gellyte/gellyte/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Inicia el servidor API de Gellyte",
	Run: func(cmd *cobra.Command, args []string) {
		runServe()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe() {
	// Iniciar Configuración
	config.InitConfig()

	// Iniciar Base de Datos
	database.InitDB()

	// Iniciar Repositorios
	userRepo := repository.NewUserRepository(database.DB)
	mediaRepo := repository.NewMediaRepository(database.DB)
	userDataRepo := repository.NewUserItemDataRepository(database.DB)

	// Iniciar Servicios
	authService := services.NewAuthService(userRepo)
	libraryService := services.NewLibraryService(mediaRepo, userDataRepo)
	playbackService := services.NewPlaybackService(mediaRepo, userDataRepo)

	// Iniciar Handlers
	h := handlers.NewHandler(authService, libraryService, playbackService)

	// Iniciar el Hub de WebSockets
	go handlers.GlobalHub.Run()

	library.OnLibraryChanged = handlers.NotifyLibraryChanged

	go func() {
		ssdp := discovery.SSDPServer{Port: config.AppConfig.Server.Port, ServerID: config.AppConfig.Jellyfin.ServerUUID}
		ssdp.Start()
	}()

	go library.WatchFolder(config.AppConfig.Library.MoviesPath, "movies")
	go library.WatchFolder(config.AppConfig.Library.SeriesPath, "series")

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.ResponseLoggerMiddleware())
	r.Use(middleware.EmbyAuthMiddleware())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// --- RUTAS COMPATIBLES ---
	r.GET("/System/Info/Public", h.GetPublicInfo)
	r.GET("/system/info/public", h.GetPublicInfo)
	r.GET("/emby/System/Info/Public", h.GetPublicInfo)
	r.GET("/System/Info", h.GetSystemInfo)
	r.GET("/system/info", h.GetSystemInfo)
	r.GET("/emby/System/Info", h.GetSystemInfo)
	r.GET("/System/Ping", h.GetPingSystem)
	r.GET("/system/ping", h.GetPingSystem)
	r.GET("/Moonfin/Ping", h.GetPingSystem)
	r.GET("/moonfin/ping", h.GetPingSystem)

	r.GET("/Users/Public", h.GetPublicUsers)
	r.GET("/users/public", h.GetPublicUsers)
	r.GET("/emby/Users/Public", h.GetPublicUsers)
	r.GET("/Users", h.GetPublicUsers)
	r.GET("/users", h.GetPublicUsers)
	r.GET("/emby/Users", h.GetPublicUsers)

	r.POST("/Users/AuthenticateByName", h.AuthenticateByName)
	r.POST("/users/authenticatebyname", h.AuthenticateByName)
	r.POST("/emby/Users/AuthenticateByName", h.AuthenticateByName)
	r.POST("/Users/:id", h.GetUserById)
	r.GET("/Users/:id", h.GetUserById)
	r.GET("/users/:id", h.GetUserById)
	r.GET("/Users/Me", h.GetCurrentUser)
	r.GET("/users/me", h.GetCurrentUser)

	r.GET("/Users/:id/Views", h.GetUserViews)
	r.GET("/users/:id/views", h.GetUserViews)
	r.GET("/DisplayPreferences/usersettings", h.GetDisplayPreferences)
	r.GET("/displaypreferences/usersettings", h.GetDisplayPreferences)

	r.GET("/Library/VirtualFolders", h.GetVirtualFolders)
	r.GET("/library/virtualfolders", h.GetVirtualFolders)
	r.GET("/Library/MediaFolders", h.GetMediaFolders)
	r.GET("/library/mediafolders", h.GetMediaFolders)
	r.GET("/Library/PhysicalPaths", h.GetPhysicalPaths)
	r.GET("/library/physicalpaths", h.GetPhysicalPaths)
	r.GET("/Items", h.GetItems)
	r.GET("/items", h.GetItems)
	r.GET("/Items/Counts", h.GetItemsCounts)
	r.GET("/items/counts", h.GetItemsCounts)
	r.GET("/Items/Filters", h.GetItemsFilters)
	r.GET("/items/filters", h.GetItemsFilters)
	r.GET("/Items/Filters2", h.GetItemsFilters)
	r.GET("/items/filters2", h.GetItemsFilters)
	r.GET("/Items/Root", h.GetItemsRoot)
	r.GET("/items/root", h.GetItemsRoot)
	r.GET("/Items/:id", h.GetItemDetails)
	r.GET("/items/:id", h.GetItemDetails)

	r.GET("/Items/:id/PlaybackInfo", h.GetPlaybackInfo)
	r.GET("/Videos/:id/stream", h.StreamVideo)
	r.POST("/Items/:id/PlaybackInfo", h.GetPlaybackInfo)
	r.GET("/Items/:id/Images/:imageType", h.GetItemImage)
	r.GET("/items/:id/images/:imageType", h.GetItemImage)
	r.POST("/Sessions/Playing", h.ReportPlaying)
	r.POST("/Sessions/Playing/Progress", h.ReportPlayingProgress)
	r.POST("/Sessions/Playing/Stopped", h.ReportPlayingStopped)
	r.GET("/Sessions", h.GetSessions)
	r.GET("/sessions", h.GetSessions)
	r.GET("/emby/Sessions", h.GetSessions)
	r.GET("/Videos/:id/:mediaSourceId/Subtitles/:index/0/Stream.vtt", h.GetSubtitleStream)
	r.GET("/Videos/:id/:mediaSourceId/Subtitles/:index/Stream.vtt", h.GetSubtitleStream)

	r.GET("/Videos/:id/main.m3u8", h.GetHlsPlaylist)
	r.GET("/Videos/:id/hls/:segmentId/stream.ts", h.GetHlsSegment)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "Endpoint not implemented", "path": c.Request.URL.Path})
	})
	r.Match([]string{"GET", "HEAD"}, "/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Gellyte API Server is running", "version": config.AppConfig.Server.Name})
	})
	r.GET("/favicon.ico", func(c *gin.Context) { c.Status(204) })
	r.GET("/QuickConnect/Enabled", h.GetQuickConnectEnabled)
	r.GET("/quickconnect/enabled", h.GetQuickConnectEnabled)
	r.GET("/Branding/Configuration", h.GetBrandingConfiguration)
	r.GET("/branding/configuration", h.GetBrandingConfiguration)
	r.GET("/System/Endpoint", h.GetEndpointInfo)
	r.GET("/Playback/BitrateTest", h.GetBitrateTest)
	r.POST("/Sessions/Capabilities", h.PostCapabilities)
	r.POST("/sessions/capabilities", h.PostCapabilities)
	r.POST("/Sessions/Capabilities/Full", h.PostCapabilities)
	
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/socket", h.GetDummySocket)
	r.GET("/UserViews", h.GetUserViews)
	r.GET("/userviews", h.GetUserViews)
	r.GET("/UserViews/GroupingOptions", h.GetGroupingOptions)
	r.GET("/userviews/groupingoptions", h.GetGroupingOptions)
	r.GET("/Shows/NextUp", h.GetNextUp)
	r.GET("/shows/nextup", h.GetNextUp)
	r.GET("/Shows/:id/Episodes", h.GetShowEpisodes)
	r.GET("/Shows/:id/Seasons", h.GetShowSeasons)
	r.GET("/UserItems/Resume", h.GetResumeItems)
	r.GET("/Items/:id/SpecialFeatures", h.GetSpecialFeatures)
	r.GET("/Items/:id/Ancestors", h.GetAncestors)
	r.GET("/Items/:id/Similar", h.GetSimilarItems)
	r.GET("/MediaSegments/:id", h.GetMediaSegments)
	r.GET("/Users/:id/Images/Primary", h.GetUserPrimaryImage)
	r.GET("/useritems/resume", h.GetResumeItems)
	r.GET("/Items/Latest", h.GetLatestItems)
	r.GET("/Items/Suggestions", h.GetSuggestions)
	r.GET("/api/ws/dashboard", h.GetDummySocket)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.AppConfig.Server.Port),
		Handler: r,
	}

	go func() {
		log.Printf("Gellyte server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Esperar señal de interrupción
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Cerrando el servidor Gellyte...")

	// Contexto para el apagado (5 segundos de gracia)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	// Detener otros servicios
	library.StopScanner()

	log.Println("Servidor finalizado de forma limpia.")
}
