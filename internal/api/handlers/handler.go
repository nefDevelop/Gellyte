package handlers

import (
	"strings"

	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/services"
)

type baseHandler struct {
	LibraryService services.LibraryService
}

// mapItem es una utilidad común para convertir MediaItem a DTO obteniendo datos de usuario.
func (b *baseHandler) mapItem(item models.MediaItem, userId string) BaseItemDto {
	userData, _ := b.LibraryService.GetUserData(userId, item.ID)
	sId := strings.ReplaceAll(config.AppConfig.Jellyfin.ServerUUID, "-", "")
	return MapMediaItemToDto(item, userData, sId)
}

type AuthHandler struct {
	baseHandler
	AuthService services.AuthService
}

type LibraryHandler struct {
	baseHandler
}

type PlaybackHandler struct {
	baseHandler
	PlaybackService services.PlaybackService
}

type SystemHandler struct {
	baseHandler
}

type SessionHandler struct {
	baseHandler
	AuthService services.AuthService
}

type WebSocketHandler struct {
	baseHandler
}

// Handler es ahora un contenedor para facilitar el registro inicial
type Handler struct {
	Auth      *AuthHandler
	Library   *LibraryHandler
	Playback  *PlaybackHandler
	System    *SystemHandler
	Session   *SessionHandler
	WebSocket *WebSocketHandler
}

func NewHandler(auth services.AuthService, lib services.LibraryService, playback services.PlaybackService) *Handler {
	base := baseHandler{LibraryService: lib}
	return &Handler{
		Auth:      &AuthHandler{baseHandler: base, AuthService: auth},
		Library:   &LibraryHandler{baseHandler: base},
		Playback:  &PlaybackHandler{baseHandler: base, PlaybackService: playback},
		System:    &SystemHandler{baseHandler: base},
		Session:   &SessionHandler{baseHandler: base, AuthService: auth},
		WebSocket: &WebSocketHandler{baseHandler: base},
	}
}
