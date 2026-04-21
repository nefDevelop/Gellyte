package handlers

import (
	"net/http"
	"time"

	"github.com/gellyte/gellyte/internal/api/middleware"
	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gin-gonic/gin"
)

// GetSessions godoc
func (h *Handler) GetSessions(c *gin.Context) {
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

	users, _ := h.AuthService.GetAllUsers()
	var user models.User
	if len(users) > 0 {
		user = users[0] // admin or first user
	}

	now := time.Now().UTC().Format(time.RFC3339)

	session := SessionInfoDto{
		PlayState: PlayerStateInfo{
			CanSeek:     true,
			VolumeLevel: 100,
			PlayMethod:  "DirectPlay",
			RepeatMode:  "RepeatNone",
		},
		RemoteEndPoint:        c.ClientIP(),
		PlayableMediaTypes:    []string{"Video"},
		Id:                    auth.Token,
		UserId:                user.ID,
		UserName:              user.Username,
		Client:                auth.Client,
		LastActivityDate:      now,
		LastPlaybackCheckIn:   now,
		DeviceName:            auth.Device,
		DeviceType:            "Mobile",
		DeviceId:              auth.DeviceId,
		ApplicationVersion:    auth.Version,
		IsActive:              true,
		SupportsMediaControl:  true,
		SupportsRemoteControl: true,
		ServerId:              config.AppConfig.Jellyfin.ServerUUID,
		SupportedCommands:     []string{"Play", "Pause", "Stop", "Seek", "NextTrack", "PreviousTrack"},
		NowPlayingQueue:       []interface{}{},
		Capabilities: ClientCapabilities{
			PlayableMediaTypes:           []string{"Video"},
			SupportedCommands:            []string{"Play", "Pause", "Stop", "Seek", "NextTrack", "PreviousTrack"},
			SupportsMediaControl:         true,
			SupportsPersistentIdentifier: true,
			SupportsSync:                 false,
		},
		AdditionalUsers: []interface{}{},
	}

	c.JSON(http.StatusOK, []SessionInfoDto{session})
}
