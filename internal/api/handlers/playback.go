package handlers

import (
	"fmt"
	"net/http"

	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gin-gonic/gin"
)

// GetPlaybackInfo godoc
// @Summary Obtener información de reproducción
// @Description Determina si el archivo se puede reproducir directamente.
// @Tags Playback
// @Produce json
// @Param id path string true "ID del item"
// @Success 200 {object} map[string]interface{}
// @Router /Items/{id}/PlaybackInfo [get]
// GetPlaybackInfo godoc
func GetPlaybackInfo(c *gin.Context) {
	id := c.Param("id")
	
	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item no encontrado"})
		return
	}

	// Por ahora simulamos que todo es Direct Play para máxima fluidez en el MVP extendido.
	// Sin embargo, preparamos la estructura para que apps como Swiftfin no den error.
	
	streamUrl := fmt.Sprintf("/Videos/%s/stream", item.ID)

	c.JSON(http.StatusOK, gin.H{
		"MediaSources": []gin.H{
			{
				"Protocol": "Http",
				"Id":       item.ID,
				"Path":     item.Path,
				"Type":     "Default",
				"Container": item.Container,
				"Size":      item.Size,
				"Name":      item.Name,
				"IsRemote":  false,
				"SupportsDirectPlay": true,
				"SupportsDirectStream": true,
				"SupportsTranscoding": true, // Decimos que sí para que la app lo considere
				"SupportsResume":      true,
				"DirectStreamUrl":     streamUrl,
				"RunTimeTicks":        item.RunTimeTicks,
				"Bitrate":             item.Bitrate,
				"MediaStreams": []gin.H{
					{
						"Type": "Video",
						"Codec": item.VideoCodec,
						"IsInterlaced": false,
						"Width":  item.Width,
						"Height": item.Height,
						"BitRate": item.Bitrate,
					},
					{
						"Type": "Audio",
						"Codec": item.AudioCodec,
						"IsDefault": true,
					},
				},
			},
		},
		"PlaySessionId": "79b3602d385642d99723ecdbf6a4773a",
	})
}

// StreamVideo godoc
// @Summary Stream de video (Direct Play)
// @Description Sirve el archivo de video con soporte de Byte-Ranges (HTTP 206).
// @Tags Playback
// @Param id path string true "ID del item"
// @Router /Videos/{id}/stream [get]
func StreamVideo(c *gin.Context) {
	id := c.Param("id")
	
	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video no encontrado"})
		return
	}

	// Gin maneja automáticamente los Byte-Ranges si usamos c.File()
	// Esto es extremadamente eficiente en RAM porque no carga el archivo entero.
	c.File(item.Path)
}

// ReportPlaying godoc
// @Summary Reportar inicio/progreso de reproducción
// @Description Sincroniza el estado de reproducción con el servidor.
// @Tags Playback
// @Success 204 "No Content"
// @Router /Sessions/Playing [post]
func ReportPlaying(c *gin.Context) {
	// Por ahora solo aceptamos el reporte para que la app no de error.
	c.Status(http.StatusNoContent)
}

// ReportPlayingProgress godoc
// @Summary Reportar progreso de reproducción
// @Tags Playback
// @Router /Sessions/Playing/Progress [post]
func ReportPlayingProgress(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
