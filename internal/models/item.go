package models

import (
	"crypto/md5"
	"encoding/hex"
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
	Overview       string    `json:"Overview,omitempty"`
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
		// Jellyfin exige UUIDs estándar de 32 caracteres (formato hexadecimal sin guiones).
		// Un hash MD5 de la ruta garantiza 32 caracteres y hace que el ID sea determinista,
		// evitando que el usuario pierda su progreso de visualización si se re-escanea.
		hash := md5.Sum([]byte(m.Path))
		m.ID = hex.EncodeToString(hash[:])
	}
	return
}
