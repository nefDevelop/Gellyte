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
	ServerVersion   = "10.11.8"
	ProductName     = "Jellyfin"
	OperatingSystem = "Linux"
)

// GetPublicInfo godoc
func (h *SystemHandler) GetPublicInfo(c *gin.Context) {
	c.Header("Content-Type", "application/json; profile=\"PascalCase\"; charset=utf-8")
	sId := strings.ReplaceAll(config.AppConfig.Jellyfin.ServerUUID, "-", "")
	c.JSON(http.StatusOK, PublicSystemInfo{
		LocalAddress:           fmt.Sprintf("http://%s", c.Request.Host),
		ServerName:             config.AppConfig.Server.Name,
		Version:                ServerVersion,
		ProductName:            ProductName,
		OperatingSystem:        OperatingSystem,
		Id:                     sId,
		StartupWizardCompleted: true,
	})
}

// GetSystemInfo godoc
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	c.Header("Content-Type", "application/json; profile=\"PascalCase\"; charset=utf-8")
	sId := strings.ReplaceAll(config.AppConfig.Jellyfin.ServerUUID, "-", "")
	c.JSON(http.StatusOK, SystemInfo{
		LocalAddress:               fmt.Sprintf("http://%s", c.Request.Host),
		ServerName:                 config.AppConfig.Server.Name,
		Version:                    ServerVersion,
		ProductName:                ProductName,
		OperatingSystem:            OperatingSystem,
		Id:                         sId,
		StartupWizardCompleted:     true,
		OperatingSystemDisplayName: "Linux",
		PackageName:                "Jellyfin",
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
func (h *SystemHandler) PostCapabilities(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// dummyBitrateData reserva 500KB estáticos de memoria que se reutilizan para todas las pruebas de velocidad.
var dummyBitrateData = make([]byte, 500000)

// GetBitrateTest devuelve datos aleatorios para que la app mida la velocidad.
func (h *SystemHandler) GetBitrateTest(c *gin.Context) {
	c.Data(http.StatusOK, "application/octet-stream", dummyBitrateData)
}

// GetEndpointInfo devuelve la URL base del servidor.
func (h *SystemHandler) GetEndpointInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"IsLocal": true,
		"Address": fmt.Sprintf("http://%s", c.Request.Host),
	})
}

// GetQuickConnectEnabled godoc
func (h *SystemHandler) GetQuickConnectEnabled(c *gin.Context) {
	c.JSON(http.StatusOK, false)
}

// InitiateQuickConnect godoc
func (h *SystemHandler) InitiateQuickConnect(c *gin.Context) {
	// Según el esquema OpenAPI de Jellyfin, 401 indica que QuickConnect no está activo.
	// Algunos clientes esperan una estructura de objeto base incluso con 401.
	c.JSON(http.StatusUnauthorized, QuickConnectResult{
		Authenticated: false,
		Secret:        "",
		Code:          "",
		DeviceId:      "",
		DeviceName:    "",
		AppName:       "",
		AppVersion:    "",
		DateAdded:     "2001-01-01T00:00:00Z",
	})
}

// GetBrandingConfiguration godoc
func (h *SystemHandler) GetBrandingConfiguration(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"LoginDisclaimer":     "",
		"CustomCss":           "",
		"SplashscreenEnabled": false,
	})
}

// GetPingSystem godoc
func (h *SystemHandler) GetPingSystem(c *gin.Context) {
	c.String(http.StatusOK, ProductName)
}

func (h *SystemHandler) GetPlugins(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func (h *SystemHandler) GetScheduledTasks(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func (h *SystemHandler) GetPackages(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func (h *SystemHandler) GetActivityLogEntries(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

func (h *SystemHandler) GetStreamyfinConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"features": []string{},
	})
}

func (h *SystemHandler) DeleteStreamyfinDevice(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
