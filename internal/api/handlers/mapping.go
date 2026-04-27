package handlers

import (
	"time"

	"github.com/gellyte/gellyte/internal/models"
)

// MapMediaItemToDto convierte un modelo de BD a un objeto de respuesta compatible con Jellyfin.
// Es una función pura que no depende de servicios, facilitando el desacoplamiento.
func MapMediaItemToDto(item models.MediaItem, userData *models.UserItemData, serverID string) BaseItemDto {
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
			LastPlayedDate:        userData.LastPlayedDate.Format("2006-01-02T15:04:05.0000000Z"),
		}
	}

	// Formato de fecha estándar de Jellyfin
	now := time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z")

	dto := BaseItemDto{
		Name:                    item.Name,
		SortName:                item.Name,
		Id:                      item.ID,
		Etag:                    item.ID, // Usamos el ID como Etag básico
		ServerId:                serverID,
		Type:                    item.Type,
		MediaType:               "Video",
		RunTimeTicks:            item.RunTimeTicks,
		IsFolder:                item.Type == "Series" || item.Type == "Season" || item.Type == "Folder" || item.Type == "CollectionFolder",
		DateCreated:             now,
		DateLastMediaAdded:      now,
		CanDelete:               false,
		PlayAccess:              "Full",
		PrimaryImageAspectRatio: 0.66,
		ImageTags: map[string]string{
			"Primary": "tag",
		},
		UserData:        userDataDto,
		Width:           item.Width,
		Height:          item.Height,
		ExternalUrls:    []interface{}{},
		MediaSources:    []interface{}{},
		ImageBlurHashes: map[string]interface{}{},
	}

	if item.Type == "Series" {
		dto.MediaType = "" // Series no tienen MediaType en Jellyfin
	}

	if item.ProductionYear > 0 {
		dto.ProductionYear = item.ProductionYear
	}
	if item.Overview != "" {
		dto.Overview = item.Overview
	}

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
