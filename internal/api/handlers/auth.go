package handlers

import (
	"encoding/hex"
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

	now := time.Now().UTC().Format(time.RFC3339)
	resp := []UserDto{}
	for _, u := range users {
		userObj := UserDto{
			Name:                      u.Username,
			ServerId:                  strings.ReplaceAll(config.AppConfig.Jellyfin.ServerUUID, "-", ""),
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

	// Generar un token con formato JWT para mayor compatibilidad con clientes de TV
	jwtHeader := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" // Header standard
	jwtPayload := hex.EncodeToString([]byte(token))     // Payload dummy basado en el token MD5
	fullToken := jwtHeader + "." + jwtPayload + ".dummy_signature"

	// Formato de fecha con 7 decimales (precisión de Jellyfin)
	now := time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z")

	sId := strings.ReplaceAll(config.AppConfig.Jellyfin.ServerUUID, "-", "")
	uId := strings.ReplaceAll(user.ID, "-", "")

	c.Header("X-Emby-Token", fullToken)
	c.Header("X-MediaBrowser-Token", fullToken)
	c.Header("Access-Control-Expose-Headers", "X-Emby-Token, X-Emby-Authorization, X-MediaBrowser-Token")
	c.Header("Content-Type", "application/json")

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
		SessionInfo: nil,
		AccessToken: fullToken,
		ServerId:    sId,
	}

	c.Header("Content-Type", "application/json")

	jsonBytes, err := json.Marshal(authResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error encoding JSON"})
		return
	}
	c.Data(http.StatusOK, "application/json", jsonBytes)
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
		"Id":                        strings.ReplaceAll(user.ID, "-", ""),
		"ServerId":                  strings.ReplaceAll(config.AppConfig.Jellyfin.ServerUUID, "-", ""),
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
