package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
		{
			"Name":           "Series",
			"Locations":      []string{"./media/series"},
			"CollectionType": "tvshows",
			"ItemId":         "22345678-1234-1234-1234-123456789012",
		},
	}
	c.JSON(http.StatusOK, folders)
}

func GetItems(c *gin.Context) {
	startIndex, _ := strconv.Atoi(c.DefaultQuery("StartIndex", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("Limit", "50"))
	parentId := c.Query("ParentId")
	itemTypes := c.Query("IncludeItemTypes")

	query := database.DB.Model(&models.MediaItem{})

	// Filtrado básico por ParentId (ID de la carpeta virtual)
	if parentId == "12345678-1234-1234-1234-123456789012" {
		query = query.Where("type = ?", "Movie")
	} else if parentId == "22345678-1234-1234-1234-123456789012" {
		// En la vista de "Series", el cliente suele pedir Series, no episodios directos.
		// Pero como en este MVP extendido estamos aplanando, devolveremos episodios si se pide explícitamente.
		if strings.Contains(itemTypes, "Movie") {
			query = query.Where("type = ?", "Movie")
		} else if strings.Contains(itemTypes, "Episode") {
			query = query.Where("type = ?", "Episode")
		} else {
			query = query.Where("type = ?", "Episode")
		}
	} else if itemTypes != "" {
		// Filtrado por tipos solicitados (estándar Jellyfin)
		types := strings.Split(itemTypes, ",")
		query = query.Where("type IN ?", types)
	}

	var dbItems []models.MediaItem
	var total int64

	query.Count(&total)
	query.Offset(startIndex).Limit(limit).Find(&dbItems)

	// Convertimos a BaseItemDto para cumplir con el esquema Jellyfin
	respItems := []BaseItemDto{}
	for _, item := range dbItems {
		respItems = append(respItems, mapToDto(item))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": total,
	})
}

// GetItemImage devuelve la imagen asociada a un item.
func GetItemImage(c *gin.Context) {
	id := c.Param("id")
	
	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	dir := filepath.Dir(item.Path)
	imgPath := findImage(dir, item.Name)
	if imgPath != "" {
		c.File(imgPath)
		return
	}

	c.Status(http.StatusNotFound)
}

// Helpers para imágenes
func hasImage(dir, itemName string) bool {
	return findImage(dir, itemName) != ""
}

func findImage(dir, itemName string) string {
	names := []string{"poster.jpg", "folder.jpg", "cover.jpg", itemName + ".jpg", "poster.png", "folder.png"}
	for _, n := range names {
		path := filepath.Join(dir, n)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// GetResumeItems devuelve items para continuar viendo (vacío por ahora).
func GetResumeItems(c *gin.Context) {
	c.JSON(http.StatusOK, BaseItemDtoQueryResult{
		Items:            []BaseItemDto{},
		TotalRecordCount: 0,
		StartIndex:       0,
	})
}

// GetItemDetails devuelve los detalles de un item específico o carpeta virtual.
func GetItemDetails(c *gin.Context) {
	id := c.Param("id")

	// Manejo de carpetas virtuales (hardcoded por ahora)
	if id == "12345678-1234-1234-1234-123456789012" {
		c.JSON(http.StatusOK, gin.H{
			"Name":           "Películas",
			"Id":             id,
			"Type":           "CollectionFolder",
			"CollectionType": "movies",
			"IsFolder":       true,
		})
		return
	}
	if id == "22345678-1234-1234-1234-123456789012" {
		c.JSON(http.StatusOK, gin.H{
			"Name":           "Series",
			"Id":             id,
			"Type":           "CollectionFolder",
			"CollectionType": "tvshows",
			"IsFolder":       true,
		})
		return
	}

	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, mapToDto(item))
}

// GetNextUp devuelve el siguiente episodio para ver (vacío por ahora).
func GetNextUp(c *gin.Context) {
	c.JSON(http.StatusOK, BaseItemDtoQueryResult{
		Items:            []BaseItemDto{},
		TotalRecordCount: 0,
		StartIndex:       0,
	})
}

// GetLatestItems devuelve los últimos archivos añadidos.
func GetLatestItems(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("Limit", "20"))
	parentId := c.Query("ParentId")
	itemTypes := c.Query("IncludeItemTypes")
	
	query := database.DB.Order("created_at desc").Limit(limit)

	if parentId == "12345678-1234-1234-1234-123456789012" {
		query = query.Where("type = ?", "Movie")
	} else if parentId == "22345678-1234-1234-1234-123456789012" {
		query = query.Where("type = ?", "Episode")
	} else if itemTypes != "" {
		types := strings.Split(itemTypes, ",")
		query = query.Where("type IN ?", types)
	}

	var dbItems []models.MediaItem
	query.Find(&dbItems)

	respItems := []BaseItemDto{}
	for _, item := range dbItems {
		respItems = append(respItems, mapToDto(item))
	}

	c.JSON(http.StatusOK, respItems)
}

// GetSuggestions devuelve sugerencias para el usuario (vacío por ahora).
func GetSuggestions(c *gin.Context) {
	c.JSON(http.StatusOK, BaseItemDtoQueryResult{
		Items:            []BaseItemDto{},
		TotalRecordCount: 0,
		StartIndex:       0,
	})
}

// mapToDto convierte un modelo MediaItem a BaseItemDto.
func mapToDto(item models.MediaItem) BaseItemDto {
	dto := BaseItemDto{
		Name:       item.Name,
		Id:         item.ID,
		ServerId:   ServerUUID,
		Type:       item.Type,
		MediaType:  "Video",
		IsFolder:   false,
		PlayAccess: "Full",
		Path:       item.Path,
		ImageTags:  make(map[string]string),
		UserData: UserItemDataDto{
			PlaybackPositionTicks: 0,
			PlayCount:             0,
			IsFavorite:            false,
			Played:                false,
		},
	}

	dir := filepath.Dir(item.Path)
	if hasImage(dir, item.Name) {
		dto.ImageTags["Primary"] = "fixed-tag"
	}

	return dto
}
