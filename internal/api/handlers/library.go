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
	if startIndex == 0 {
		startIndex, _ = strconv.Atoi(c.Query("startIndex"))
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("Limit", "50"))
	if limit == 50 && c.Query("limit") != "" {
		limit, _ = strconv.Atoi(c.Query("limit"))
	}

	parentId := c.Query("ParentId")
	if parentId == "" {
		parentId = c.Query("parentId")
	}
	itemTypes := c.Query("IncludeItemTypes")
	if itemTypes == "" {
		itemTypes = c.Query("includeItemTypes")
	}
	searchTerm := c.Query("SearchTerm")
	if searchTerm == "" {
		searchTerm = c.Query("searchTerm")
	}

	query := database.DB.Model(&models.MediaItem{})

	// Búsqueda por nombre
	if searchTerm != "" {
		query = query.Where("name LIKE ?", "%"+searchTerm+"%")
	}

	// Filtrado básico por ParentId (ID de la carpeta virtual o carpeta física)
	if parentId == "12345678-1234-1234-1234-123456789012" {
		query = query.Where("type = ?", "Movie")
	} else if parentId == "22345678-1234-1234-1234-123456789012" {
		query = query.Where("type = ?", "Series")
	} else if parentId != "" {
		// Navegación jerárquica: devolvemos los hijos directos del ParentId
		query = query.Where("parent_id = ?", parentId)
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
	imageType := c.Param("imageType")

	if imageType == "Thumb" {
		thumbPath := filepath.Join(dir, "thumb.jpg")
		if _, err := os.Stat(thumbPath); err == nil {
			c.File(thumbPath)
			return
		}
	}

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
	if limit == 20 && c.Query("limit") != "" {
		limit, _ = strconv.Atoi(c.Query("limit"))
	}
	parentId := c.Query("ParentId")
	if parentId == "" {
		parentId = c.Query("parentId")
	}
	itemTypes := c.Query("IncludeItemTypes")
	if itemTypes == "" {
		itemTypes = c.Query("includeItemTypes")
	}
	
	query := database.DB.Model(&models.MediaItem{})

	if parentId == "12345678-1234-1234-1234-123456789012" {
		query = query.Where("type = ?", "Movie")
	} else if parentId == "22345678-1234-1234-1234-123456789012" {
		query = query.Where("type = ?", "Episode")
	} else if itemTypes != "" {
		types := strings.Split(itemTypes, ",")
		query = query.Where("type IN ?", types)
	}

	var dbItems []models.MediaItem
	query.Order("created_at desc").Limit(limit).Find(&dbItems)

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
	parentId := ""
	if item.Type == "Movie" {
		parentId = "12345678-1234-1234-1234-123456789012"
	} else if item.Type == "Episode" {
		parentId = "22345678-1234-1234-1234-123456789012"
	}

	isFolder := item.Type == "Series" || item.Type == "Season" || item.Type == "CollectionFolder" || item.Type == "Folder"

	dto := BaseItemDto{
		Name:                    item.Name,
		Id:                      item.ID,
		ServerId:                ServerUUID,
		Type:                    item.Type,
		MediaType:               "Video",
		IsFolder:                isFolder,
		PlayAccess:              "Full",
		Path:                    item.Path,
		ParentId:                parentId,
		RunTimeTicks:            item.RunTimeTicks,
		Width:                   item.Width,
		Height:                  item.Height,
		ProductionYear:          item.ProductionYear,
		PrimaryImageAspectRatio: 0.66,
		Overview:                item.Overview,
		ImageTags:               make(map[string]string),
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
	// Añadir Thumbnail si existe
	if _, err := os.Stat(filepath.Join(dir, "thumb.jpg")); err == nil {
		dto.ImageTags["Thumb"] = "thumb-tag"
	}

	return dto
}
