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
			LastPlayedDate:        userData.LastPlayedDate.Format(time.RFC3339),
		}
	}

	dto := BaseItemDto{
		Name:         item.Name,
		Id:           item.ID,
		ServerId:     serverID,
		Type:         item.Type,
		RunTimeTicks: item.RunTimeTicks,
		IsFolder:     item.Type == "Series" || item.Type == "Season" || item.Type == "Folder" || item.Type == "CollectionFolder",
		ImageTags: map[string]string{
			"Primary": "tag",
		},
		UserData: userDataDto,
		Width:    item.Width,
		Height:   item.Height,
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
