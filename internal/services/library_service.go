package services

import (
	"github.com/gellyte/gellyte/internal/models"
	"github.com/gellyte/gellyte/internal/repository"
)

type LibraryService interface {
	GetItems(params GetItemsParams) ([]models.MediaItem, int64, error)
	GetItem(id string) (*models.MediaItem, error)
	GetUserData(userID, itemID string) (*models.UserItemData, error)
	GetLatestItems(limit int, itemTypes []string) ([]models.MediaItem, error)
	GetResumeItems(userID string, limit int) ([]models.MediaItem, error)
	GetNextUpItems(userID string, limit int) ([]models.MediaItem, error)
	GetSuggestions(userID string, limit int) ([]models.MediaItem, error)
}

type GetItemsParams struct {
	ParentID   string
	ItemTypes  []string
	SearchTerm string
	IDs        []string
	StartIndex int
	Limit      int
}

type libraryService struct {
	mediaRepo    repository.MediaRepository
	userDataRepo repository.UserItemDataRepository
}

func NewLibraryService(mediaRepo repository.MediaRepository, userDataRepo repository.UserItemDataRepository) LibraryService {
	return &libraryService{
		mediaRepo:    mediaRepo,
		userDataRepo: userDataRepo,
	}
}

func (s *libraryService) GetItems(p GetItemsParams) ([]models.MediaItem, int64, error) {
	if len(p.IDs) > 0 {
		items, err := s.mediaRepo.GetByIDs(p.IDs)
		return items, int64(len(items)), err
	}

	if p.ParentID != "" {
		items, err := s.mediaRepo.GetByParentID(p.ParentID)
		if err == nil && len(items) > 0 {
			return items, int64(len(items)), nil
		}
	}

	return s.mediaRepo.Search(p.SearchTerm, p.ItemTypes, p.StartIndex, p.Limit)
}

func (s *libraryService) GetItem(id string) (*models.MediaItem, error) {
	return s.mediaRepo.GetByID(id)
}

func (s *libraryService) GetUserData(userID, itemID string) (*models.UserItemData, error) {
	return s.userDataRepo.Get(userID, itemID)
}

func (s *libraryService) GetLatestItems(limit int, itemTypes []string) ([]models.MediaItem, error) {
	return s.mediaRepo.GetLatest(limit, itemTypes)
}

func (s *libraryService) GetResumeItems(userID string, limit int) ([]models.MediaItem, error) {
	userDatas, err := s.userDataRepo.GetResume(userID)
	if err != nil {
		return nil, err
	}

	var items []models.MediaItem
	for _, ud := range userDatas {
		if item, err := s.mediaRepo.GetByID(ud.MediaItemID); err == nil {
			items = append(items, *item)
			if len(items) >= limit {
				break
			}
		}
	}
	return items, nil
}

func (s *libraryService) GetNextUpItems(userID string, limit int) ([]models.MediaItem, error) {
	played, err := s.userDataRepo.GetPlayed(userID)
	if err != nil {
		return nil, err
	}

	seenSeries := make(map[string]bool)
	var nextUpItems []models.MediaItem

	for _, ud := range played {
		lastEpisode, err := s.mediaRepo.GetByID(ud.MediaItemID)
		if err != nil || lastEpisode.Type != "Episode" {
			continue
		}

		// Obtener serie padre
		seriesID := ""
		parent, err := s.mediaRepo.GetByID(lastEpisode.ParentID)
		if err == nil {
			if parent.Type == "Season" {
				seriesID = parent.ParentID
			} else {
				seriesID = parent.ID
			}
		}

		if seriesID != "" && seenSeries[seriesID] {
			continue
		}
		if seriesID != "" {
			seenSeries[seriesID] = true
		}

		nextEpisode, err := s.mediaRepo.GetNextEpisode(lastEpisode.ParentID, lastEpisode.Name)
		if err == nil {
			// Comprobar si ya fue visto
			if data, err := s.userDataRepo.Get(userID, nextEpisode.ID); err != nil || !data.Played {
				nextUpItems = append(nextUpItems, *nextEpisode)
				if len(nextUpItems) >= limit {
					break
				}
			}
		}
	}

	return nextUpItems, nil
}

func (s *libraryService) GetSuggestions(userID string, limit int) ([]models.MediaItem, error) {
	// Por ahora devolvemos los items más recientes como sugerencia
	return s.mediaRepo.GetLatest(limit, nil)
}
