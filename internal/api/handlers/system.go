package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetPublicInfo godoc
func GetPublicInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"LocalAddress":           fmt.Sprintf("http://%s", c.Request.Host),
		"ServerName":             "Gellyte",
		"Version":                "10.11.8",
		"ProductName":            "Jellyfin Server",
		"OperatingSystem":        "Linux",
		"Id":                     "83e4c49d-9273-4556-9a5d-4952011702f3",
		"StartupWizardCompleted": true,
	})
}

// GetSystemInfo godoc
func GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"SystemUpdateLevel":          "Available",
		"OperatingSystem":            "Linux",
		"ServerName":                 "Gellyte",
		"Version":                    "10.11.8",
		"ServerVersion":              "10.11.8",
		"Id":                         "83e4c49d-9273-4556-9a5d-4952011702f3",
		"HasUpdateAvailable":         false,
		"CanSelfRestart":             false,
		"CanSelfUpdate":              false,
		"WebSocketPortNumber":        8080,
		"SupportsHttps":              false,
		"SupportsLibraryUninstall":   false,
		"HasPendingRestart":          false,
		"IsShuttingDown":             false,
		"SupportsPatcher":            false,
		"CompletedInstallations":     []string{},
		"CanLaunchWebBrowser":        false,
		"HardwareAccelerationDrivers": []string{},
		"HasToken":                   true,
		"EncoderLocation":            "None",
	})
}

// PostCapabilities godoc
func PostCapabilities(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// GetBitrateTest devuelve datos aleatorios para que la app mida la velocidad.
func GetBitrateTest(c *gin.Context) {
	size := 500000 // 500kb
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

// GetDummySocket simplemente acepta la conexión pero no hace nada por ahora.
func GetDummySocket(c *gin.Context) {
	// Jellyfin espera un upgrade a WebSocket. Por ahora devolvemos un error controlado
	// para que la app pase a modo polling si es necesario.
	c.String(http.StatusNotImplemented, "WebSocket not implemented yet")
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
	// Devolvemos un objeto vacío, ya que no tenemos branding personalizado.
	c.JSON(http.StatusOK, gin.H{})
}
