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
func GetVirtualFolders(c *gin.Context) {
	folders := []gin.H{
		{
			"Name":           "Películas",
			"Locations":      []string{"./media/peliculas"},
			"CollectionType": "movies",
			"ItemId":         MoviesLibraryID,
		},
		{
			"Name":           "Series",
			"Locations":      []string{"./media/series"},
			"CollectionType": "tvshows",
			"ItemId":         SeriesLibraryID,
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
	parentId = strings.ReplaceAll(parentId, "-", "") // Normalizar UUIDs de Jellyfin
	itemTypes := c.Query("IncludeItemTypes")
	if itemTypes == "" {
		itemTypes = c.Query("includeItemTypes")
	}
	searchTerm := c.Query("SearchTerm")
	if searchTerm == "" {
		searchTerm = c.Query("searchTerm")
	}
	ids := c.Query("ids")
	if ids == "" {
		ids = c.Query("Ids")
	}

	query := database.DB.Model(&models.MediaItem{})

	// Filtrado directo por IDs específicos (prioritario para el cliente web)
	if ids != "" {
		idList := strings.Split(ids, ",")
		query = query.Where("id IN ?", idList)
	}

	// Búsqueda por nombre
	if searchTerm != "" {
		query = query.Where("name LIKE ?", "%"+searchTerm+"%")
	}

	// Filtrado básico por ParentId (ID de la carpeta virtual o carpeta física)
	moviesLibNorm := strings.ReplaceAll(MoviesLibraryID, "-", "")
	seriesLibNorm := strings.ReplaceAll(SeriesLibraryID, "-", "")

	if parentId == moviesLibNorm {
		query = query.Where("type = ?", "Movie")
	} else if parentId == seriesLibNorm {
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
		userId = AdminUUID
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

	// Evitar consultas a la base de datos si el cliente envía IDs inválidos de JavaScript
	if id == "" || id == "undefined" || id == "null" {
		c.Status(http.StatusNotFound)
		return
	}

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
	userId := c.Param("id")
	if userId == "" {
		c.Status(http.StatusNotFound)
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	initial := "?"
	if len(user.Username) > 0 {
		initial = strings.ToUpper(string([]rune(user.Username)[0]))
	}

	bgColor := "#8e44ad"

	svg := `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200" viewBox="0 0 200 200">
  <circle cx="100" cy="100" r="100" fill="` + bgColor + `"/>
  <text x="100" y="135" font-family="Arial, Helvetica, sans-serif" font-size="100" fill="white" font-weight="bold" text-anchor="middle">` + initial + `</text>
</svg>`

	c.Data(http.StatusOK, "image/svg+xml", []byte(svg))
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
	id := c.Param("id")
	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.JSON(http.StatusOK, []BaseItemDto{})
		return
	}

	userId := c.GetString("UserID")
	if userId == "" {
		userId = AdminUUID
	}

	ancestors := []BaseItemDto{}
	
	// Add parent if exists
	if item.ParentID != "" {
		var parent models.MediaItem
		if err := database.DB.Where("id = ?", item.ParentID).First(&parent).Error; err == nil {
			ancestors = append(ancestors, mapToDto(parent, userId))
		}
	}

	c.JSON(http.StatusOK, ancestors)
}

// GetSimilarItems handles requests for similar items.
func GetSimilarItems(c *gin.Context) {
	id := c.Param("id")
	var item models.MediaItem
	if err := database.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"Items": []BaseItemDto{}, "TotalRecordCount": 0})
		return
	}

	var similar []models.MediaItem
	database.DB.Where("type = ? AND id != ?", item.Type, item.ID).Limit(12).Find(&similar)

	userId := c.GetString("UserID")
	if userId == "" {
		userId = AdminUUID
	}

	respItems := []BaseItemDto{}
	for _, s := range similar {
		respItems = append(respItems, mapToDto(s, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": len(respItems),
	})
}

// GetMediaSegments handles requests for media segments.
func GetMediaSegments(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Items":            []BaseItemDto{},
		"TotalRecordCount": 0,
	})
}

// GetItemDetails devuelve los detalles de un item específico o carpeta virtual.
func GetItemDetails(c *gin.Context) {
	id := c.Param("id")

	// Evitar IDs inválidos
	if id == "" || id == "undefined" || id == "null" {
		c.Status(http.StatusNotFound)
		return
	}

	idNormalized := strings.ReplaceAll(id, "-", "")
	moviesLibNorm := strings.ReplaceAll(MoviesLibraryID, "-", "")
	seriesLibNorm := strings.ReplaceAll(SeriesLibraryID, "-", "")

	// Manejo de carpetas virtuales (hardcoded por ahora)
	if idNormalized == moviesLibNorm {
		c.JSON(http.StatusOK, gin.H{
			"Name":           "Películas",
			"Id":             MoviesLibraryID,
			"Type":           "CollectionFolder",
			"CollectionType": "movies",
			"IsFolder":       true,
		})
		return
	}
	if idNormalized == seriesLibNorm {
		c.JSON(http.StatusOK, gin.H{
			"Name":           "Series",
			"Id":             SeriesLibraryID,
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
		userId = AdminUUID
	}

	c.JSON(http.StatusOK, mapToDto(item, userId))
}

// GetNextUp devuelve el siguiente episodio disponible para ver en las series activas del usuario.
func GetNextUp(c *gin.Context) {
	userId := c.GetString("UserID")
	if userId == "" {
		userId = AdminUUID
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
		var nextEpisode models.MediaItem
		err := database.DB.Where("parent_id = ? AND name > ? AND type = ?", lastEpisode.ParentID, lastEpisode.Name, "Episode").Order("name asc").First(&nextEpisode).Error

		if err == nil {
			// Comprobar si el siguiente ya ha sido visto
			var nextData models.UserItemData
			database.DB.Where("user_id = ? AND media_item_id = ?", userId, nextEpisode.ID).First(&nextData)
			if !nextData.Played {
				nextUpItems = append(nextUpItems, mapToDto(nextEpisode, userId))
			}
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
	parentId = strings.ReplaceAll(parentId, "-", "") // Normalizar UUIDs de Jellyfin
	itemTypes := c.Query("IncludeItemTypes")
	if itemTypes == "" {
		itemTypes = c.Query("includeItemTypes")
	}

	query := database.DB.Model(&models.MediaItem{})

	moviesLibNorm := strings.ReplaceAll(MoviesLibraryID, "-", "")
	seriesLibNorm := strings.ReplaceAll(SeriesLibraryID, "-", "")

	if parentId == moviesLibNorm {
		query = query.Where("type = ?", "Movie")
	} else if parentId == seriesLibNorm {
		query = query.Where("type = ?", "Episode")
	} else if itemTypes != "" {
		types := strings.Split(itemTypes, ",")
		query = query.Where("type IN ?", types)
	}

	var dbItems []models.MediaItem
	query.Order("created_at desc").Limit(limit).Find(&dbItems)

	userId := c.GetString("UserID")
	if userId == "" {
		userId = AdminUUID
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
		userId = AdminUUID
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

// GetSuggestions devuelve sugerencias para el usuario.
func GetSuggestions(c *gin.Context) {
	userId := c.Query("userId")
	if userId == "" {
		userId = AdminUUID
	}

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
		parentId = MoviesLibraryID
	} else if item.Type == "Episode" {
		parentId = SeriesLibraryID
	}

	isFolder := item.Type == "Series" || item.Type == "Season" || item.Type == "CollectionFolder" || item.Type == "Folder"

	userData := UserItemDataDto{
		PlaybackPositionTicks: 0,
		PlayCount:             0,
		IsFavorite:            false,
		Played:                false,
	}

	if userId != "" {
		var dbUserData models.UserItemData
		result := database.DB.Where("user_id = ? AND media_item_id = ?", userId, item.ID).Find(&dbUserData)
		if result.Error == nil && result.RowsAffected > 0 {
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
	if _, err := os.Stat(filepath.Join(dir, "thumb.jpg")); err == nil {
		dto.ImageTags["Thumb"] = "thumb-tag"
	}

	if item.RunTimeTicks > 0 {
		dto.MediaSources = []interface{}{
			gin.H{
				"Id":           item.ID,
				"Protocol":     "Http",
				"Container":    item.Container,
				"RunTimeTicks": item.RunTimeTicks,
				"Bitrate":      item.Bitrate,
				"MediaStreams": []gin.H{
					{
						"Type":   "Video",
						"Codec":  item.VideoCodec,
						"Width":  item.Width,
						"Height": item.Height,
					},
					{
						"Type":  "Audio",
						"Codec": item.AudioCodec,
					},
				},
			},
		}
	} else {
		dto.MediaSources = []interface{}{}
	}

	return dto
}
