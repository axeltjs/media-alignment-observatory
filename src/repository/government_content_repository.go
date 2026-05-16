package repository

import (
	"context"
	"errors"
	"time"

	"github.com/hafidluqman50/maoi/src/model"
	"gorm.io/gorm"
)

type GovernmentContentRepository struct {
	DB *gorm.DB
}

type GovernmentContentCreateInput struct {
	Title         string
	URL           string
	Content       string
	CleanContent  string
	EmbeddingJSON string
	Agency        string
	PublishedAt   time.Time
}

func (r *GovernmentContentRepository) Create(ctx context.Context, input GovernmentContentCreateInput) (int64, error) {
	gc := model.GovernmentContent{
		Title:         input.Title,
		URL:           input.URL,
		Content:       input.Content,
		CleanContent:  input.CleanContent,
		EmbeddingJSON: input.EmbeddingJSON,
		Agency:        input.Agency,
		PublishedAt:   input.PublishedAt,
	}
	if err := r.DB.WithContext(ctx).Create(&gc).Error; err != nil {
		return 0, err
	}
	return gc.ID, nil
}

func (r *GovernmentContentRepository) ExistsByURL(ctx context.Context, url string) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&model.GovernmentContent{}).Where("url = ?", url).Count(&count).Error
	return count > 0, err
}

func (r *GovernmentContentRepository) FindByID(ctx context.Context, id int64) (model.GovernmentContent, bool, error) {
	var gc model.GovernmentContent
	err := r.DB.WithContext(ctx).First(&gc, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.GovernmentContent{}, false, nil
	}
	return gc, err == nil, err
}

func (r *GovernmentContentRepository) FindWithEmbedding(ctx context.Context) ([]model.GovernmentContent, error) {
	var contents []model.GovernmentContent
	err := r.DB.WithContext(ctx).
		Where("embedding_json IS NOT NULL AND embedding_json != ''").
		Find(&contents).Error
	return contents, err
}

func (r *GovernmentContentRepository) FindWithoutEmbedding(ctx context.Context, limit int) ([]model.GovernmentContent, error) {
	var contents []model.GovernmentContent
	err := r.DB.WithContext(ctx).
		Where("embedding_json IS NULL OR embedding_json = ''").
		Limit(limit).
		Find(&contents).Error
	return contents, err
}

func (r *GovernmentContentRepository) UpdateEmbedding(ctx context.Context, id int64, embeddingJSON, cleanContent string) error {
	return r.DB.WithContext(ctx).Model(&model.GovernmentContent{}).Where("id = ?", id).Updates(map[string]any{
		"embedding_json": embeddingJSON,
		"clean_content":  cleanContent,
	}).Error
}

func (r *GovernmentContentRepository) FindAll(ctx context.Context, limit, offset int) ([]model.GovernmentContent, int64, error) {
	var contents []model.GovernmentContent
	var total int64
	q := r.DB.WithContext(ctx).Model(&model.GovernmentContent{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("published_at DESC").Limit(limit).Offset(offset).Find(&contents).Error
	return contents, total, err
}
