package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"runtime"
	"runtime/debug"
	"time"

	_ "net/http/pprof"

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

	// Configurar Gin en modo Release para ahorrar memoria y mejorar performance
	gin.SetMode(gin.ReleaseMode)

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

	// Profiling (pprof) - Solo accesible internamente o vía debug
	r.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.ResponseLoggerMiddleware())
	r.Use(middleware.EmbyAuthMiddleware())

	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/swagger/doc.json", func(c *gin.Context) {
		c.File("./docs/swagger.json")
	})
	r.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "/doc.json" {
			c.File("./docs/swagger.json")
			return
		}
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<title>Gellyte API Documentation</title>
				<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
				<link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/favicon-32x32.png" sizes="32x32" />
			</head>
			<body>
				<div id="swagger-ui"></div>
				<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
				<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
				<script>
					window.onload = function() {
						const ui = SwaggerUIBundle({
							url: "/swagger/doc.json",
							dom_id: '#swagger-ui',
							deepLinking: true,
							presets: [
								SwaggerUIBundle.presets.apis,
								SwaggerUIStandalonePreset
							],
							layout: "StandaloneLayout"
						});
					};
				</script>
			</body>
			</html>
		`)
	})

	// --- RUTAS COMPATIBLES ---
	r.GET("/System/Info/Public", h.System.GetPublicInfo)
	r.GET("/system/info/public", h.System.GetPublicInfo)
	r.GET("/emby/System/Info/Public", h.System.GetPublicInfo)
	r.GET("/System/Info", h.System.GetSystemInfo)
	r.GET("/system/info", h.System.GetSystemInfo)
	r.GET("/emby/System/Info", h.System.GetSystemInfo)
	r.GET("/System/Ping", h.System.GetPingSystem)
	r.GET("/system/ping", h.System.GetPingSystem)
	r.GET("/Moonfin/Ping", h.System.GetPingSystem)
	r.GET("/moonfin/ping", h.System.GetPingSystem)

	r.GET("/Users/Public", h.Auth.GetPublicUsers)
	r.GET("/users/public", h.Auth.GetPublicUsers)
	r.GET("/emby/Users/Public", h.Auth.GetPublicUsers)
	r.GET("/Users", h.Auth.GetPublicUsers)
	r.GET("/users", h.Auth.GetPublicUsers)
	r.GET("/emby/Users", h.Auth.GetPublicUsers)

	r.POST("/Users/AuthenticateByName", h.Auth.AuthenticateByName)
	r.POST("/users/authenticatebyname", h.Auth.AuthenticateByName)
	r.POST("/emby/Users/AuthenticateByName", h.Auth.AuthenticateByName)
	r.POST("/Users/:id", h.Auth.GetUserById)
	r.GET("/Users/:id", h.Auth.GetUserById)
	r.GET("/users/:id", h.Auth.GetUserById)
	r.GET("/Users/Me", h.Auth.GetCurrentUser)
	r.GET("/users/me", h.Auth.GetCurrentUser)

	r.GET("/Users/:id/Views", h.Auth.GetUserViews)
	r.GET("/users/:id/views", h.Auth.GetUserViews)
	r.GET("/DisplayPreferences/usersettings", h.Auth.GetDisplayPreferences)
	r.GET("/displaypreferences/usersettings", h.Auth.GetDisplayPreferences)

	r.GET("/Library/VirtualFolders", h.Library.GetVirtualFolders)
	r.GET("/library/virtualfolders", h.Library.GetVirtualFolders)
	r.GET("/Library/MediaFolders", h.Library.GetMediaFolders)
	r.GET("/library/mediafolders", h.Library.GetMediaFolders)
	r.GET("/Library/PhysicalPaths", h.Library.GetPhysicalPaths)
	r.GET("/library/physicalpaths", h.Library.GetPhysicalPaths)
	r.GET("/Items", h.Library.GetItems)
	r.GET("/items", h.Library.GetItems)
	r.GET("/Items/Counts", h.Library.GetItemsCounts)
	r.GET("/items/counts", h.Library.GetItemsCounts)
	r.GET("/Items/Filters", h.Library.GetItemsFilters)
	r.GET("/items/filters", h.Library.GetItemsFilters)
	r.GET("/Items/Filters2", h.Library.GetItemsFilters)
	r.GET("/items/filters2", h.Library.GetItemsFilters)
	r.GET("/Items/Root", h.Library.GetItemsRoot)
	r.GET("/items/root", h.Library.GetItemsRoot)
	r.GET("/Items/:id", h.Library.GetItemDetails)
	r.GET("/items/:id", h.Library.GetItemDetails)

	r.GET("/Items/:id/PlaybackInfo", h.Playback.GetPlaybackInfo)
	r.GET("/Videos/:id/stream", h.Playback.StreamVideo)
	r.POST("/Items/:id/PlaybackInfo", h.Playback.GetPlaybackInfo)
	r.GET("/Items/:id/Images/:imageType", h.Library.GetItemImage)
	r.GET("/items/:id/images/:imageType", h.Library.GetItemImage)
	r.POST("/Sessions/Playing", h.Playback.ReportPlaying)
	r.POST("/Sessions/Playing/Progress", h.Playback.ReportPlayingProgress)
	r.POST("/Sessions/Playing/Stopped", h.Playback.ReportPlayingStopped)
	r.POST("/Sessions/Logout", h.Session.Logout)
	r.GET("/Sessions", h.Session.GetSessions)
	r.GET("/sessions", h.Session.GetSessions)
	r.GET("/emby/Sessions", h.Session.GetSessions)
	r.GET("/Videos/:id/:mediaSourceId/Subtitles/:index/0/Stream.vtt", h.Playback.GetSubtitleStream)
	r.GET("/Videos/:id/:mediaSourceId/Subtitles/:index/Stream.vtt", h.Playback.GetSubtitleStream)

	r.GET("/Videos/:id/main.m3u8", h.Playback.GetHlsPlaylist)
	r.GET("/Videos/:id/hls/:segmentId/stream.ts", h.Playback.GetHlsSegment)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "Endpoint not implemented", "path": c.Request.URL.Path})
	})
	r.Match([]string{"GET", "HEAD"}, "/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"Message": "Gellyte Server is running",
			"Version": handlers.ServerVersion,
		})
	})

	r.GET("/Streamyfin/config", h.System.GetStreamyfinConfig)
	r.DELETE("/Streamyfin/device/:id", h.System.DeleteStreamyfinDevice)
	r.GET("/favicon.ico", func(c *gin.Context) { c.Status(204) })
	r.GET("/QuickConnect/Enabled", h.System.GetQuickConnectEnabled)
	r.GET("/quickconnect/enabled", h.System.GetQuickConnectEnabled)
	r.POST("/QuickConnect/Initiate", h.System.InitiateQuickConnect)
	r.POST("/quickconnect/initiate", h.System.InitiateQuickConnect)
	r.GET("/Plugins", h.System.GetPlugins)
	r.GET("/ScheduledTasks", h.System.GetScheduledTasks)
	r.GET("/Packages", h.System.GetPackages)
	r.GET("/System/ActivityLog/Entries", h.System.GetActivityLogEntries)
	r.GET("/Branding/Configuration", h.System.GetBrandingConfiguration)
	r.GET("/branding/configuration", h.System.GetBrandingConfiguration)
	r.GET("/System/Endpoint", h.System.GetEndpointInfo)
	r.GET("/Playback/BitrateTest", h.System.GetBitrateTest)
	r.POST("/Sessions/Capabilities", h.System.PostCapabilities)
	r.POST("/sessions/capabilities", h.System.PostCapabilities)
	r.POST("/Sessions/Capabilities/Full", h.System.PostCapabilities)

	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/socket", h.WebSocket.GetDummySocket)
	r.GET("/UserViews", h.Auth.GetUserViews)
	r.GET("/userviews", h.Auth.GetUserViews)
	r.GET("/UserViews/GroupingOptions", h.Library.GetGroupingOptions)
	r.GET("/userviews/groupingoptions", h.Library.GetGroupingOptions)
	r.GET("/Shows/NextUp", h.Library.GetNextUp)
	r.GET("/shows/nextup", h.Library.GetNextUp)
	r.GET("/Shows/:id/Episodes", h.Library.GetShowEpisodes)
	r.GET("/Shows/:id/Seasons", h.Library.GetShowSeasons)
	r.GET("/UserItems/Resume", h.Library.GetResumeItems)
	r.GET("/Items/:id/SpecialFeatures", h.Library.GetSpecialFeatures)
	r.GET("/Items/:id/Ancestors", h.Library.GetAncestors)
	r.GET("/Items/:id/Similar", h.Library.GetSimilarItems)
	r.GET("/MediaSegments/:id", h.Library.GetMediaSegments)
	r.GET("/Users/:id/Images/Primary", h.Auth.GetUserPrimaryImage)
	r.GET("/UserImage", h.Auth.GetUserImage)
	r.GET("/userimage", h.Auth.GetUserImage)
	r.GET("/useritems/resume", h.Library.GetResumeItems)
	r.GET("/Items/Latest", h.Library.GetLatestItems)
	r.GET("/Items/Suggestions", h.Library.GetSuggestions)
	r.GET("/api/ws/dashboard", h.WebSocket.GetDummySocket)

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

	// Tarea en segundo plano para liberar memoria tras el arranque inicial
	go func() {
		// Esperar un tiempo prudencial para que termine el escaneo inicial
		time.Sleep(1 * time.Minute)
		log.Println("[Memory] Realizando limpieza de memoria post-arranque...")
		runtime.GC()
		debug.FreeOSMemory()
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
