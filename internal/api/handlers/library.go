package handlers

import (
	"net/http"
	"strconv"

	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gin-gonic/gin"
)

// GetVirtualFolders godoc
// @Summary Listar carpetas virtuales
// @Description Devuelve las bibliotecas (Películas, Series, etc.) configuradas.
// @Tags Library
// @Produce json
// @Success 200 {object} []map[string]interface{}
// @Router /Library/VirtualFolders [get]
func GetVirtualFolders(c *gin.Context) {
	folders := []gin.H{
		{
			"Name":           "Películas",
			"Locations":      []string{"./media/peliculas"},
			"CollectionType": "movies",
			"ItemId":         "12345678-1234-1234-1234-123456789012", // Debe coincidir con auth.go
		},
	}
	c.JSON(http.StatusOK, folders)
}

// GetItems godoc
// @Summary Listar items de la biblioteca
// @Description Devuelve los contenidos de una biblioteca con filtros y paginación.
// @Tags Library
// @Produce json
// @Param StartIndex query int false "Índice de inicio"
// @Param Limit query int false "Límite de resultados"
// @Success 200 {object} map[string]interface{}
// @Router /Items [get]
func GetItems(c *gin.Context) {
	startIndex, _ := strconv.Atoi(c.DefaultQuery("StartIndex", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("Limit", "50"))

	var items []models.MediaItem
	var total int64

	// Consultar la base de datos con paginación
	database.DB.Model(&models.MediaItem{}).Count(&total)
	database.DB.Offset(startIndex).Limit(limit).Find(&items)

	// Jellyfin espera una estructura específica
	c.JSON(http.StatusOK, gin.H{
		"Items":            items,
		"TotalRecordCount": total,
	})
}
