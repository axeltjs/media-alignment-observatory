package repository

import (
	"context"
	"errors"
	"time"

	"github.com/hafidluqman50/maoi/src/model"
	"gorm.io/gorm"
)

type SourceRepository struct {
	DB *gorm.DB
}

type SourceCreateInput struct {
	Name     string
	RSSUrl   string
	BaseUrl  string
	Category string
}

type SourceUpdateInput struct {
	Name        *string
	RSSUrl      *string
	BaseUrl     *string
	Category    *string
	IsActive    *bool
	LastFetched *time.Time
}

func (r *SourceRepository) Create(ctx context.Context, input SourceCreateInput) (int64, error) {
	source := model.Source{
		Name:     input.Name,
		RSSUrl:   input.RSSUrl,
		BaseUrl:  input.BaseUrl,
		Category: input.Category,
		IsActive: true,
	}
	if err := r.DB.WithContext(ctx).Create(&source).Error; err != nil {
		return 0, err
	}
	return source.ID, nil
}

func (r *SourceRepository) FindAll(ctx context.Context) ([]model.Source, error) {
	var sources []model.Source
	err := r.DB.WithContext(ctx).Find(&sources).Error
	return sources, err
}

func (r *SourceRepository) FindActive(ctx context.Context) ([]model.Source, error) {
	var sources []model.Source
	err := r.DB.WithContext(ctx).Where("is_active = ?", true).Find(&sources).Error
	return sources, err
}

func (r *SourceRepository) FindByID(ctx context.Context, id int64) (model.Source, bool, error) {
	var source model.Source
	err := r.DB.WithContext(ctx).First(&source, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Source{}, false, nil
	}
	return source, err == nil, err
}

func (r *SourceRepository) Update(ctx context.Context, id int64, input SourceUpdateInput) error {
	updates := map[string]any{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.RSSUrl != nil {
		updates["rss_url"] = *input.RSSUrl
	}
	if input.BaseUrl != nil {
		updates["base_url"] = *input.BaseUrl
	}
	if input.Category != nil {
		updates["category"] = *input.Category
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if input.LastFetched != nil {
		updates["last_fetched"] = *input.LastFetched
	}
	return r.DB.WithContext(ctx).Model(&model.Source{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SourceRepository) Delete(ctx context.Context, id int64) error {
	return r.DB.WithContext(ctx).Delete(&model.Source{}, id).Error
}
