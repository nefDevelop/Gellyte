package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gin-gonic/gin"
)

// Constantes para evitar valores mágicos y facilitar la configuración.
const (
	serverVersion   = "10.11.8"
	productName     = "Jellyfin Server"
	operatingSystem = "Linux"
)

// GetPublicInfo godoc
func (h *Handler) GetPublicInfo(c *gin.Context) {
	c.JSON(http.StatusOK, PublicSystemInfo{
		LocalAddress:           fmt.Sprintf("http://%s", c.Request.Host),
		ServerName:             config.AppConfig.Server.Name,
		Version:                serverVersion,
		ProductName:            productName,
		OperatingSystem:        operatingSystem,
		Id:                     strings.ReplaceAll(config.AppConfig.Jellyfin.ServerUUID, "-", ""),
		StartupWizardCompleted: true,
	})
}

// GetSystemInfo godoc
func (h *Handler) GetSystemInfo(c *gin.Context) {
	sId := strings.ReplaceAll(config.AppConfig.Jellyfin.ServerUUID, "-", "")
	c.JSON(http.StatusOK, SystemInfo{
		LocalAddress:               fmt.Sprintf("http://%s", c.Request.Host),
		ServerName:                 config.AppConfig.Server.Name,
		Version:                    serverVersion,
		ProductName:                productName,
		OperatingSystem:            operatingSystem,
		Id:                         sId,
		StartupWizardCompleted:     true,
		OperatingSystemDisplayName:  "Linux",
		PackageName:                "Gellyte",
		HasPendingRestart:          false,
		IsShuttingDown:             false,
		SupportsLibraryMonitor:     true,
		WebSocketPortNumber:        config.AppConfig.Server.Port,
		CompletedInstallations:     []interface{}{},
		CanSelfRestart:             true,
		CanLaunchWebBrowser:        false,
		ProgramDataPath:            "/var/lib/gellyte",
		WebPath:                    "/usr/share/gellyte/web",
		ItemsByNamePath:            "/var/lib/gellyte/items",
		CachePath:                  "/var/cache/gellyte",
		LogPath:                    "/var/log/gellyte",
		InternalMetadataPath:       "/var/lib/gellyte/metadata",
		TranscodingTempPath:        config.AppConfig.Transcoder.TempPath,
		CastReceiverApplications:   []interface{}{},
		HasUpdateAvailable:         false,
		EncoderLocation:            "System",
		SystemArchitecture:         "X64",
	})
}

// PostCapabilities godoc
func (h *Handler) PostCapabilities(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// GetBitrateTest devuelve datos aleatorios para que la app mida la velocidad.
func (h *Handler) GetBitrateTest(c *gin.Context) {
	// 500kb de datos para la prueba de velocidad.
	const size = 500000
	data := make([]byte, size)
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// GetEndpointInfo devuelve la URL base del servidor.
func (h *Handler) GetEndpointInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"IsLocal": true,
		"Address": fmt.Sprintf("http://%s", c.Request.Host),
	})
}

// GetQuickConnectEnabled godoc
func (h *Handler) GetQuickConnectEnabled(c *gin.Context) {
	c.JSON(http.StatusOK, false)
}

// InitiateQuickConnect godoc
func (h *Handler) InitiateQuickConnect(c *gin.Context) {
	// Según el esquema OpenAPI de Jellyfin, 401 indica que QuickConnect no está activo.
	c.Status(http.StatusUnauthorized)
}

// GetBrandingConfiguration godoc
func (h *Handler) GetBrandingConfiguration(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"LoginDisclaimer":     "",
		"CustomCss":           "",
		"SplashscreenEnabled": false,
	})
}

// GetPingSystem godoc
func (h *Handler) GetPingSystem(c *gin.Context) {
	c.String(http.StatusOK, productName)
}
