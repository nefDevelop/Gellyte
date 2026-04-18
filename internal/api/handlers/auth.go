package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gellyte/gellyte/internal/api/middleware"
	"github.com/gellyte/gellyte/internal/database"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gin-gonic/gin"
)

const ServerUUID = "83e4c49d-9273-4556-9a5d-4952011702f3"
const AdminUUID = "53896590-3b41-46a4-9591-96b054a8e3f6"

type AuthRequest struct {
	Username string `json:"Username"`
	Pw       string `json:"Pw"`
}

func GetPublicUsers(c *gin.Context) {
	var users []models.User
	database.DB.Find(&users)

	log.Printf("[Auth] /Users/Public -> Encontrados: %d", len(users))

	// Formato de fecha simplificado sin nanosegundos para máxima compatibilidad
	now := time.Now().UTC().Format(time.RFC3339)
	resp := []gin.H{}
	for _, u := range users {
		resp = append(resp, gin.H{
			"Name":                      u.Username,
			"ServerId":                  ServerUUID,
			"ServerName":                "Gellyte",
			"Id":                        u.ID,
			"PrimaryImageTag":           "",
			"HasPassword":               true,
			"HasConfiguredPassword":     true,
			"HasConfiguredEasyPassword": true,
			"EnableAutoLogin":           true,
			"LastLoginDate":             now,
			"LastActivityDate":          now,
			"Configuration":             getDefaultConfiguration(),
			"Policy":                    getDefaultPolicy(u.IsAdmin),
			"PrimaryImageAspectRatio":   1.0,
		})
	}
	c.JSON(http.StatusOK, resp)
}

func AuthenticateByName(c *gin.Context) {
	var username string
	var pw string
	var req AuthRequest

	clientAuth, _ := c.Get("auth")
	authInfo, ok := clientAuth.(middleware.EmbyAuth)
	if !ok {
		authInfo = middleware.EmbyAuth{
			Client:   "Findroid",
			Device:   "Mobile",
			DeviceId: "unknown",
			Version:  "1.0.0",
		}
	}

	username = c.Query("username")
	if err := c.ShouldBindJSON(&req); err == nil {
		if username == "" {
			username = req.Username
		}
		pw = req.Pw
	}

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado"})
		return
	}

	if user.Password != "" && pw != user.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password incorrecta"})
		return
	}

	token := "0309117604954714b10508a8e100f90c"
	c.Header("X-Emby-Token", token)
	c.Header("X-MediaBrowser-Token", token)

	now := time.Now().UTC().Format(time.RFC3339)

	c.JSON(http.StatusOK, gin.H{
		"User": gin.H{
			"Name":                      user.Username,
			"ServerId":                  ServerUUID,
			"ServerName":                "Gellyte",
			"Id":                        user.ID,
			"PrimaryImageTag":           "",
			"HasPassword":               true,
			"HasConfiguredPassword":     true,
			"HasConfiguredEasyPassword": true,
			"EnableAutoLogin":           true,
			"LastLoginDate":             now,
			"LastActivityDate":          now,
			"Configuration":             getDefaultConfiguration(),
			"Policy":                    getDefaultPolicy(user.IsAdmin),
			"PrimaryImageAspectRatio":   1.0,
		},
		"SessionInfo": gin.H{
			"PlayState": gin.H{
				"PositionTicks":       0,
				"CanSeek":             true,
				"IsPaused":            false,
				"IsMuted":             false,
				"VolumeLevel":         100,
				"AudioStreamIndex":    0,
				"SubtitleStreamIndex": -1,
				"MediaSourceId":       "",
				"PlayMethod":          "DirectPlay",
				"RepeatMode":          "RepeatNone",
				"PlaybackOrder":       "Default",
				"LiveStreamId":        "",
			},
			"AdditionalUsers": []gin.H{},
			"Capabilities": gin.H{
				"PlayableMediaTypes": []string{"Audio", "Video"},
				"SupportedCommands": []string{
					"MoveUp", "MoveDown", "MoveLeft", "MoveRight", "PageUp", "PageDown",
					"PreviousLetter", "NextLetter", "ToggleOsd", "ToggleContextMenu",
					"Select", "Back", "TakeScreenshot", "SendKey", "SendString",
					"GoHome", "GoToSettings", "VolumeUp", "VolumeDown", "Mute",
					"Unmute", "ToggleMute", "SetVolume", "SetAudioStreamIndex",
					"SetSubtitleStreamIndex", "DisplayContent", "GoToSearch",
					"DisplayMessage", "SetRepeatMode", "ChannelUp", "ChannelDown",
					"Guide", "ToggleStats", "PlayMediaItem", "PlayTrailers",
				},
				"SupportsMediaControl":         true,
				"SupportsPersistentIdentifier": true,
				"SupportsContentUploading":     false,
				"MessageCallbackUrl":           "",
				"SupportsSync":                 false,
				"DeviceProfile":                nil,
				"AppStoreUrl":                  "",
				"IconUrl":                      "",
			},
			"RemoteEndPoint":           c.ClientIP(),
			"PlayableMediaTypes":       []string{"Audio", "Video"},
			"Id":                       token,
			"UserId":                   user.ID,
			"UserName":                 user.Username,
			"Client":                   authInfo.Client,
			"LastActivityDate":         now,
			"LastPlaybackCheckIn":      now,
			"LastPausedDate":           now,
			"DeviceName":               authInfo.Device,
			"DeviceType":               "Mobile",
			"DeviceId":                 authInfo.DeviceId,
			"ApplicationVersion":       authInfo.Version,
			"NowPlayingItem":           nil,
			"NowViewingItem":           nil,
			"TranscodingInfo":          nil,
			"PlaylistItemId":           "",
			"IsActive":                 true,
			"SupportsMediaControl":     true,
			"SupportsRemoteControl":    true,
			"NowPlayingQueue":          []gin.H{},
			"NowPlayingQueueFullItems": []gin.H{},
			"HasCustomDeviceName":      false,
			"ServerId":                 ServerUUID,
			"UserPrimaryImageTag":      "",
			"SupportedCommands": []string{
				"MoveUp", "MoveDown", "MoveLeft", "MoveRight", "PageUp", "PageDown",
				"Select", "Back", "GoHome", "GoToSettings",
			},
		},
		"AccessToken": token,
		"ServerId":    ServerUUID,
	})
}

func GetCurrentUser(c *gin.Context) {
	var user models.User
	if err := database.DB.Where("username = ?", "admin").First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario admin no encontrado en la DB"})
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)

	c.JSON(http.StatusOK, gin.H{
		"Name":                      user.Username,
		"ServerId":                  ServerUUID,
		"ServerName":                "Gellyte",
		"Id":                        user.ID,
		"PrimaryImageTag":           "",
		"HasPassword":               true,
		"HasConfiguredPassword":     true,
		"HasConfiguredEasyPassword": true,
		"EnableAutoLogin":           true,
		"LastLoginDate":             now,
		"LastActivityDate":          now,
		"Configuration":             getDefaultConfiguration(),
		"Policy":                    getDefaultPolicy(user.IsAdmin),
		"PrimaryImageAspectRatio":   1.0,
	})
}

func GetUserById(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Name":                      user.Username,
		"Id":                        user.ID,
		"ServerId":                  ServerUUID,
		"HasPassword":               true,
		"HasConfiguredPassword":     true,
		"HasConfiguredEasyPassword": true,
		"EnableAutoLogin":           true,
		"PrimaryImageAspectRatio":   1.0,
		"Policy":                    getDefaultPolicy(user.IsAdmin),
		"Configuration":             getDefaultConfiguration(),
	})
}

func GetUserViews(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Items": []gin.H{{
			"Name":           "Películas",
			"ServerId":       ServerUUID,
			"Id":             "12345678-1234-1234-1234-123456789012",
			"Type":           "CollectionFolder",
			"CollectionType": "movies",
			"IsFolder":       true,
		}},
		"TotalRecordCount": 1,
	})
}

func GetDisplayPreferences(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Id":               "default",
		"ViewType":         "Default",
		"SortBy":           "SortName",
		"SortOrder":        "Ascending",
		"RememberIndexing": false,
		"RememberSorting":  false,
	})
}

// --- HELPER FUNCTIONS ---

func getDefaultPolicy(isAdmin bool) gin.H {
	return gin.H{
		"IsAdministrator":                  isAdmin,
		"IsHidden":                         false,
		"EnableCollectionManagement":       true,
		"EnableSubtitleManagement":         true,
		"EnableLyricManagement":            true,
		"IsDisabled":                       false,
		"MaxParentalRating":                0,
		"MaxParentalSubRating":             0,
		"BlockedTags":                      []string{},
		"AllowedTags":                      []string{},
		"EnableUserPreferenceAccess":       true,
		"AccessSchedules":                  []gin.H{},
		"BlockUnratedItems":                []string{},
		"EnableRemoteControlOfOtherUsers":  true,
		"EnableSharedDeviceControl":        true,
		"EnableRemoteAccess":               true,
		"EnableLiveTvManagement":           true,
		"EnableLiveTvAccess":               true,
		"EnableMediaPlayback":              true,
		"EnableAudioPlaybackTranscoding":   true,
		"EnableVideoPlaybackTranscoding":   true,
		"EnablePlaybackRemuxing":           true,
		"ForceRemoteSourceTranscoding":     false,
		"EnableContentDeletion":            true,
		"EnableContentDeletionFromFolders": []string{},
		"EnableContentDownloading":         true,
		"EnableSyncTranscoding":            true,
		"EnableMediaConversion":            true,
		"EnabledDevices":                   []string{},
		"EnableAllDevices":                 true,
		"EnabledChannels":                  []string{},
		"EnableAllChannels":                true,
		"EnabledFolders":                   []string{},
		"EnableAllFolders":                 true,
		"InvalidLoginAttemptCount":         0,
		"LoginAttemptsBeforeLockout":       0,
		"MaxActiveSessions":                0,
		"EnablePublicSharing":              true,
		"BlockedMediaFolders":              []string{},
		"BlockedChannels":                  []string{},
		"RemoteClientBitrateLimit":         0,
		"AuthenticationProviderId":         "Default",
		"PasswordResetProviderId":          "Default",
		"SyncPlayAccess":                   "CreateAndJoinGroups",
	}
}

func getDefaultConfiguration() gin.H {
	return gin.H{
		"AudioLanguagePreference":    "es",
		"PlayDefaultAudioTrack":      true,
		"SubtitleLanguagePreference": "es",
		"DisplayMissingEpisodes":     false,
		"GroupedFolders":             []string{},
		"SubtitleMode":               "Default",
		"DisplayCollectionsView":     false,
		"EnableLocalPassword":        true,
		"OrderedViews":               []string{},
		"LatestItemsExcludes":        []string{},
		"MyMediaExcludes":            []string{},
		"HidePlayedInLatest":         true,
		"RememberAudioSelections":    true,
		"RememberSubtitleSelections": true,
		"EnableNextEpisodeAutoPlay":  true,
		"CastReceiverId":             "",
	}
}
