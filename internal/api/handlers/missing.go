package handlers

import (
	"net/http"

	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gin-gonic/gin"
)

// GetItemsCounts godoc
func GetItemsCounts(c *gin.Context) {
	var movieCount, seriesCount, episodeCount int64
	database.DB.Model(&models.MediaItem{}).Where("type = ?", "Movie").Count(&movieCount)
	database.DB.Model(&models.MediaItem{}).Where("type = ?", "Series").Count(&seriesCount)
	database.DB.Model(&models.MediaItem{}).Where("type = ?", "Episode").Count(&episodeCount)

	c.JSON(http.StatusOK, gin.H{
		"MovieCount":      movieCount,
		"SeriesCount":     seriesCount,
		"EpisodeCount":    episodeCount,
		"ArtistCount":     0,
		"ProgramCount":    0,
		"TrailerCount":    0,
		"SongCount":       0,
		"AlbumCount":      0,
		"MusicVideoCount": 0,
		"BoxSetCount":     0,
		"BookCount":       0,
		"ItemCount":       movieCount + seriesCount + episodeCount,
	})
}

// GetItemsFilters godoc
func GetItemsFilters(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Genres":          []string{},
		"Tags":            []string{},
		"OfficialRatings": []string{},
		"Years":           []int{},
	})
}

// GetItemsRoot godoc
func GetItemsRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Name":           "Server",
		"Id":             ServerUUID,
		"IsFolder":       true,
		"Type":           "AggregateFolder",
		"CollectionType": "folders",
	})
}

// GetMediaFolders godoc
func GetMediaFolders(c *gin.Context) {
	// Reutilizamos la lógica de VirtualFolders que ya escanea las colecciones
	c.JSON(http.StatusOK, gin.H{
		"Items": []gin.H{
			{
				"Name":           "Películas",
				"ServerId":       ServerUUID,
				"Id":             "12345678-1234-1234-1234-123456789012",
				"Type":           "CollectionFolder",
				"CollectionType": "movies",
			},
			{
				"Name":           "Series",
				"ServerId":       ServerUUID,
				"Id":             "22345678-1234-1234-1234-123456789012",
				"Type":           "CollectionFolder",
				"CollectionType": "tvshows",
			},
		},
		"TotalRecordCount": 2,
	})
}

// GetPhysicalPaths godoc
func GetPhysicalPaths(c *gin.Context) {
	c.JSON(http.StatusOK, []string{})
}

// GetGroupingOptions godoc
func GetGroupingOptions(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{})
}

// GetShowEpisodes godoc
func GetShowEpisodes(c *gin.Context) {
	seriesId := c.Param("id")
	if seriesId == "" {
		seriesId = c.Query("seriesId")
	}
	seasonId := c.Query("seasonId")

	var dbItems []models.MediaItem
	var total int64
	query := database.DB.Model(&models.MediaItem{}).Where("type = ?", "Episode")

	if seasonId != "" {
		query = query.Where("parent_id = ?", seasonId)
	} else if seriesId != "" {
		var seasonIds []string
		database.DB.Model(&models.MediaItem{}).Where("type = ? AND parent_id = ?", "Season", seriesId).Pluck("id", &seasonIds)
		if len(seasonIds) > 0 {
			query = query.Where("parent_id IN ? OR parent_id = ?", seasonIds, seriesId)
		} else {
			query = query.Where("parent_id = ?", seriesId)
		}
	}

	query.Count(&total)
	query.Find(&dbItems)

	userId := c.GetString("UserID")
	if userId == "" {
		userId = "53896590-3b41-46a4-9591-96b054a8e3f6"
	}

	respItems := []BaseItemDto{}
	for _, item := range dbItems {
		respItems = append(respItems, mapToDto(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": total,
	})
}

// GetShowSeasons godoc
func GetShowSeasons(c *gin.Context) {
	seriesId := c.Param("id")

	var dbItems []models.MediaItem
	var total int64
	query := database.DB.Model(&models.MediaItem{}).Where("type = ? AND parent_id = ?", "Season", seriesId)

	query.Count(&total)
	query.Find(&dbItems)

	userId := c.GetString("UserID")
	if userId == "" {
		userId = "53896590-3b41-46a4-9591-96b054a8e3f6"
	}

	respItems := []BaseItemDto{}
	for _, item := range dbItems {
		respItems = append(respItems, mapToDto(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": total,
	})
}
