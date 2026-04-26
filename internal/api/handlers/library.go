package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/services"
	"github.com/gin-gonic/gin"
)

// GetVirtualFolders godoc
func (h *Handler) GetVirtualFolders(c *gin.Context) {
	libOptions := LibraryOptions{
		Enabled:               true,
		EnableRealtimeMonitor: true,
		SaveLocalMetadata:     true,
		PathInfos:             []PathInfo{},
		TypeOptions:           []TypeOptions{},
	}

	folders := []VirtualFolderDto{
		{
			"Películas",
			[]string{config.AppConfig.Library.MoviesPath},
			"movies",
			libOptions,
			config.AppConfig.Jellyfin.MoviesLibraryID,
			config.AppConfig.Jellyfin.MoviesLibraryID,
			nil,
			nil,
		},
		{
			"Series",
			[]string{config.AppConfig.Library.SeriesPath},
			"tvshows",
			libOptions,
			config.AppConfig.Jellyfin.SeriesLibraryID,
			config.AppConfig.Jellyfin.SeriesLibraryID,
			nil,
			nil,
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
	moviesLibNorm := strings.ReplaceAll(config.AppConfig.Jellyfin.MoviesLibraryID, "-", "")
	seriesLibNorm := strings.ReplaceAll(config.AppConfig.Jellyfin.SeriesLibraryID, "-", "")

	actualParentID := parentId
	switch parentId {
	case moviesLibNorm:
		itemTypes = []string{"Movie"}
		actualParentID = ""
	case seriesLibNorm:
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
		userId = config.AppConfig.Jellyfin.AdminUUID
	}

	respItems := []BaseItemDto{}
	for _, item := range dbItems {
		respItems = append(respItems, h.mapToDto(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": total,
		"StartIndex":       startIndex,
	})
}

func (h *Handler) GetItemDetails(c *gin.Context) {
	id := c.Param("id")

	// Verificar si es una carpeta virtual (biblioteca raíz)
	moviesLibNorm := strings.ReplaceAll(config.AppConfig.Jellyfin.MoviesLibraryID, "-", "")
	seriesLibNorm := strings.ReplaceAll(config.AppConfig.Jellyfin.SeriesLibraryID, "-", "")
	idNorm := strings.ReplaceAll(id, "-", "")

	switch idNorm {
	case moviesLibNorm:
		c.JSON(http.StatusOK, gin.H{
			"Name":           "Películas",
			"Id":             config.AppConfig.Jellyfin.MoviesLibraryID,
			"Type":           "CollectionFolder",
			"CollectionType": "movies",
		})
		return
	case seriesLibNorm:
		c.JSON(http.StatusOK, gin.H{
			"Name":           "Series",
			"Id":             config.AppConfig.Jellyfin.SeriesLibraryID,
			"Type":           "CollectionFolder",
			"CollectionType": "tvshows",
		})
		return
	}

	item, err := h.LibraryService.GetItem(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item no encontrado"})
		return
	}

	userId := c.GetString("UserID")
	if userId == "" {
		userId = config.AppConfig.Jellyfin.AdminUUID
	}

	c.JSON(http.StatusOK, h.mapToDto(*item, userId))
}

func (h *Handler) GetItemImage(c *gin.Context) {
	id := c.Param("id")
	item, err := h.LibraryService.GetItem(id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// Buscar thumb.jpg en la misma carpeta que el archivo de video
	dir := filepath.Dir(item.Path)
	thumbPath := filepath.Join(dir, "thumb.jpg")

	if _, err := os.Stat(thumbPath); err == nil {
		c.File(thumbPath)
		return
	}

	c.Status(http.StatusNotFound)
}

func (h *Handler) GetUserPrimaryImage(c *gin.Context) {
	// Jellyfin envía SVGs simples o imágenes para el avatar del usuario
	svg := `<svg width="100" height="100" xmlns="http://www.w3.org/2000/svg"><rect width="100" height="100" fill="#00a5ff"/><text x="50%" y="50%" font-family="Arial" font-size="40" fill="white" text-anchor="middle" dy=".3em">G</text></svg>`
	c.Header("Content-Type", "image/svg+xml")
	c.String(http.StatusOK, svg)
}

func (h *Handler) GetUserImage(c *gin.Context) {
	// Alias para GetUserPrimaryImage usado por algunos clientes
	h.GetUserPrimaryImage(c)
}

func (h *Handler) GetNextUp(c *gin.Context) {
	userId := c.GetString("UserID")
	if userId == "" {
		userId = config.AppConfig.Jellyfin.AdminUUID
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("Limit", "24"))
	items, err := h.LibraryService.GetNextUpItems(userId, limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Items": []interface{}{}, "TotalRecordCount": 0})
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

func (h *Handler) GetResumeItems(c *gin.Context) {
	userId := c.GetString("UserID")
	if userId == "" {
		userId = config.AppConfig.Jellyfin.AdminUUID
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("Limit", "24"))
	items, err := h.LibraryService.GetResumeItems(userId, limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Items": []interface{}{}, "TotalRecordCount": 0})
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

func (h *Handler) GetSuggestions(c *gin.Context) {
	userId := c.Query("userId")
	if userId == "" {
		userId = config.AppConfig.Jellyfin.AdminUUID
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("Limit", "12"))
	items, err := h.LibraryService.GetSuggestions(userId, limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Items": []interface{}{}, "TotalRecordCount": 0})
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

// GetLatestItems godoc
func (h *Handler) GetLatestItems(c *gin.Context) {
	userId := c.GetString("UserID")
	if userId == "" {
		userId = config.AppConfig.Jellyfin.AdminUUID
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("Limit", "16"))
	itemTypesStr := c.Query("IncludeItemTypes")
	var itemTypes []string
	if itemTypesStr != "" {
		itemTypes = strings.Split(itemTypesStr, ",")
	}

	items, err := h.LibraryService.GetLatestItems(limit, itemTypes)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	respItems := []BaseItemDto{}
	for _, item := range items {
		respItems = append(respItems, h.mapToDto(item, userId))
	}

	c.JSON(http.StatusOK, respItems)
}

func (h *Handler) GetSpecialFeatures(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func (h *Handler) GetAncestors(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func (h *Handler) GetSimilarItems(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Items": []interface{}{}, "TotalRecordCount": 0})
}

func (h *Handler) GetMediaSegments(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Items": []interface{}{}, "TotalRecordCount": 0})
}

// mapToDto convierte un modelo de BD a un objeto de respuesta compatible con Jellyfin.
func (h *Handler) mapToDto(item models.MediaItem, userId string) BaseItemDto {
	// Obtener datos extras del usuario (progreso, favoritos, etc)
	userData, _ := h.LibraryService.GetUserData(userId, item.ID)
	
	userDataDto := UserItemDataDto{
		IsFavorite: false,
		Played:     false,
	}
	if userData != nil {
		userDataDto = UserItemDataDto{
			PlaybackPositionTicks: userData.PlaybackPositionTicks,
			PlayCount:             userData.PlayCount,
			IsFavorite:            userData.IsFavorite,
			Played:                userData.Played,
			LastPlayedDate:        userData.LastPlayedDate.Format(time.RFC3339),
		}
	}

	dto := BaseItemDto{
		Name:         item.Name,
		Id:           item.ID,
		ServerId:     config.AppConfig.Jellyfin.ServerUUID,
		Type:         item.Type,
		RunTimeTicks: item.RunTimeTicks,
		IsFolder:     item.Type == "Series" || item.Type == "Season" || item.Type == "Folder" || item.Type == "CollectionFolder",
		ImageTags: map[string]string{
			"Primary": "tag", // Dummy tag para activar carga de imágenes en el cliente
		},
		UserData: userDataDto,
	}

	if item.ProductionYear > 0 {
		dto.ProductionYear = item.ProductionYear
	}
	if item.Overview != "" {
		dto.Overview = item.Overview
	}

	// Si es un episodio, añadir Season/Series data
	if item.Type == "Episode" {
		dto.IndexNumber = item.IndexNumber
		dto.ParentIndexNumber = item.ParentIndexNumber
		dto.SeriesName = item.SeriesName
		dto.SeriesId = item.SeriesID
		dto.SeasonId = item.ParentID
		dto.SeasonName = item.SeasonName
	}

	return dto
}
