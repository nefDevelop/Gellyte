package handlers

import (
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
	Username string `json:"Username" json:"username"`
	Pw       string `json:"Pw" json:"pw"`
}

func GetPublicUsers(c *gin.Context) {
	var users []models.User
	database.DB.Find(&users)

	now := time.Now().UTC().Format(time.RFC3339)
	resp := []UserDto{}
	for _, u := range users {
		userObj := UserDto{
			Name:                      u.Username,
			ServerId:                  ServerUUID,
			ServerName:                "Gellyte",
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
			PrimaryImageTag:           "",
		}
		resp = append(resp, userObj)
	}
	
	c.JSON(http.StatusOK, resp)
}

func AuthenticateByName(c *gin.Context) {
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

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado"})
		return
	}

	if user.Password != "" && pw != user.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password incorrecta"})
		return
	}

	token := "4841785566774a8481419457813a849b" // Sin guiones para máxima compatibilidad
	c.Header("X-Emby-Token", token)
	c.Header("X-MediaBrowser-Token", token)
	c.Header("Access-Control-Expose-Headers", "X-Emby-Token, X-Emby-Authorization, X-MediaBrowser-Token")

	now := time.Now().UTC().Format(time.RFC3339)

	authResult := AuthenticationResult{
		User: UserDto{
			Name:                      user.Username,
			ServerId:                  ServerUUID,
			ServerName:                "Gellyte",
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
			PrimaryImageTag:           "",
		},
		SessionInfo: SessionInfoDto{
			PlayState: PlayerStateInfo{
				CanSeek:    true,
				VolumeLevel: 100,
				PlayMethod:  "DirectPlay",
				RepeatMode:  "RepeatNone",
			},
			RemoteEndPoint:       c.ClientIP(),
			PlayableMediaTypes:   []string{"Audio", "Video"},
			Id:                   token,
			UserId:               user.ID,
			UserName:             user.Username,
			Client:               authInfo.Client,
			LastActivityDate:     now,
			LastPlaybackCheckIn:  now,
			LastPausedDate:       nil,
			DeviceName:           authInfo.Device,
			DeviceType:           "Mobile",
			DeviceId:             authInfo.DeviceId,
			ApplicationVersion:   authInfo.Version,
			IsActive:             true,
			SupportsMediaControl: true,
			SupportsRemoteControl: true,
			NowPlayingItem:       nil,
			NowViewingItem:       nil,
			ServerId:             ServerUUID,
			SupportedCommands:    []string{"Play", "Pause", "Stop", "Seek", "NextTrack", "PreviousTrack"},
			NowPlayingQueue:      []interface{}{},
			Capabilities: ClientCapabilities{
				PlayableMediaTypes:   []string{"Audio", "Video"},
				SupportedCommands:    []string{"Play", "Pause", "Stop", "Seek", "NextTrack", "PreviousTrack"},
				SupportsMediaControl: true,
				SupportsPersistentIdentifier: true,
				SupportsSync:         false,
				DeviceProfile: gin.H{
					"Name": authInfo.Device,
					"SupportedMediaTypes": []string{"Audio", "Video"},
					"DirectPlayProfiles": []interface{}{},
					"TranscodingProfiles": []interface{}{},
					"ContainerProfiles": []interface{}{},
					"CodecProfiles": []interface{}{},
					"SubtitleProfiles": []interface{}{},
				},
				AppStoreUrl: "",
				IconUrl:     "",
			},
			AdditionalUsers: []interface{}{},
		},
		AccessToken: token,
		ServerId:    ServerUUID,
	}

	c.JSON(http.StatusOK, authResult)
}

func GetCurrentUser(c *gin.Context) {
	var user models.User
	database.DB.Where("username = ?", "admin").First(&user)

	now := time.Now().UTC().Format(time.RFC3339)
	c.JSON(http.StatusOK, UserDto{
		Name:                      user.Username,
		ServerId:                  ServerUUID,
		ServerName:                "Gellyte",
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
		PrimaryImageTag:           "",
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
		"PrimaryImageTag":           "",
		"HasPassword":               true,
		"HasConfiguredPassword":     true,
		"HasConfiguredEasyPassword": true,
		"EnableAutoLogin":           true,
		"PrimaryImageAspectRatio":   1.0,
		"Policy":                    getDefaultPolicyDto(user.IsAdmin),
		"Configuration":             getDefaultConfigurationDto(),
	})
}

func GetUserViews(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Items": []gin.H{{
			"Name":            "Películas",
			"ServerId":        ServerUUID,
			"Id":              "12345678-1234-1234-1234-123456789012",
			"Type":            "UserView",
			"CollectionType":  "movies",
			"ImageTags":       gin.H{},
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
		EnableLyricManagement:            true,
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
		EnableLiveTvManagement:           true,
		EnableLiveTvAccess:               true,
		EnableMediaPlayback:              true,
		EnableAudioPlaybackTranscoding:   true,
		EnableVideoPlaybackTranscoding:   true,
		EnablePlaybackRemuxing:           true,
		ForceRemoteSourceTranscoding:     false,
		EnableContentDeletion:            true,
		EnableContentDeletionFromFolders: []string{},
		EnableContentDownloading:         true,
		EnableSyncTranscoding:            true,
		EnableMediaConversion:            true,
		EnabledDevices:                   []string{},
		EnableAllDevices:                 true,
		EnabledChannels:                  []string{},
		EnableAllChannels:                true,
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
