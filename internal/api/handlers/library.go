package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/services"
	"github.com/gin-gonic/gin"
)

// GetVirtualFolders godoc
func (h *Handler) GetVirtualFolders(c *gin.Context) {
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

func (h *Handler) GetItems(c *gin.Context) {
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
	itemTypesStr := c.Query("IncludeItemTypes")
	if itemTypesStr == "" {
		itemTypesStr = c.Query("includeItemTypes")
	}
	var itemTypes []string
	if itemTypesStr != "" {
		itemTypes = strings.Split(itemTypesStr, ",")
	}

	searchTerm := c.Query("SearchTerm")
	if searchTerm == "" {
		searchTerm = c.Query("searchTerm")
	}
	idsStr := c.Query("ids")
	if idsStr == "" {
		idsStr = c.Query("Ids")
	}
	var ids []string
	if idsStr != "" {
		ids = strings.Split(idsStr, ",")
	}

	// Lógica de carpetas virtuales trasladada del handler al servicio/params
	moviesLibNorm := strings.ReplaceAll(MoviesLibraryID, "-", "")
	seriesLibNorm := strings.ReplaceAll(SeriesLibraryID, "-", "")

	actualParentID := parentId
	if parentId == moviesLibNorm {
		itemTypes = []string{"Movie"}
		actualParentID = ""
	} else if parentId == seriesLibNorm {
		itemTypes = []string{"Series"}
		actualParentID = ""
	}

	dbItems, total, err := h.LibraryService.GetItems(services.GetItemsParams{
		ParentID:   actualParentID,
		ItemTypes:  itemTypes,
		SearchTerm: searchTerm,
		IDs:        ids,
		StartIndex: startIndex,
		Limit:      limit,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userId := c.GetString("UserID")
	if userId == "" {
		userId = AdminUUID
	}

	respItems := []BaseItemDto{}
	for _, item := range dbItems {
		respItems = append(respItems, h.mapToDto(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": total,
	})
}

// GetItemImage devuelve la imagen asociada a un item.
func (h *Handler) GetItemImage(c *gin.Context) {
	id := c.Param("id")

	if id == "" || id == "undefined" || id == "null" {
		c.Status(http.StatusNotFound)
		return
	}

	item, err := h.LibraryService.GetItem(id)
	if err != nil {
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

// Helpers para imágenes (pueden seguir siendo internos o movidos a utils)
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

func hasImage(dir, itemName string) bool {
	return findImage(dir, itemName) != ""
}

// GetUserPrimaryImage handles requests for user primary images.
func (h *Handler) GetUserPrimaryImage(c *gin.Context) {
	userId := c.Param("id")
	if userId == "" {
		c.Status(http.StatusNotFound)
		return
	}

	user, err := h.AuthService.GetUserByID(userId)
	if err != nil {
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
func (h *Handler) GetSpecialFeatures(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Items":            []BaseItemDto{},
		"TotalRecordCount": 0,
	})
}

// GetAncestors handles requests for item ancestors.
func (h *Handler) GetAncestors(c *gin.Context) {
	id := c.Param("id")
	item, err := h.LibraryService.GetItem(id)
	if err != nil {
		c.JSON(http.StatusOK, []BaseItemDto{})
		return
	}

	userId := c.GetString("UserID")
	if userId == "" {
		userId = AdminUUID
	}

	ancestors := []BaseItemDto{}
	
	if item.ParentID != "" {
		if parent, err := h.LibraryService.GetItem(item.ParentID); err == nil {
			ancestors = append(ancestors, h.mapToDto(*parent, userId))
		}
	}

	c.JSON(http.StatusOK, ancestors)
}

// GetSimilarItems handles requests for similar items.
func (h *Handler) GetSimilarItems(c *gin.Context) {
	id := c.Param("id")
	item, err := h.LibraryService.GetItem(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Items": []BaseItemDto{}, "TotalRecordCount": 0})
		return
	}

	similar, _, _ := h.LibraryService.GetItems(services.GetItemsParams{
		ItemTypes: []string{item.Type},
		Limit:     12,
	})

	userId := c.GetString("UserID")
	if userId == "" {
		userId = AdminUUID
	}

	respItems := []BaseItemDto{}
	for _, s := range similar {
		if s.ID != item.ID {
			respItems = append(respItems, h.mapToDto(s, userId))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": len(respItems),
	})
}

// GetMediaSegments handles requests for media segments.
func (h *Handler) GetMediaSegments(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Items":            []BaseItemDto{},
		"TotalRecordCount": 0,
	})
}

// GetItemDetails devuelve los detalles de un item específico o carpeta virtual.
func (h *Handler) GetItemDetails(c *gin.Context) {
	id := c.Param("id")

	if id == "" || id == "undefined" || id == "null" {
		c.Status(http.StatusNotFound)
		return
	}

	idNormalized := strings.ReplaceAll(id, "-", "")
	moviesLibNorm := strings.ReplaceAll(MoviesLibraryID, "-", "")
	seriesLibNorm := strings.ReplaceAll(SeriesLibraryID, "-", "")

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

	item, err := h.LibraryService.GetItem(id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	userId := c.GetString("UserID")
	if userId == "" {
		userId = AdminUUID
	}

	c.JSON(http.StatusOK, h.mapToDto(*item, userId))
}

// GetNextUp devuelve el siguiente episodio disponible para ver en las series activas del usuario.
func (h *Handler) GetNextUp(c *gin.Context) {
	userId := c.GetString("UserID")
	if userId == "" {
		userId = AdminUUID
	}

	nextUpItems, err := h.LibraryService.GetNextUpItems(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respItems := []BaseItemDto{}
	for _, item := range nextUpItems {
		respItems = append(respItems, h.mapToDto(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": len(respItems),
	})
}

// GetLatestItems devuelve los últimos archivos añadidos.
func (h *Handler) GetLatestItems(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("Limit", "20"))
	if limit == 20 && c.Query("limit") != "" {
		limit, _ = strconv.Atoi(c.Query("limit"))
	}
	parentId := c.Query("ParentId")
	if parentId == "" {
		parentId = c.Query("parentId")
	}
	parentId = strings.ReplaceAll(parentId, "-", "")

	itemTypesStr := c.Query("IncludeItemTypes")
	if itemTypesStr == "" {
		itemTypesStr = c.Query("includeItemTypes")
	}
	var itemTypes []string
	if itemTypesStr != "" {
		itemTypes = strings.Split(itemTypesStr, ",")
	}

	moviesLibNorm := strings.ReplaceAll(MoviesLibraryID, "-", "")
	seriesLibNorm := strings.ReplaceAll(SeriesLibraryID, "-", "")

	if parentId == moviesLibNorm {
		itemTypes = []string{"Movie"}
	} else if parentId == seriesLibNorm {
		itemTypes = []string{"Episode"}
	}

	dbItems, err := h.LibraryService.GetLatestItems(limit, itemTypes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userId := c.GetString("UserID")
	if userId == "" {
		userId = AdminUUID
	}

	respItems := []BaseItemDto{}
	for _, item := range dbItems {
		respItems = append(respItems, h.mapToDto(item, userId))
	}

	c.JSON(http.StatusOK, respItems)
}

// GetResumeItems devuelve items con progreso pendiente (Continuar viendo).
func (h *Handler) GetResumeItems(c *gin.Context) {
	userId := c.GetString("UserID")
	if userId == "" {
		userId = AdminUUID
	}

	items, err := h.LibraryService.GetResumeItems(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respItems := []BaseItemDto{}
	for _, item := range items {
		respItems = append(respItems, h.mapToDto(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": len(respItems),
	})
}

// GetSuggestions devuelve sugerencias para el usuario.
func (h *Handler) GetSuggestions(c *gin.Context) {
	userId := c.Query("userId")
	if userId == "" {
		userId = AdminUUID
	}

	items, err := h.LibraryService.GetLatestItems(10, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respItems := []BaseItemDto{}
	for _, item := range items {
		respItems = append(respItems, h.mapToDto(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": len(respItems),
	})
}

// mapToDto convierte un modelo MediaItem a BaseItemDto.
func (h *Handler) mapToDto(item models.MediaItem, userId string) BaseItemDto {
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
		if dbUserData, err := h.LibraryService.GetUserData(userId, item.ID); err == nil {
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
