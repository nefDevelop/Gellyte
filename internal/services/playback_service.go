package services

import (
	"os/exec"

	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/repository"
	"github.com/gellyte/gellyte/internal/transcoder"
)

type PlaybackService interface {
	GetItem(id string) (*models.MediaItem, error)
	StartTranscode(item models.MediaItem, opts transcoder.TranscodeOptions) *exec.Cmd
	GetHLSSegment(item models.MediaItem, segmentIndex int, segmentDuration int, opts transcoder.TranscodeOptions) *exec.Cmd
	UpdateProgress(userID, itemID string, ticks int64, isPaused bool) error
}

type playbackService struct {
	mediaRepo    repository.MediaRepository
	userDataRepo repository.UserItemDataRepository
}

func NewPlaybackService(mediaRepo repository.MediaRepository, userDataRepo repository.UserItemDataRepository) PlaybackService {
	return &playbackService{
		mediaRepo:    mediaRepo,
		userDataRepo: userDataRepo,
	}
}

func (s *playbackService) GetItem(id string) (*models.MediaItem, error) {
	return s.mediaRepo.GetByID(id)
}

func (s *playbackService) StartTranscode(item models.MediaItem, opts transcoder.TranscodeOptions) *exec.Cmd {
	return transcoder.BuildTranscodeCmd(item, opts)
}

func (s *playbackService) GetHLSSegment(item models.MediaItem, segmentIndex int, segmentDuration int, opts transcoder.TranscodeOptions) *exec.Cmd {
	return transcoder.BuildHLSSegmentCmd(item, segmentIndex, segmentDuration, opts)
}

func (s *playbackService) UpdateProgress(userID, itemID string, ticks int64, isPaused bool) error {
	data, err := s.userDataRepo.Get(userID, itemID)
	if err != nil {
		data = &models.UserItemData{
			UserID:      userID,
			MediaItemID: itemID,
		}
	}

	data.PlaybackPositionTicks = ticks
	// Si estamos cerca del final (>90% por ejemplo), podríamos marcar como reproducido
	// Pero por ahora solo actualizamos ticks.
	
	return s.userDataRepo.Upsert(data)
}
