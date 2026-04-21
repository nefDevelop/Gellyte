package repository

import (
	"github.com/gellyte/gellyte/internal/models"
	"gorm.io/gorm"
)

type UserItemDataRepository interface {
	Get(userID, mediaItemID string) (*models.UserItemData, error)
	Upsert(data *models.UserItemData) error
	GetResume(userID string) ([]models.UserItemData, error)
	GetPlayed(userID string) ([]models.UserItemData, error)
}

type GormUserItemDataRepository struct {
	db *gorm.DB
}

func NewUserItemDataRepository(db *gorm.DB) UserItemDataRepository {
	return &GormUserItemDataRepository{db: db}
}

func (r *GormUserItemDataRepository) Get(userID, mediaItemID string) (*models.UserItemData, error) {
	var data models.UserItemData
	if err := r.db.Where("user_id = ? AND media_item_id = ?", userID, mediaItemID).First(&data).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *GormUserItemDataRepository) Upsert(data *models.UserItemData) error {
	return r.db.Save(data).Error
}

func (r *GormUserItemDataRepository) GetResume(userID string) ([]models.UserItemData, error) {
	var results []models.UserItemData
	if err := r.db.Where("user_id = ? AND playback_position_ticks > 0 AND played = false", userID).Order("last_played_date desc").Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (r *GormUserItemDataRepository) GetPlayed(userID string) ([]models.UserItemData, error) {
	var results []models.UserItemData
	if err := r.db.Where("user_id = ? AND played = ?", userID, true).Order("last_played_date desc").Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
