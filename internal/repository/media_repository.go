package repository

import (
	"github.com/gellyte/gellyte/internal/models"
	"gorm.io/gorm"
)

type MediaRepository interface {
	GetByID(id string) (*models.MediaItem, error)
	GetByParentID(parentID string) ([]models.MediaItem, error)
	Search(term string, types []string, startIndex, limit int) ([]models.MediaItem, int64, error)
	GetByIDs(ids []string) ([]models.MediaItem, error)
	GetByType(itemType string, limit int) ([]models.MediaItem, error)
	GetLatest(limit int, itemTypes []string) ([]models.MediaItem, error)
	GetNextEpisode(parentID, currentName string) (*models.MediaItem, error)
	Count(itemType string) (int64, error)
}

type GormMediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &GormMediaRepository{db: db}
}

func (r *GormMediaRepository) GetByID(id string) (*models.MediaItem, error) {
	var item models.MediaItem
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormMediaRepository) GetByParentID(parentID string) ([]models.MediaItem, error) {
	var items []models.MediaItem
	if err := r.db.Where("parent_id = ?", parentID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *GormMediaRepository) Search(term string, types []string, startIndex, limit int) ([]models.MediaItem, int64, error) {
	var items []models.MediaItem
	var total int64
	query := r.db.Model(&models.MediaItem{})

	if term != "" {
		query = query.Where("name LIKE ?", "%"+term+"%")
	}
	if len(types) > 0 {
		query = query.Where("type IN ?", types)
	}

	query.Count(&total)
	if err := query.Offset(startIndex).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *GormMediaRepository) GetByIDs(ids []string) ([]models.MediaItem, error) {
	var items []models.MediaItem
	if err := r.db.Where("id IN ?", ids).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *GormMediaRepository) GetByType(itemType string, limit int) ([]models.MediaItem, error) {
	var items []models.MediaItem
	query := r.db.Where("type = ?", itemType)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *GormMediaRepository) GetLatest(limit int, itemTypes []string) ([]models.MediaItem, error) {
	var items []models.MediaItem
	query := r.db.Order("created_at desc")
	if len(itemTypes) > 0 {
		query = query.Where("type IN ?", itemTypes)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *GormMediaRepository) GetNextEpisode(parentID, currentName string) (*models.MediaItem, error) {
	var item models.MediaItem
	if err := r.db.Where("parent_id = ? AND name > ? AND type = ?", parentID, currentName, "Episode").Order("name asc").First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormMediaRepository) Count(itemType string) (int64, error) {
	var count int64
	query := r.db.Model(&models.MediaItem{})
	if itemType != "" {
		query = query.Where("type = ?", itemType)
	}
	query.Count(&count)
	return count, nil
}
