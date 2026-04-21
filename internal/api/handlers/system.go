package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Constantes para evitar valores mágicos y facilitar la configuración.
const (
	serverName      = "Gellyte"
	serverVersion   = "10.11.8"
	productName     = "Jellyfin Server"
	operatingSystem = "Linux"
	encoderLocation = "/usr/bin/ffmpeg"
)

// GetPublicInfo godoc
func GetPublicInfo(c *gin.Context) {
	c.JSON(http.StatusOK, PublicSystemInfo{
		LocalAddress:           fmt.Sprintf("http://%s", c.Request.Host),
		ServerName:             serverName,
		Version:                serverVersion,
		ProductName:            productName,
		OperatingSystem:        operatingSystem,
		Id:                     ServerUUID,
		StartupWizardCompleted: true,
	})
}

// GetSystemInfo godoc
func GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, SystemInfo{
		SystemUpdateLevel:           "None",
		OperatingSystem:             operatingSystem,
		ServerName:                  serverName,
		Version:                     serverVersion,
		ServerVersion:               serverVersion,
		Id:                          ServerUUID,
		HasUpdateAvailable:          false,
		CanSelfRestart:              false,
		CanSelfUpdate:               false,
		WebSocketPortNumber:         8081, // TODO: Debería obtenerse de la configuración
		SupportsHttps:               false,
		SupportsLibraryUninstall:    false,
		HasPendingRestart:           false,
		IsShuttingDown:              false,
		SupportsPatcher:             false,
		CompletedInstallations:      []string{},
		CanLaunchWebBrowser:         false,
		HardwareAccelerationDrivers: []string{},
		HasToken:                    true,
		EncoderLocation:             encoderLocation,
	})
}

// PostCapabilities godoc
func PostCapabilities(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// GetBitrateTest devuelve datos aleatorios para que la app mida la velocidad.
func GetBitrateTest(c *gin.Context) {
	// 500kb de datos para la prueba de velocidad.
	const size = 500000
	data := make([]byte, size)
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// GetEndpointInfo devuelve la URL base del servidor.
func GetEndpointInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"IsLocal": true,
		"Address": fmt.Sprintf("http://%s", c.Request.Host),
	})
}

// GetQuickConnectEnabled godoc
// @Summary Comprobar si QuickConnect está habilitado
// @Description Devuelve si la función QuickConnect está disponible.
// @Tags System
// @Produce json
// @Success 200 {boolean} boolean
// @Router /QuickConnect/Enabled [get]
func GetQuickConnectEnabled(c *gin.Context) {
	c.JSON(http.StatusOK, false)
}

// GetBrandingConfiguration godoc
// @Summary Obtener configuración de branding
// @Description Devuelve la configuración de branding del servidor.
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /Branding/Configuration [get]
func GetBrandingConfiguration(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"LoginDisclaimer":     "",
		"CustomCss":           "",
		"SplashscreenEnabled": false,
	})
}

// GetPingSystem godoc
func GetPingSystem(c *gin.Context) {
	// Jellyfin clients typically expect a string "Jellyfin Server" or a simple 200 response
	c.String(http.StatusOK, productName)
}
