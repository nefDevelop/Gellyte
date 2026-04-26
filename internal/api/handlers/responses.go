package handlers

// Estos modelos están sincronizados 1:1 con el esquema 10.11.8 proporcionado por el usuario.
// Se mantiene el formato PascalCase y se eliminan punteros para garantizar que nunca haya nulls en campos clave.

type PublicSystemInfo struct {
	LocalAddress           string `json:"LocalAddress"`
	ServerName             string `json:"ServerName"`
	Version                string `json:"Version"`
	ProductName            string `json:"ProductName"`
	OperatingSystem        string `json:"OperatingSystem"`
	Id                     string `json:"Id"`
	StartupWizardCompleted bool   `json:"StartupWizardCompleted"`
}

type SystemInfo struct {
	LocalAddress               string        `json:"LocalAddress"`
	ServerName                 string        `json:"ServerName"`
	Version                    string        `json:"Version"`
	ProductName                string        `json:"ProductName"`
	OperatingSystem            string        `json:"OperatingSystem"`
	Id                         string        `json:"Id"`
	StartupWizardCompleted     bool          `json:"StartupWizardCompleted"`
	OperatingSystemDisplayName  string        `json:"OperatingSystemDisplayName"`
	PackageName                string        `json:"PackageName"`
	HasPendingRestart          bool          `json:"HasPendingRestart"`
	IsShuttingDown             bool          `json:"IsShuttingDown"`
	SupportsLibraryMonitor     bool          `json:"SupportsLibraryMonitor"`
	WebSocketPortNumber        int           `json:"WebSocketPortNumber"`
	CompletedInstallations     []interface{} `json:"CompletedInstallations"`
	CanSelfRestart             bool          `json:"CanSelfRestart"`
	CanLaunchWebBrowser        bool          `json:"CanLaunchWebBrowser"`
	ProgramDataPath            string        `json:"ProgramDataPath"`
	WebPath                    string        `json:"WebPath"`
	ItemsByNamePath            string        `json:"ItemsByNamePath"`
	CachePath                  string        `json:"CachePath"`
	LogPath                    string        `json:"LogPath"`
	InternalMetadataPath       string        `json:"InternalMetadataPath"`
	TranscodingTempPath        string        `json:"TranscodingTempPath"`
	CastReceiverApplications   []interface{} `json:"CastReceiverApplications"`
	HasUpdateAvailable         bool          `json:"HasUpdateAvailable"`
	EncoderLocation            string        `json:"EncoderLocation"`
	SystemArchitecture         string        `json:"SystemArchitecture"`
}

type UserDto struct {
	Name                      string            `json:"Name"`
	ServerId                  string            `json:"ServerId"`
	ServerName                string            `json:"ServerName"`
	Id                        string            `json:"Id"`
	PrimaryImageTag           string            `json:"PrimaryImageTag"`
	HasPassword               bool              `json:"HasPassword"`
	HasConfiguredPassword     bool              `json:"HasConfiguredPassword"`
	HasConfiguredEasyPassword bool              `json:"HasConfiguredEasyPassword"`
	EnableAutoLogin           bool              `json:"EnableAutoLogin"`
	LastLoginDate             string            `json:"LastLoginDate"`
	LastActivityDate          string            `json:"LastActivityDate"`
	Configuration             UserConfiguration `json:"Configuration"`
	Policy                    UserPolicy        `json:"Policy"`
	PrimaryImageAspectRatio   float64           `json:"PrimaryImageAspectRatio"`
}

type UserConfiguration struct {
	AudioLanguagePreference    string   `json:"AudioLanguagePreference"`
	PlayDefaultAudioTrack      bool     `json:"PlayDefaultAudioTrack"`
	SubtitleLanguagePreference string   `json:"SubtitleLanguagePreference"`
	DisplayMissingEpisodes     bool     `json:"DisplayMissingEpisodes"`
	GroupedFolders             []string `json:"GroupedFolders"`
	SubtitleMode               string   `json:"SubtitleMode"`
	DisplayCollectionsView     bool     `json:"DisplayCollectionsView"`
	EnableLocalPassword        bool     `json:"EnableLocalPassword"`
	OrderedViews               []string `json:"OrderedViews"`
	LatestItemsExcludes        []string `json:"LatestItemsExcludes"`
	MyMediaExcludes            []string `json:"MyMediaExcludes"`
	HidePlayedInLatest         bool     `json:"HidePlayedInLatest"`
	RememberAudioSelections    bool     `json:"RememberAudioSelections"`
	RememberSubtitleSelections bool     `json:"RememberSubtitleSelections"`
	EnableNextEpisodeAutoPlay  bool     `json:"EnableNextEpisodeAutoPlay"`
	CastReceiverId             string   `json:"CastReceiverId"`
	ResumePlayerState          bool     `json:"ResumePlayerState"`
	SyncPlayLikes              bool     `json:"SyncPlayLikes"`
	EnableCinemaMode           bool     `json:"EnableCinemaMode"`
	HidePlayedInSongs          bool     `json:"HidePlayedInSongs"`
	HidePlayedInVideos         bool     `json:"HidePlayedInVideos"`
	SkipSongsNotPlayed         bool     `json:"SkipSongsNotPlayed"`
}

type UserPolicy struct {
	IsAdministrator                  bool     `json:"IsAdministrator"`
	IsHidden                         bool     `json:"IsHidden"`
	EnableCollectionManagement       bool     `json:"EnableCollectionManagement"`
	EnableSubtitleManagement         bool     `json:"EnableSubtitleManagement"`
	EnableLyricManagement            bool     `json:"EnableLyricManagement"`
	IsDisabled                       bool     `json:"IsDisabled"`
	MaxParentalRating                int      `json:"MaxParentalRating"`
	MaxParentalSubRating             int      `json:"MaxParentalSubRating"`
	BlockedTags                      []string `json:"BlockedTags"`
	AllowedTags                      []string `json:"AllowedTags"`
	EnableUserPreferenceAccess       bool     `json:"EnableUserPreferenceAccess"`
	AccessSchedules                  []interface{} `json:"AccessSchedules"`
	BlockUnratedItems                []interface{} `json:"BlockUnratedItems"`
	EnableRemoteControlOfOtherUsers  bool     `json:"EnableRemoteControlOfOtherUsers"`
	EnableSharedDeviceControl        bool     `json:"EnableSharedDeviceControl"`
	EnableRemoteAccess               bool     `json:"EnableRemoteAccess"`
	EnableLiveTvManagement           bool     `json:"EnableLiveTvManagement"`
	EnableLiveTvAccess               bool     `json:"EnableLiveTvAccess"`
	EnableMediaPlayback              bool     `json:"EnableMediaPlayback"`
	EnableAudioPlaybackTranscoding   bool     `json:"EnableAudioPlaybackTranscoding"`
	EnableVideoPlaybackTranscoding   bool     `json:"EnableVideoPlaybackTranscoding"`
	EnablePlaybackRemuxing           bool     `json:"EnablePlaybackRemuxing"`
	ForceRemoteSourceTranscoding     bool     `json:"ForceRemoteSourceTranscoding"`
	EnableContentDeletion            bool     `json:"EnableContentDeletion"`
	EnableContentDeletionFromFolders []string `json:"EnableContentDeletionFromFolders"`
	EnableContentDownloading         bool     `json:"EnableContentDownloading"`
	EnableSyncTranscoding            bool     `json:"EnableSyncTranscoding"`
	EnableMediaConversion            bool     `json:"EnableMediaConversion"`
	EnabledDevices                   []string `json:"EnabledDevices"`
	EnableAllDevices                 bool     `json:"EnableAllDevices"`
	EnabledChannels                  []string `json:"EnabledChannels"`
	EnableAllChannels                bool     `json:"EnableAllChannels"`
	EnabledFolders                   []string `json:"EnabledFolders"`
	EnableAllFolders                 bool     `json:"EnableAllFolders"`
	InvalidLoginAttemptCount         int      `json:"InvalidLoginAttemptCount"`
	LoginAttemptsBeforeLockout       int      `json:"LoginAttemptsBeforeLockout"`
	MaxActiveSessions                int      `json:"MaxActiveSessions"`
	EnablePublicSharing              bool     `json:"EnablePublicSharing"`
	EnableSubtitleDownloading        bool     `json:"EnableSubtitleDownloading"`
	EnablePlaybackStreaming          bool     `json:"EnablePlaybackStreaming"`
	EnableSharedDevice               bool     `json:"EnableSharedDevice"`
	BlockedMediaFolders              []string `json:"BlockedMediaFolders"`
	BlockedChannels                  []string `json:"BlockedChannels"`
	RemoteClientBitrateLimit         int      `json:"RemoteClientBitrateLimit"`
	AuthenticationProviderId         string   `json:"AuthenticationProviderId"`
	PasswordResetProviderId          string   `json:"PasswordResetProviderId"`
	SyncPlayAccess                   string   `json:"SyncPlayAccess"`
}

type AuthenticationResult struct {
	User        UserDto         `json:"User"`
	SessionInfo *SessionInfoDto `json:"SessionInfo"`
	AccessToken string          `json:"AccessToken"`
	ServerId    string          `json:"ServerId"`
}

type SessionInfoDto struct {
	PlayState                PlayerStateInfo    `json:"PlayState"`
	AdditionalUsers          []interface{}      `json:"AdditionalUsers"`
	Capabilities             ClientCapabilities `json:"Capabilities"`
	RemoteEndPoint           string             `json:"RemoteEndPoint"`
	PlayableMediaTypes       []string           `json:"PlayableMediaTypes"`
	Id                       string             `json:"Id"`
	UserId                   string             `json:"UserId"`
	UserName                 string             `json:"UserName"`
	Client                   string             `json:"Client"`
	LastActivityDate         string             `json:"LastActivityDate"`
	LastPlaybackCheckIn      string             `json:"LastPlaybackCheckIn"`
	LastPausedDate           *string            `json:"LastPausedDate"`
	DeviceName               string             `json:"DeviceName"`
	DeviceType               string             `json:"DeviceType"`
	DeviceId                 string             `json:"DeviceId"`
	ApplicationVersion       string             `json:"ApplicationVersion"`
	TranscodingInfo          *TranscodingInfo   `json:"TranscodingInfo"`
	IsActive                 bool               `json:"IsActive"`
	SupportsMediaControl     bool               `json:"SupportsMediaControl"`
	SupportsRemoteControl    bool               `json:"SupportsRemoteControl"`
	NowPlayingItem           interface{}        `json:"NowPlayingItem"`
	NowViewingItem           interface{}        `json:"NowViewingItem"`
	HasCustomDeviceName      bool               `json:"HasCustomDeviceName"`
	PlaylistItemId           string             `json:"PlaylistItemId"`
	ServerId                 string             `json:"ServerId"`
	UserPrimaryImageTag      string             `json:"UserPrimaryImageTag"`
	SupportedCommands        []string           `json:"SupportedCommands"`
	NowPlayingQueue          []interface{}      `json:"NowPlayingQueue"`
	NowPlayingQueueFullItems []interface{}      `json:"NowPlayingQueueFullItems"`
}

type TranscodingInfo struct {
	AudioCodec               string  `json:"AudioCodec"`
	VideoCodec               string  `json:"VideoCodec"`
	Container                string  `json:"Container"`
	IsVideoDirect            bool    `json:"IsVideoDirect"`
	IsAudioDirect            bool    `json:"IsAudioDirect"`
	Bitrate                  int     `json:"Bitrate"`
	Framerate                float64 `json:"Framerate"`
	CompletionPercentage     float64 `json:"CompletionPercentage"`
	Width                    int     `json:"Width"`
	Height                   int     `json:"Height"`
	AudioChannels            int     `json:"AudioChannels"`
	HardwareAccelerationType string  `json:"HardwareAccelerationType"`
	TranscodeReasons         string  `json:"TranscodeReasons"`
}

type PlayerStateInfo struct {
	PositionTicks       int64  `json:"PositionTicks"`
	CanSeek             bool   `json:"CanSeek"`
	IsPaused            bool   `json:"IsPaused"`
	IsMuted             bool   `json:"IsMuted"`
	VolumeLevel         int    `json:"VolumeLevel"`
	AudioStreamIndex    int    `json:"AudioStreamIndex"`
	SubtitleStreamIndex int    `json:"SubtitleStreamIndex"`
	PlayMethod          string `json:"PlayMethod"`
	RepeatMode          string `json:"RepeatMode"`
}

type ClientCapabilities struct {
	PlayableMediaTypes           []string    `json:"PlayableMediaTypes"`
	SupportedCommands            []string    `json:"SupportedCommands"`
	SupportsMediaControl         bool        `json:"SupportsMediaControl"`
	SupportsPersistentIdentifier bool        `json:"SupportsPersistentIdentifier"`
	SupportsSync                 bool        `json:"SupportsSync"`
	DeviceProfile                interface{} `json:"DeviceProfile"`
	AppStoreUrl                  string      `json:"AppStoreUrl"`
	IconUrl                      string      `json:"IconUrl"`
}

type BaseItemDto struct {
	Name                    string                 `json:"Name"`
	Id                      string                 `json:"Id"`
	ServerId                string                 `json:"ServerId"`
	Type                    string                 `json:"Type"`
	MediaType               string                 `json:"MediaType"`
	IsFolder                bool                   `json:"IsFolder"`
	CanDelete               bool                   `json:"CanDelete"`
	PlayAccess              string                 `json:"PlayAccess"`
	PrimaryImageAspectRatio float64                `json:"PrimaryImageAspectRatio"`
	ImageTags               map[string]string      `json:"ImageTags"`
	UserData                UserItemDataDto        `json:"UserData"`
	CollectionType          string                 `json:"CollectionType"`
	Path                    string                 `json:"Path"`
	ParentId                string                 `json:"ParentId"`
	Width                   int                    `json:"Width,omitempty"`
	Height                  int                    `json:"Height,omitempty"`
	Overview                string                 `json:"Overview"`
	RunTimeTicks            int64                  `json:"RunTimeTicks"`
	ProductionYear          int                    `json:"ProductionYear"`
	IndexNumber             int                    `json:"IndexNumber,omitempty"`
	ParentIndexNumber       int                    `json:"ParentIndexNumber,omitempty"`
	SeriesName              string                 `json:"SeriesName,omitempty"`
	SeriesId                string                 `json:"SeriesId,omitempty"`
	SeasonId                string                 `json:"SeasonId,omitempty"`
	SeasonName              string                 `json:"SeasonName,omitempty"`
	ExternalUrls            []interface{}          `json:"ExternalUrls"`
	MediaSources            []interface{}          `json:"MediaSources"`
	ImageBlurHashes         map[string]interface{} `json:"ImageBlurHashes"`
}

type UserItemDataDto struct {
	PlaybackPositionTicks int64   `json:"PlaybackPositionTicks"`
	PlayCount             int     `json:"PlayCount"`
	IsFavorite            bool    `json:"IsFavorite"`
	Played                bool    `json:"Played"`
	LastPlayedDate        string  `json:"LastPlayedDate"`
	Rating                float64 `json:"Rating"`
}

type BaseItemDtoQueryResult struct {
	Items            []BaseItemDto `json:"Items"`
	TotalRecordCount int           `json:"TotalRecordCount"`
	StartIndex       int           `json:"StartIndex"`
}

type VirtualFolderDto struct {
	Name               string         `json:"Name"`
	Locations          []string       `json:"Locations"`
	CollectionType     string         `json:"CollectionType"`
	LibraryOptions     LibraryOptions `json:"LibraryOptions"`
	ItemId             string         `json:"ItemId"`
	PrimaryImageItemId string         `json:"PrimaryImageItemId"`
	RefreshProgress    *float64       `json:"RefreshProgress"`
	RefreshStatus      *string        `json:"RefreshStatus"`
}

type LibraryOptions struct {
	Enabled                                 bool          `json:"Enabled"`
	EnablePhotos                            bool          `json:"EnablePhotos"`
	EnableRealtimeMonitor                   bool          `json:"EnableRealtimeMonitor"`
	EnableLUFSScan                          bool          `json:"EnableLUFSScan"`
	EnableChapterImageExtraction            bool          `json:"EnableChapterImageExtraction"`
	ExtractChapterImagesDuringLibraryScan   bool          `json:"ExtractChapterImagesDuringLibraryScan"`
	EnableTrickplayImageExtraction          bool          `json:"EnableTrickplayImageExtraction"`
	ExtractTrickplayImagesDuringLibraryScan bool          `json:"ExtractTrickplayImagesDuringLibraryScan"`
	PathInfos                               []PathInfo    `json:"PathInfos"`
	SaveLocalMetadata                       bool          `json:"SaveLocalMetadata"`
	EnableInternetProviders                 bool          `json:"EnableInternetProviders"`
	EnableAutomaticSeriesGrouping           bool          `json:"EnableAutomaticSeriesGrouping"`
	EnableEmbeddedTitles                    bool          `json:"EnableEmbeddedTitles"`
	EnableEmbeddedExtrasTitles              bool          `json:"EnableEmbeddedExtrasTitles"`
	EnableEmbeddedEpisodeInfos              bool          `json:"EnableEmbeddedEpisodeInfos"`
	AutomaticRefreshIntervalDays            int           `json:"AutomaticRefreshIntervalDays"`
	PreferredMetadataLanguage               string        `json:"PreferredMetadataLanguage"`
	MetadataCountryCode                     string        `json:"MetadataCountryCode"`
	SeasonZeroDisplayName                   string        `json:"SeasonZeroDisplayName"`
	MetadataSavers                          []string      `json:"MetadataSavers"`
	DisabledLocalMetadataReaders            []string      `json:"DisabledLocalMetadataReaders"`
	LocalMetadataReaderOrder                []string      `json:"LocalMetadataReaderOrder"`
	DisabledSubtitleFetchers                []string      `json:"DisabledSubtitleFetchers"`
	SubtitleFetcherOrder                    []string      `json:"SubtitleFetcherOrder"`
	SkipSubtitlesIfEmbeddedSubtitlesPresent bool          `json:"SkipSubtitlesIfEmbeddedSubtitlesPresent"`
	SkipSubtitlesIfAudioTrackMatches        bool          `json:"SkipSubtitlesIfAudioTrackMatches"`
	SubtitleDownloadLanguages               []string      `json:"SubtitleDownloadLanguages"`
	RequirePerfectSubtitleMatch             bool          `json:"RequirePerfectSubtitleMatch"`
	SaveSubtitlesWithMedia                  bool          `json:"SaveSubtitlesWithMedia"`
	AutomaticallyAddToCollection            bool          `json:"AutomaticallyAddToCollection"`
	AllowEmbeddedSubtitles                  string        `json:"AllowEmbeddedSubtitles"`
	TypeOptions                             []TypeOptions `json:"TypeOptions"`
}

type PathInfo struct {
	Path string `json:"Path"`
}

type TypeOptions struct {
	Type                 string         `json:"Type"`
	MetadataFetchers     []string       `json:"MetadataFetchers"`
	MetadataFetcherOrder []string       `json:"MetadataFetcherOrder"`
	ImageFetchers        []string       `json:"ImageFetchers"`
	ImageFetcherOrder    []string       `json:"ImageFetcherOrder"`
	ImageOptions         []ImageOptions `json:"ImageOptions"`
}

type ImageOptions struct {
	Type     string `json:"Type"`
	Limit    int    `json:"Limit"`
	MinWidth int    `json:"MinWidth"`
}
