package models

import (
	"time"
)

// UserItemData representa la relación entre un usuario y un contenido (progreso, favoritos, etc).
type UserItemData struct {
	UserID                string    `gorm:"primaryKey" json:"UserId"`
	MediaItemID           string    `gorm:"primaryKey" json:"MediaItemId"`
	PlaybackPositionTicks int64     `json:"PlaybackPositionTicks"`
	PlayCount             int       `json:"PlayCount"`
	IsFavorite            bool      `json:"IsFavorite"`
	Played                bool      `json:"Played"`
	LastPlayedDate        time.Time `json:"LastPlayedDate"`
}
