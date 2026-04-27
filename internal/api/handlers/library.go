package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/services"
	"github.com/gellyte/gellyte/pkg/utils"
	"github.com/gin-gonic/gin"
)

// GetVirtualFolders godoc
func (h *LibraryHandler) GetVirtualFolders(c *gin.Context) {
	libOptions := LibraryOptions{
		Enabled:               true,
		EnableRealtimeMonitor: true,
		SaveLocalMetadata:     true,
		PathInfos:             []PathInfo{},
		TypeOptions:           []TypeOptions{},
	}

	moviesId := config.AppConfig.Jellyfin.MoviesLibraryID
	seriesId := config.AppConfig.Jellyfin.SeriesLibraryID

	folders := []VirtualFolderDto{
		{
			"Películas",
			[]string{config.AppConfig.Library.MoviesPath},
			"movies",
			libOptions,
			moviesId,
			moviesId,
			nil,
			nil,
		},
		{
			"Series",
			[]string{config.AppConfig.Library.SeriesPath},
			"tvshows",
			libOptions,
			seriesId,
			seriesId,
			nil,
			nil,
		},
	}
	c.JSON(http.StatusOK, folders)
}

func (h *LibraryHandler) GetItems(c *gin.Context) {
	startIndex, _ := strconv.Atoi(c.DefaultQuery("StartIndex", "0"))
	if startIndex == 0 {
		startIndex, _ = strconv.Atoi(c.Query("startIndex"))
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("Limit", "50"))
	if limit == 50 && c.Query("limit") != "" {
		limit, _ = strconv.Atoi(c.Query("limit"))
	}

	userId := c.Param("id")
	if userId == "" {
		userId = c.GetString("UserID")
	}
	if userId == "" {
		userId = config.AppConfig.Jellyfin.AdminUUID
	}

	parentId := utils.NormalizeID(c.Query("ParentId"))
	if parentId == "" {
		parentId = utils.NormalizeID(c.Query("parentId"))
	}
	// No normalizar aquí para mantener los guiones si vienen en la query
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
		for i, id := range ids {
			ids[i] = utils.NormalizeID(id)
		}
	}

	// Lógica de carpetas virtuales trasladada del handler al servicio/params
	moviesLib := utils.NormalizeID(config.AppConfig.Jellyfin.MoviesLibraryID)
	seriesLib := utils.NormalizeID(config.AppConfig.Jellyfin.SeriesLibraryID)

	actualParentID := parentId
	switch parentId {
	case moviesLib:
		itemTypes = []string{"Movie"}
		actualParentID = ""
	case seriesLib:
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

	respItems := make([]BaseItemDto, 0, len(dbItems))
	for _, item := range dbItems {
		respItems = append(respItems, h.mapItem(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": total,
		"StartIndex":       startIndex,
	})
}

func (h *LibraryHandler) GetItemDetails(c *gin.Context) {
	id := utils.NormalizeID(c.Param("id"))

	// Verificar si es una carpeta virtual (biblioteca raíz)
	moviesLib := utils.NormalizeID(config.AppConfig.Jellyfin.MoviesLibraryID)
	seriesLib := utils.NormalizeID(config.AppConfig.Jellyfin.SeriesLibraryID)

	switch id {
	case moviesLib:
		c.JSON(http.StatusOK, gin.H{
			"Name":           "Películas",
			"Id":             config.AppConfig.Jellyfin.MoviesLibraryID,
			"Type":           "CollectionFolder",
			"CollectionType": "movies",
		})
		return
	case seriesLib:
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

	c.JSON(http.StatusOK, h.mapItem(*item, userId))
}

func (h *LibraryHandler) GetItemImage(c *gin.Context) {
	id := utils.NormalizeID(c.Param("id"))

	// Si es una carpeta virtual (biblioteca raíz), devolvemos un placeholder temático
	moviesLib := utils.NormalizeID(config.AppConfig.Jellyfin.MoviesLibraryID)
	seriesLib := utils.NormalizeID(config.AppConfig.Jellyfin.SeriesLibraryID)

	if id == moviesLib || id == seriesLib {
		color := "#00a5ff" // Azul Jellyfin
		text := "CINE"
		if id == seriesLib {
			text = "TV"
		}
		placeholder := fmt.Sprintf(`<svg width="200" height="300" xmlns="http://www.w3.org/2000/svg"><rect width="200" height="300" fill="%s"/><text x="50%" y="50%" font-family="Arial" font-size="40" font-weight="bold" fill="white" text-anchor="middle" dy=".3em">%s</text></svg>`, color, text)
		c.Header("Content-Type", "image/svg+xml")
		c.Header("Cache-Control", "public, max-age=31536000")
		c.String(http.StatusOK, placeholder)
		return
	}

	item, err := h.LibraryService.GetItem(id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	dir := filepath.Dir(item.Path)
	filename := filepath.Base(item.Path)
	nameWithoutExt := filename[:len(filename)-len(filepath.Ext(filename))]

	// Lista de nombres comunes para carátulas (Primary)
	// Jellyfin busca estos archivos en el mismo directorio que el video
	possibleNames := []string{
		"poster.jpg", "poster.png", "poster.jpeg",
		"folder.jpg", "folder.png", "folder.jpeg",
		"cover.jpg", "cover.png", "cover.jpeg",
		"default.jpg", "default.png",
		nameWithoutExt + "-poster.jpg",
		nameWithoutExt + "-poster.png",
		nameWithoutExt + ".jpg",
		nameWithoutExt + ".png",
		nameWithoutExt + ".jpeg",
	}

	for _, name := range possibleNames {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			c.Header("Cache-Control", "public, max-age=31536000")
			c.File(p)
			return
		}
	}

	// Si es una serie, también buscamos en la carpeta superior por si acaso
	if item.Type == "Series" || item.Type == "Episode" {
		parentDir := filepath.Dir(dir)
		for _, name := range possibleNames {
			p := filepath.Join(parentDir, name)
			if _, err := os.Stat(p); err == nil {
				c.Header("Cache-Control", "public, max-age=31536000")
				c.File(p)
				return
			}
		}
	}

	// Si no hay imagen, devolvemos un placeholder SVG con la inicial del nombre
	initial := "M"
	if len(item.Name) > 0 {
		initial = string(item.Name[0])
	}
	placeholder := fmt.Sprintf(`<svg width="200" height="300" xmlns="http://www.w3.org/2000/svg"><rect width="200" height="300" fill="#222"/><text x="50%" y="50%" font-family="Arial" font-size="60" fill="#444" text-anchor="middle" dy=".3em">%s</text></svg>`, initial)
	c.Header("Content-Type", "image/svg+xml")
	c.String(http.StatusOK, placeholder)
}

func (h *LibraryHandler) GetNextUp(c *gin.Context) {
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

	respItems := make([]BaseItemDto, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, h.mapItem(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": len(respItems),
	})
}

func (h *LibraryHandler) GetResumeItems(c *gin.Context) {
	userId := c.Param("id")
	if userId == "" {
		userId = c.GetString("UserID")
	}
	if userId == "" {
		userId = config.AppConfig.Jellyfin.AdminUUID
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("Limit", "24"))
	items, err := h.LibraryService.GetResumeItems(userId, limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Items": []interface{}{}, "TotalRecordCount": 0, "StartIndex": 0})
		return
	}

	respItems := make([]BaseItemDto, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, h.mapItem(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": len(respItems),
		"StartIndex":       0,
	})
}

func (h *LibraryHandler) GetSuggestions(c *gin.Context) {
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

	respItems := make([]BaseItemDto, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, h.mapItem(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": len(respItems),
	})
}

// GetLatestItems godoc
func (h *LibraryHandler) GetLatestItems(c *gin.Context) {
	userId := c.Param("id")
	if userId == "" {
		userId = c.GetString("UserID")
	}
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

	respItems := make([]BaseItemDto, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, h.mapItem(item, userId))
	}

	// Devolvemos tanto el array plano como el objeto envuelto para máxima compatibilidad
	c.JSON(http.StatusOK, respItems)
}

func (h *LibraryHandler) GetSpecialFeatures(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func (h *LibraryHandler) GetAncestors(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func (h *LibraryHandler) GetSimilarItems(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Items": []interface{}{}, "TotalRecordCount": 0})
}

func (h *LibraryHandler) GetMediaSegments(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Items": []interface{}{}, "TotalRecordCount": 0})
}
