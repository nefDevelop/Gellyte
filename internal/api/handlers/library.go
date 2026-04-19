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

	userId := c.GetString("UserID")
	if userId == "" {
		userId = "53896590-3b41-46a4-9591-96b054a8e3f6"
	}

	// Convertimos a BaseItemDto para cumplir con el esquema Jellyfin
	respItems := []BaseItemDto{}
	for _, item := range dbItems {
		respItems = append(respItems, mapToDto(item, userId))
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

// GetUserPrimaryImage handles requests for user primary images.
func GetUserPrimaryImage(c *gin.Context) {
	// For now, return a 404 as we don't have user images implemented.
	// In a real scenario, you would fetch the user's image based on the ID.
	c.Status(http.StatusNotFound)
}

// GetSpecialFeatures handles requests for special features.
func GetSpecialFeatures(c *gin.Context) {
	// For now, return an empty result as special features are not implemented.
	c.JSON(http.StatusOK, gin.H{
		"Items":            []BaseItemDto{},
		"TotalRecordCount": 0,
	})
}

// GetAncestors handles requests for item ancestors.
func GetAncestors(c *gin.Context) {
	// For now, return an empty result as ancestors are not fully implemented.
	c.JSON(http.StatusOK, []BaseItemDto{})
}

// GetSimilarItems handles requests for similar items.
func GetSimilarItems(c *gin.Context) {
	// For now, return an empty result as similar items logic is not implemented.
	c.JSON(http.StatusOK, gin.H{
		"Items":            []BaseItemDto{},
		"TotalRecordCount": 0,
	})
}

// GetMediaSegments handles requests for media segments.
func GetMediaSegments(c *gin.Context) {
	// For now, return an empty result as media segments are not implemented.
	// This endpoint is often used for trickplay or other advanced streaming features.
	c.JSON(http.StatusOK, gin.H{
		"Items":            []BaseItemDto{},
		"TotalRecordCount": 0,
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

	userId := c.GetString("UserID")
	if userId == "" {
		userId = "53896590-3b41-46a4-9591-96b054a8e3f6"
	}

	c.JSON(http.StatusOK, mapToDto(item, userId))
}

// GetNextUp devuelve el siguiente episodio disponible para ver en las series activas del usuario.
func GetNextUp(c *gin.Context) {
	userId := c.GetString("UserID")
	if userId == "" {
		userId = "53896590-3b41-46a4-9591-96b054a8e3f6"
	}

	// 1. Encontrar los últimos episodios reproducidos de cada serie
	var playedEpisodes []models.UserItemData
	database.DB.Where("user_id = ? AND played = ?", userId, true).Order("last_played_date desc").Find(&playedEpisodes)

	seenSeries := make(map[string]bool)
	var nextUpItems []BaseItemDto

	for _, ud := range playedEpisodes {
		var lastEpisode models.MediaItem
		if err := database.DB.Where("id = ? AND type = ?", ud.MediaItemID, "Episode").First(&lastEpisode).Error; err != nil {
			continue
		}

		// Obtener la serie padre para no repetir
		seriesId := ""
		var parent models.MediaItem
		database.DB.Where("id = ?", lastEpisode.ParentID).First(&parent)
		if parent.Type == "Season" {
			seriesId = parent.ParentID
		} else {
			seriesId = parent.ID
		}

		if seenSeries[seriesId] {
			continue
		}
		seenSeries[seriesId] = true

		// 2. Buscar el "siguiente" episodio en la base de datos
		// El siguiente episodio debería estar en la misma temporada (ParentID) y tener un nombre alfabéticamente posterior
		var nextEpisode models.MediaItem
		err := database.DB.Where("parent_id = ? AND name > ? AND type = ?", lastEpisode.ParentID, lastEpisode.Name, "Episode").Order("name asc").First(&nextEpisode).Error

		if err == nil {
			// Comprobar si el siguiente ya ha sido visto
			var nextData models.UserItemData
			database.DB.Where("user_id = ? AND media_item_id = ?", userId, nextEpisode.ID).First(&nextData)
			if !nextData.Played {
				nextUpItems = append(nextUpItems, mapToDto(nextEpisode, userId))
			}
		} else {
			// TODO: Si no hay más en esta temporada, buscar el primer episodio de la siguiente temporada
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            nextUpItems,
		"TotalRecordCount": len(nextUpItems),
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

	userId := c.GetString("UserID")
	if userId == "" {
		userId = "53896590-3b41-46a4-9591-96b054a8e3f6"
	}

	respItems := []BaseItemDto{}
	for _, item := range dbItems {
		respItems = append(respItems, mapToDto(item, userId))
	}

	c.JSON(http.StatusOK, respItems)
}

// GetResumeItems devuelve items con progreso pendiente (Continuar viendo).
func GetResumeItems(c *gin.Context) {
	userId := c.GetString("UserID")
	if userId == "" {
		userId = "53896590-3b41-46a4-9591-96b054a8e3f6"
	}

	var userDatas []models.UserItemData
	database.DB.Where("user_id = ? AND playback_position_ticks > 0 AND played = false", userId).Order("last_played_date desc").Find(&userDatas)

	respItems := []BaseItemDto{}
	for _, ud := range userDatas {
		var item models.MediaItem
		if err := database.DB.Where("id = ?", ud.MediaItemID).First(&item).Error; err == nil {
			respItems = append(respItems, mapToDto(item, userId))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": len(respItems),
	})
}

// GetSuggestions devuelve sugerencias para el usuario (usamos los últimos añadidos por ahora).
func GetSuggestions(c *gin.Context) {
	userId := c.GetString("UserID")

	var items []models.MediaItem
	database.DB.Order("created_at desc").Limit(10).Find(&items)

	respItems := []BaseItemDto{}
	for _, item := range items {
		respItems = append(respItems, mapToDto(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": len(respItems),
	})
}

// mapToDto convierte un modelo MediaItem a BaseItemDto.
func mapToDto(item models.MediaItem, userId string) BaseItemDto {
	parentId := ""
	if item.Type == "Movie" {
		parentId = "12345678-1234-1234-1234-123456789012"
	} else if item.Type == "Episode" {
		parentId = "22345678-1234-1234-1234-123456789012"
	}

	isFolder := item.Type == "Series" || item.Type == "Season" || item.Type == "CollectionFolder" || item.Type == "Folder"

	userData := UserItemDataDto{
		PlaybackPositionTicks: 0,
		PlayCount:             0,
		IsFavorite:            false,
		Played:                false,
	}

	// Cargar datos reales del usuario si existe sesión
	if userId != "" {
		var dbUserData models.UserItemData
		if err := database.DB.Where("user_id = ? AND media_item_id = ?", userId, item.ID).First(&dbUserData).Error; err == nil {
			userData.PlaybackPositionTicks = dbUserData.PlaybackPositionTicks
			userData.PlayCount = dbUserData.PlayCount
			userData.IsFavorite = dbUserData.IsFavorite
			userData.Played = dbUserData.Played
		}
	}

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
		UserData:                userData,
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
