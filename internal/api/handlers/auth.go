package handlers

import (
	"encoding/json"
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

	now := time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z")
	sId := strings.ReplaceAll(config.AppConfig.Jellyfin.ServerUUID, "-", "")
	
	resp := []UserDto{}
	for _, u := range users {
		userObj := UserDto{
			Name:                      u.Username,
			ServerId:                  sId,
			ServerName:                config.AppConfig.Server.Name,
			Id:                        strings.ReplaceAll(u.ID, "-", ""),
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

	// Intentar obtener credenciales de múltiples fuentes (JSON, Form, Query)
	username := c.Query("Username")
	if username == "" {
		username = c.PostForm("Username")
	}
	pw := c.PostForm("Pw")

	if err := c.ShouldBindJSON(&req); err == nil {
		if username == "" {
			username = req.Username
		}
		if pw == "" {
			pw = req.Pw
		}
	}

	if username == "" {
		username = c.Query("username") // Fallback lowercase
	}

	user, token, err := h.AuthService.Authenticate(username, pw, authInfo.DeviceId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Formato de fecha con 7 decimales (precisión de Jellyfin)
	now := time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z")

	sId := strings.ReplaceAll(config.AppConfig.Jellyfin.ServerUUID, "-", "")
	uId := strings.ReplaceAll(user.ID, "-", "")

	// Detectar si es una TV
	deviceType := "Mobile"
	clientLower := strings.ToLower(authInfo.Client)
	deviceLower := strings.ToLower(authInfo.Device)
	if strings.Contains(clientLower, "tv") || strings.Contains(deviceLower, "tv") || strings.Contains(clientLower, "box") {
		deviceType = "Tv"
	}

	// Generar un JWT dummy más realista (Base64)
	jwtHeader := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	jwtPayload := "eyJ1c2VySWQiOiI" + uId + "Iiwic2VydmVySWQiOiI" + sId + "In0"
	fullToken := jwtHeader + "." + jwtPayload + ".dummy_signature"

	c.Header("X-Emby-Token", fullToken)
	c.Header("X-Emby-Authorization", "Token=\""+fullToken+"\"")
	c.Header("X-MediaBrowser-Token", fullToken)
	c.Header("Access-Control-Expose-Headers", "X-Emby-Token, X-Emby-Authorization, X-MediaBrowser-Token")

	authResult := AuthenticationResult{
		User: UserDto{
			Name:                      user.Username,
			ServerId:                  sId,
			ServerName:                config.AppConfig.Server.Name,
			Id:                        uId,
			HasPassword:               true,
			HasConfiguredPassword:     true,
			HasConfiguredEasyPassword: true,
			EnableAutoLogin:           false,
			LastLoginDate:             now,
			LastActivityDate:          now,
			Configuration: UserConfiguration{
				AudioLanguagePreference:    "es",
				SubtitleLanguagePreference: "es",
				DisplayMissingEpisodes:     true,
				SubtitleMode:               "Default",
				EnableNextEpisodeAutoPlay:  true,
				ResumePlayerState:          true,
				SyncPlayLikes:              true,
				EnableCinemaMode:           false,
				HidePlayedInSongs:          false,
				HidePlayedInVideos:         false,
				SkipSongsNotPlayed:         false,
			},
			Policy:                  getDefaultPolicyDto(user.IsAdmin),
			PrimaryImageAspectRatio: 1.0,
			PrimaryImageTag:         "tag",
		},
		SessionInfo: &SessionInfoDto{
			PlayState: PlayerStateInfo{
				CanSeek:             true,
				IsPaused:            false,
				IsMuted:             false,
				VolumeLevel:         100,
				AudioStreamIndex:    0,
				SubtitleStreamIndex: -1,
				PlayMethod:          "DirectPlay",
				RepeatMode:          "RepeatNone",
			},
			AdditionalUsers:    []interface{}{},
			Capabilities: ClientCapabilities{
				PlayableMediaTypes:           []string{"Audio", "Video"},
				SupportedCommands:            []string{"MoveUp", "MoveDown", "MoveLeft", "MoveRight", "Select", "Back", "Play", "Pause", "Stop", "TogglePause"},
				SupportsMediaControl:         true,
				SupportsPersistentIdentifier: true,
			},
			RemoteEndPoint:      c.ClientIP(),
			PlayableMediaTypes:  []string{"Audio", "Video"},
			Id:                  token, // Usamos el token interno como ID de sesión
			UserId:              uId,
			UserName:            user.Username,
			Client:              authInfo.Client,
			LastActivityDate:    now,
			LastPlaybackCheckIn: now,
			DeviceName:          authInfo.Device,
			DeviceType:          deviceType,
			DeviceId:            authInfo.DeviceId,
			ApplicationVersion:  authInfo.Version,
			IsActive:            true,
			SupportsMediaControl: true,
			SupportsRemoteControl: true,
			HasCustomDeviceName:  false,
			ServerId:            sId,
			SupportedCommands:   []string{"MoveUp", "MoveDown", "MoveLeft", "MoveRight", "Select", "Back", "Play", "Pause", "Stop", "TogglePause"},
			NowPlayingQueue:     []interface{}{},
			NowPlayingQueueFullItems: []interface{}{},
		},
		AccessToken: fullToken,
		ServerId:    sId,
	}

	jsonBytes, err := json.Marshal(authResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error encoding JSON"})
		return
	}
	c.Data(http.StatusOK, "application/json; profile=\"PascalCase\"", jsonBytes)
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

	now := time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z")
	sId := strings.ReplaceAll(config.AppConfig.Jellyfin.ServerUUID, "-", "")
	uId := strings.ReplaceAll(user.ID, "-", "")

	c.JSON(http.StatusOK, UserDto{
		Name:                      user.Username,
		Id:                        uId,
		ServerId:                  sId,
		ServerName:                config.AppConfig.Server.Name,
		PrimaryImageTag:           "tag",
		HasPassword:               true,
		HasConfiguredPassword:     true,
		HasConfiguredEasyPassword: true,
		EnableAutoLogin:           true,
		LastLoginDate:             now,
		LastActivityDate:          now,
		PrimaryImageAspectRatio:   1.0,
		Policy:                    getDefaultPolicyDto(user.IsAdmin),
		Configuration:             getDefaultConfigurationDto(),
	})
}

func (h *Handler) GetUserViews(c *gin.Context) {
	sId := strings.ReplaceAll(config.AppConfig.Jellyfin.ServerUUID, "-", "")
	
	views := []BaseItemDto{
		{
			Name:                    "Películas",
			Id:                      config.AppConfig.Jellyfin.MoviesLibraryID,
			ServerId:                sId,
			Type:                    "CollectionFolder",
			CollectionType:          "movies",
			IsFolder:                true,
			PlayAccess:              "Full",
			PrimaryImageAspectRatio: 0.66,
			ImageTags:               map[string]string{},
		},
		{
			Name:                    "Series",
			Id:                      config.AppConfig.Jellyfin.SeriesLibraryID,
			ServerId:                sId,
			Type:                    "CollectionFolder",
			CollectionType:          "tvshows",
			IsFolder:                true,
			PlayAccess:              "Full",
			PrimaryImageAspectRatio: 0.66,
			ImageTags:               map[string]string{},
		},
	}

	c.JSON(http.StatusOK, BaseItemDtoQueryResult{
		Items:            views,
		TotalRecordCount: len(views),
		StartIndex:       0,
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
		EnableLyricManagement:            false,
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
		EnableLiveTvManagement:           false,
		EnableLiveTvAccess:               false,
		EnableMediaPlayback:              true,
		EnableAudioPlaybackTranscoding:   false,
		EnableVideoPlaybackTranscoding:   true,
		EnablePlaybackRemuxing:           true,
		ForceRemoteSourceTranscoding:     false,
		EnableContentDeletion:            true,
		EnableContentDeletionFromFolders: []string{},
		EnableContentDownloading:         false,
		EnableSyncTranscoding:            false,
		EnableMediaConversion:            true,
		EnabledDevices:                   []string{},
		EnableAllDevices:                 true,
		EnabledChannels:                  []string{},
		EnableAllChannels:                false,
		EnabledFolders: []string{
			config.AppConfig.Jellyfin.MoviesLibraryID,
			config.AppConfig.Jellyfin.SeriesLibraryID,
		},
		EnableAllFolders:           true,
		InvalidLoginAttemptCount:   0,
		LoginAttemptsBeforeLockout: 0,
		MaxActiveSessions:          0,
		EnablePublicSharing:        true,
		BlockedMediaFolders:        []string{},
		BlockedChannels:            []string{},
		RemoteClientBitrateLimit:   0,
		AuthenticationProviderId:   "Default",
		PasswordResetProviderId:    "Default",
		SyncPlayAccess:             "CreateAndJoinGroups",
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
