package handlers

import (
	"github.com/gellyte/gellyte/internal/services"
)

type Handler struct {
	AuthService     services.AuthService
	LibraryService  services.LibraryService
	PlaybackService services.PlaybackService
}

func NewHandler(auth services.AuthService, lib services.LibraryService, playback services.PlaybackService) *Handler {
	return &Handler{
		AuthService:     auth,
		LibraryService:  lib,
		PlaybackService: playback,
	}
}
