package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gellyte/gellyte/internal/api/middleware"
	"github.com/gellyte/gellyte/internal/config"
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gin-gonic/gin"
)

type AuthRequest struct {
	Username string `json:"Username"`
	Pw       string `json:"Pw"`
}

func (h *Handler) GetPublicUsers(c *gin.Context) {
	users, err := h.AuthService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	resp := []UserDto{}
	for _, u := range users {
		userObj := UserDto{
			Name:                      u.Username,
			ServerId:                  config.AppConfig.Jellyfin.ServerUUID,
			ServerName:                config.AppConfig.Server.Name,
			Id:                        u.ID,
			HasPassword:               true,
			HasConfiguredPassword:     true,
			HasConfiguredEasyPassword: true,
			EnableAutoLogin:           true,
			LastLoginDate:             now,
			LastActivityDate:          now,
			Configuration:             getDefaultConfigurationDto(),
			Policy:                    getDefaultPolicyDto(u.IsAdmin),
			PrimaryImageAspectRatio:   1.0,
			PrimaryImageTag:           "tag",
		}
		resp = append(resp, userObj)
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) AuthenticateByName(c *gin.Context) {
	var req AuthRequest
	clientAuth, _ := c.Get("auth")
	authInfo, ok := clientAuth.(middleware.EmbyAuth)
	if !ok {
		authInfo = middleware.EmbyAuth{Client: "Generic", Device: "Unknown", DeviceId: "unknown", Version: "1.0.0"}
	}

	username := c.Query("username")
	var pw string
	if err := c.ShouldBindJSON(&req); err == nil {
		if username == "" {
			username = req.Username
		}
		pw = req.Pw
	}

	user, token, err := h.AuthService.Authenticate(username, pw, authInfo.DeviceId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.Header("X-Emby-Token", token)
	c.Header("X-MediaBrowser-Token", token)
	c.Header("Access-Control-Expose-Headers", "X-Emby-Token, X-Emby-Authorization, X-MediaBrowser-Token")

	now := time.Now().UTC().Format(time.RFC3339)

	deviceType := "Mobile"
	clientLower := strings.ToLower(authInfo.Client)
	deviceLower := strings.ToLower(authInfo.Device)
	if strings.Contains(clientLower, "tv") || strings.Contains(deviceLower, "tv") || strings.Contains(clientLower, "box") {
		deviceType = "Tv"
	}

	// Crear un SessionId consistente (MD5 del deviceId)
	sHash := md5.Sum([]byte(authInfo.DeviceId))
	sessionId := hex.EncodeToString(sHash[:])

	authResult := AuthenticationResult{
		User: UserDto{
			Name:                      user.Username,
			ServerId:                  config.AppConfig.Jellyfin.ServerUUID,
			ServerName:                config.AppConfig.Server.Name,
			Id:                        user.ID,
			HasPassword:               true,
			HasConfiguredPassword:     true,
			HasConfiguredEasyPassword: true,
			EnableAutoLogin:           true,
			LastLoginDate:             now,
			LastActivityDate:          now,
			Configuration:             getDefaultConfigurationDto(),
			Policy:                    getDefaultPolicyDto(user.IsAdmin),
			PrimaryImageAspectRatio:   1.0,
			PrimaryImageTag:           "tag",
		},
		SessionInfo: SessionInfoDto{
			PlayState: PlayerStateInfo{
				CanSeek:     true,
				VolumeLevel: 100,
				PlayMethod:  "DirectPlay",
				RepeatMode:  "RepeatNone",
			},
			RemoteEndPoint:        c.ClientIP(),
			PlayableMediaTypes:    []string{"Video"},
			Id:                    sessionId,
			UserId:                user.ID,
			UserName:              user.Username,
			Client:                authInfo.Client,
			LastActivityDate:      now,
			LastPlaybackCheckIn:   now,
			LastPausedDate:        nil,
			DeviceName:            authInfo.Device,
			DeviceType:            deviceType,
			DeviceId:              authInfo.DeviceId,
			ApplicationVersion:    authInfo.Version,
			IsActive:              true,
			SupportsMediaControl:  true,
			SupportsRemoteControl: true,
			NowPlayingItem:        nil,
			NowViewingItem:        nil,
			ServerId:              config.AppConfig.Jellyfin.ServerUUID,
			SupportedCommands:     []string{"Play", "Pause", "Stop", "Seek", "NextTrack", "PreviousTrack"},
			NowPlayingQueue:       []interface{}{},
			Capabilities: ClientCapabilities{
				PlayableMediaTypes:           []string{"Video"},
				SupportedCommands:            []string{"Play", "Pause", "Stop", "Seek", "NextTrack", "PreviousTrack"},
				SupportsMediaControl:         true,
				SupportsPersistentIdentifier: true,
				SupportsSync:                 false,
				DeviceProfile: gin.H{
					"Name":                authInfo.Device,
					"SupportedMediaTypes": []string{"Video"},
					"DirectPlayProfiles":  []interface{}{},
					"TranscodingProfiles": []interface{}{},
					"ContainerProfiles":   []interface{}{},
					"CodecProfiles":       []interface{}{},
					"SubtitleProfiles":    []interface{}{},
				},
				AppStoreUrl: "",
				IconUrl:     "",
			},
			AdditionalUsers: []interface{}{},
		},
		AccessToken: token,
		ServerId:    config.AppConfig.Jellyfin.ServerUUID,
	}

	c.JSON(http.StatusOK, authResult)
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
	users, _ := h.AuthService.GetAllUsers()
	var adminUser *models.User
	for _, u := range users {
		if u.Username == "admin" {
			adminUser = &u
			break
		}
	}
	
	if adminUser == nil && len(users) > 0 {
		adminUser = &users[0]
	}

	if adminUser == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No users found"})
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	c.JSON(http.StatusOK, UserDto{
		Name:                      adminUser.Username,
		ServerId:                  config.AppConfig.Jellyfin.ServerUUID,
		ServerName:                config.AppConfig.Server.Name,
		Id:                        adminUser.ID,
		HasPassword:               true,
		HasConfiguredPassword:     true,
		HasConfiguredEasyPassword: true,
		EnableAutoLogin:           true,
		LastLoginDate:             now,
		LastActivityDate:          now,
		Configuration:             getDefaultConfigurationDto(),
		Policy:                    getDefaultPolicyDto(adminUser.IsAdmin),
		PrimaryImageAspectRatio:   1.0,
		PrimaryImageTag:           "tag",
	})
}

func (h *Handler) GetUserById(c *gin.Context) {
	id := c.Param("id")
	user, err := h.AuthService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Name":                      user.Username,
		"Id":                        user.ID,
		"ServerId":                  config.AppConfig.Jellyfin.ServerUUID,
		"PrimaryImageTag":           "tag",
		"HasPassword":               true,
		"HasConfiguredPassword":     true,
		"HasConfiguredEasyPassword": true,
		"EnableAutoLogin":           true,
		"PrimaryImageAspectRatio":   1.0,
		"Policy":                    getDefaultPolicyDto(user.IsAdmin),
		"Configuration":             getDefaultConfigurationDto(),
	})
}

func (h *Handler) GetUserViews(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Items": []gin.H{
			{
				"Name":           "Películas",
				"ServerId":       config.AppConfig.Jellyfin.ServerUUID,
				"Id":             config.AppConfig.Jellyfin.MoviesLibraryID,
				"Type":           "CollectionFolder",
				"CollectionType": "movies",
				"ImageTags":      gin.H{},
				"UserData": gin.H{
					"PlaybackPositionTicks": 0,
					"PlayCount":             0,
					"IsFavorite":            false,
					"Played":                false,
				},
				"IsFolder":                true,
				"CanDelete":               false,
				"IsFavorite":              false,
				"PlayAccess":              "Full",
				"PrimaryImageAspectRatio": 1.0,
			},
			{
				"Name":           "Series",
				"ServerId":       config.AppConfig.Jellyfin.ServerUUID,
				"Id":             config.AppConfig.Jellyfin.SeriesLibraryID,
				"Type":           "CollectionFolder",
				"CollectionType": "tvshows",
				"ImageTags":      gin.H{},
				"UserData": gin.H{
					"PlaybackPositionTicks": 0,
					"PlayCount":             0,
					"IsFavorite":            false,
					"Played":                false,
				},
				"IsFolder":                true,
				"CanDelete":               false,
				"IsFavorite":              false,
				"PlayAccess":              "Full",
				"PrimaryImageAspectRatio": 1.0,
			},
		},
		"TotalRecordCount": 2,
	})
}

func (h *Handler) GetDisplayPreferences(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Id":               "default",
		"ViewType":         "Default",
		"SortBy":           "SortName",
		"SortOrder":        "Ascending",
		"RememberIndexing": false,
		"RememberSorting":  false,
		"CustomPrefs":      gin.H{},
		"Client":           "Generic",
	})
}

// --- HELPER FUNCTIONS ---

func getDefaultPolicyDto(isAdmin bool) UserPolicy {
	return UserPolicy{
		IsAdministrator:                  isAdmin,
		IsHidden:                         false,
		EnableCollectionManagement:       true,
		EnableSubtitleManagement:         true,
		EnableLyricManagement:            false, // Ocultar UI de música y letras
		IsDisabled:                       false,
		MaxParentalRating:                0,
		MaxParentalSubRating:             0,
		BlockedTags:                      []string{},
		AllowedTags:                      []string{},
		EnableUserPreferenceAccess:       true,
		AccessSchedules:                  []interface{}{},
		BlockUnratedItems:                []interface{}{},
		EnableRemoteControlOfOtherUsers:  true,
		EnableSharedDeviceControl:        true,
		EnableRemoteAccess:               true,
		EnableLiveTvManagement:           false, // Ocultar el menú de Live TV
		EnableLiveTvAccess:               false, // Ocultar el menú de Live TV
		EnableMediaPlayback:              true,
		EnableAudioPlaybackTranscoding:   false, // Indicar que no manejamos audio puro
		EnableVideoPlaybackTranscoding:   true,
		EnablePlaybackRemuxing:           true,
		ForceRemoteSourceTranscoding:     false,
		EnableContentDeletion:            true,
		EnableContentDeletionFromFolders: []string{},
		EnableContentDownloading:         false, // Ocultar botones de descarga
		EnableSyncTranscoding:            false, // Ocultar UI de sincronización/SyncPlay
		EnableMediaConversion:            true,
		EnabledDevices:                   []string{},
		EnableAllDevices:                 true,
		EnabledChannels:                  []string{},
		EnableAllChannels:                false, // Ocultar el menú de Canales
		EnabledFolders:                   []string{},
		EnableAllFolders:                 true,
		InvalidLoginAttemptCount:         0,
		LoginAttemptsBeforeLockout:       0,
		MaxActiveSessions:                0,
		EnablePublicSharing:              true,
		BlockedMediaFolders:              []string{},
		BlockedChannels:                  []string{},
		RemoteClientBitrateLimit:         0,
		AuthenticationProviderId:         "Default",
		PasswordResetProviderId:          "Default",
		SyncPlayAccess:                   "CreateAndJoinGroups",
	}
}

func getDefaultConfigurationDto() UserConfiguration {
	return UserConfiguration{
		AudioLanguagePreference:    "es",
		PlayDefaultAudioTrack:      true,
		SubtitleLanguagePreference: "es",
		DisplayMissingEpisodes:     false,
		GroupedFolders:             []string{},
		SubtitleMode:               "Default",
		DisplayCollectionsView:     false,
		EnableLocalPassword:        true,
		OrderedViews:               []string{},
		LatestItemsExcludes:        []string{},
		MyMediaExcludes:            []string{},
		HidePlayedInLatest:         true,
		RememberAudioSelections:    true,
		RememberSubtitleSelections: true,
		EnableNextEpisodeAutoPlay:  true,
		CastReceiverId:             "",
	}
}
