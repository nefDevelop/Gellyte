package handlers

import (
	"net/http"
	"time"

	"github.com/gellyte/gellyte/internal/api/middleware"
	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gin-gonic/gin"
)

// GetSessions godoc
// @Summary Obtener lista de sesiones
// @Description Devuelve una lista de las sesiones activas en el servidor.
// @Tags Session
// @Produce json
// @Success 200 {array} SessionInfoDto
// @Router /Sessions [get]
func GetSessions(c *gin.Context) {
	// Intentamos obtener la info de auth del middleware
	val, exists := c.Get("auth")
	if !exists {
		c.JSON(http.StatusOK, []SessionInfoDto{})
		return
	}

	auth, ok := val.(middleware.EmbyAuth)
	if !ok {
		c.JSON(http.StatusOK, []SessionInfoDto{})
		return
	}

	// Buscamos al usuario admin por defecto o el que corresponda al token si hubiera lógica de tokens real
	var user models.User
	database.DB.Where("username = ?", "admin").First(&user)

	now := time.Now().UTC().Format(time.RFC3339)

	// Creamos una sesión "fake" pero realista basada en la petición actual
	session := SessionInfoDto{
		PlayState: PlayerStateInfo{
			CanSeek:     true,
			VolumeLevel: 100,
			PlayMethod:  "DirectPlay",
			RepeatMode:  "RepeatNone",
		},
		RemoteEndPoint:       c.ClientIP(),
		PlayableMediaTypes:   []string{"Audio", "Video"},
		Id:                   auth.Token,
		UserId:               user.ID,
		UserName:             user.Username,
		Client:               auth.Client,
		LastActivityDate:     now,
		LastPlaybackCheckIn:  now,
		DeviceName:           auth.Device,
		DeviceType:           "Mobile",
		DeviceId:             auth.DeviceId,
		ApplicationVersion:   auth.Version,
		IsActive:             true,
		SupportsMediaControl: true,
		SupportsRemoteControl: true,
		ServerId:             ServerUUID,
		SupportedCommands:    []string{"Play", "Pause", "Stop", "Seek", "NextTrack", "PreviousTrack"},
		NowPlayingQueue:      []interface{}{},
		Capabilities: ClientCapabilities{
			PlayableMediaTypes:   []string{"Audio", "Video"},
			SupportedCommands:    []string{"Play", "Pause", "Stop", "Seek", "NextTrack", "PreviousTrack"},
			SupportsMediaControl: true,
			SupportsPersistentIdentifier: true,
			SupportsSync:         false,
		},
		AdditionalUsers: []interface{}{},
	}

	c.JSON(http.StatusOK, []SessionInfoDto{session})
}
