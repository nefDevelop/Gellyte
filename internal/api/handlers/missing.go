package handlers

import (
	"net/http"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/services"
	"github.com/gin-gonic/gin"
)

// GetItemsCounts godoc
func (h *LibraryHandler) GetItemsCounts(c *gin.Context) {
	var movieCount, seriesCount, episodeCount int64
	
	_, movieCount, _ = h.LibraryService.GetItems(services.GetItemsParams{ItemTypes: []string{"Movie"}, Limit: 1})
	_, seriesCount, _ = h.LibraryService.GetItems(services.GetItemsParams{ItemTypes: []string{"Series"}, Limit: 1})
	_, episodeCount, _ = h.LibraryService.GetItems(services.GetItemsParams{ItemTypes: []string{"Episode"}, Limit: 1})

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
func (h *LibraryHandler) GetItemsFilters(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Genres":          []string{},
		"Tags":            []string{},
		"OfficialRatings": []string{},
		"Years":           []int{},
	})
}

// GetItemsRoot godoc
func (h *LibraryHandler) GetItemsRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Name":           "Server",
		"Id":             config.AppConfig.Jellyfin.ServerUUID,
		"IsFolder":       true,
		"Type":           "AggregateFolder",
		"CollectionType": "folders",
	})
}

// GetMediaFolders godoc
func (h *LibraryHandler) GetMediaFolders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Items": []gin.H{
			{
				"Name":           "Películas",
				"ServerId":       config.AppConfig.Jellyfin.ServerUUID,
				"Id":             config.AppConfig.Jellyfin.MoviesLibraryID,
				"Type":           "CollectionFolder",
				"CollectionType": "movies",
			},
			{
				"Name":           "Series",
				"ServerId":       config.AppConfig.Jellyfin.ServerUUID,
				"Id":             config.AppConfig.Jellyfin.SeriesLibraryID,
				"Type":           "CollectionFolder",
				"CollectionType": "tvshows",
			},
		},
		"TotalRecordCount": 2,
	})
}

// GetPhysicalPaths godoc
func (h *LibraryHandler) GetPhysicalPaths(c *gin.Context) {
	c.JSON(http.StatusOK, []string{})
}

// GetGroupingOptions godoc
func (h *LibraryHandler) GetGroupingOptions(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{})
}

// GetShowEpisodes godoc
func (h *LibraryHandler) GetShowEpisodes(c *gin.Context) {
	seriesId := c.Param("id")
	if seriesId == "" {
		seriesId = c.Query("seriesId")
	}
	seasonId := c.Query("seasonId")

	var items []models.MediaItem
	var total int64

	if seasonId != "" {
		items, total, _ = h.LibraryService.GetItems(services.GetItemsParams{ParentID: seasonId, ItemTypes: []string{"Episode"}})
	} else {
		seasons, _, _ := h.LibraryService.GetItems(services.GetItemsParams{ParentID: seriesId, ItemTypes: []string{"Season"}})
		if len(seasons) > 0 {
			var seasonIDs []string
			for _, s := range seasons {
				seasonIDs = append(seasonIDs, s.ID)
			}
			items, total, _ = h.LibraryService.GetItems(services.GetItemsParams{ItemTypes: []string{"Episode"}})
			filteredItems := []models.MediaItem{}
			for _, it := range items {
				for _, sid := range seasonIDs {
					if it.ParentID == sid {
						filteredItems = append(filteredItems, it)
						break
					}
				}
			}
			items = filteredItems
			total = int64(len(items))
		} else {
			items, total, _ = h.LibraryService.GetItems(services.GetItemsParams{ParentID: seriesId, ItemTypes: []string{"Episode"}})
		}
	}

	userId := c.GetString("UserID")
	if userId == "" {
		userId = config.AppConfig.Jellyfin.AdminUUID
	}

	respItems := []BaseItemDto{}
	for _, item := range items {
		respItems = append(respItems, h.mapItem(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": total,
	})
}

// GetShowSeasons godoc
func (h *LibraryHandler) GetShowSeasons(c *gin.Context) {
	seriesId := c.Param("id")

	items, total, _ := h.LibraryService.GetItems(services.GetItemsParams{ParentID: seriesId, ItemTypes: []string{"Season"}})

	userId := c.GetString("UserID")
	if userId == "" {
		userId = config.AppConfig.Jellyfin.AdminUUID
	}

	respItems := []BaseItemDto{}
	for _, item := range items {
		respItems = append(respItems, h.mapItem(item, userId))
	}

	c.JSON(http.StatusOK, gin.H{
		"Items":            respItems,
		"TotalRecordCount": total,
	})
}
