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
			"ItemId":         "12345678-1234-1234-1234-123456789012",
		},
	}
	c.JSON(http.StatusOK, folders)
}

func GetItems(c *gin.Context) {
	startIndex, _ := strconv.Atoi(c.DefaultQuery("StartIndex", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("Limit", "50"))

	var items []models.MediaItem
	var total int64

	database.DB.Model(&models.MediaItem{}).Count(&total)
	database.DB.Offset(startIndex).Limit(limit).Find(&items)

	c.JSON(http.StatusOK, gin.H{
		"Items":            items,
		"TotalRecordCount": total,
	})
}
