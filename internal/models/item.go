package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// MediaItem representa cualquier contenido (Película, Serie, Carpeta) en el sistema Gellyte.
type MediaItem struct {
	ID             string    `gorm:"primaryKey" json:"Id"`
	Name           string    `gorm:"index" json:"Name"`
	Path           string    `json:"Path"`
	Type           string    `gorm:"index" json:"Type"` // Movie, Series, Episode, Folder
	ParentID       string    `gorm:"index" json:"ParentId,omitempty"`
	RunTimeTicks   int64     `json:"RunTimeTicks,omitempty"`
	ProductionYear int       `json:"ProductionYear,omitempty"`
	Container      string    `json:"Container,omitempty"`
	Width          int       `json:"Width,omitempty"`
	Height         int       `json:"Height,omitempty"`
	Bitrate        int64     `json:"Bitrate,omitempty"`
	VideoCodec     string    `json:"VideoCodec,omitempty"`
	AudioCodec     string    `json:"AudioCodec,omitempty"`
	Size           int64     `json:"Size,omitempty"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

// BeforeCreate genera un ID único para el item antes de guardarlo en la DB.
func (m *MediaItem) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		// Generamos un ID basado en el tiempo actual (nanosegundos)
		// Suficiente para este MVP.
		m.ID = fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return
}
