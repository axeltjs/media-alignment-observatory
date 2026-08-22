package repository

import (
	"context"
	"errors"
	"time"

	"github.com/hafidluqman50/maoi/src/model"
	"gorm.io/gorm"
)

type ArticleRepository struct {
	DB *gorm.DB
}

type ArticleCreateInput struct {
	SourceID      int64
	Title         string
	URL           string
	Content       string
	CleanContent  string
	EmbeddingJSON string
	PublishedAt   time.Time
}

func (r *ArticleRepository) Create(ctx context.Context, input ArticleCreateInput) (int64, error) {
	article := model.Article{
		SourceID:      input.SourceID,
		Title:         input.Title,
		URL:           input.URL,
		Content:       input.Content,
		CleanContent:  input.CleanContent,
		EmbeddingJSON: input.EmbeddingJSON,
		PublishedAt:   input.PublishedAt,
	}
	if err := r.DB.WithContext(ctx).Create(&article).Error; err != nil {
		return 0, err
	}
	return article.ID, nil
}

func (r *ArticleRepository) ExistsByURL(ctx context.Context, url string) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&model.Article{}).Where("url = ?", url).Count(&count).Error
	return count > 0, err
}

func (r *ArticleRepository) FindByID(ctx context.Context, id int64) (model.Article, bool, error) {
	var article model.Article
	err := r.DB.WithContext(ctx).Preload("Source").First(&article, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Article{}, false, nil
	}
	return article, err == nil, err
}

func (r *ArticleRepository) FindBySourceID(ctx context.Context, sourceID int64, limit, offset int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64
	q := r.DB.WithContext(ctx).Model(&model.Article{}).Where("source_id = ?", sourceID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("published_at DESC").Limit(limit).Offset(offset).Find(&articles).Error
	return articles, total, err
}

func (r *ArticleRepository) FindWithoutEmbedding(ctx context.Context, limit int) ([]model.Article, error) {
	var articles []model.Article
	err := r.DB.WithContext(ctx).
		Where("embedding_json IS NULL OR embedding_json = ''").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func (r *ArticleRepository) FindInTimeWindow(ctx context.Context, from, to time.Time) ([]model.Article, error) {
	var articles []model.Article
	err := r.DB.WithContext(ctx).
		Preload("Source").
		Where("published_at BETWEEN ? AND ?", from, to).
		Order("published_at ASC").
		Find(&articles).Error
	return articles, err
}

func (r *ArticleRepository) UpdateEmbedding(ctx context.Context, id int64, embeddingJSON, cleanContent string) error {
	return r.DB.WithContext(ctx).Model(&model.Article{}).Where("id = ?", id).Updates(map[string]any{
		"embedding_json": embeddingJSON,
		"clean_content":  cleanContent,
	}).Error
}

func (r *ArticleRepository) FindRecent(ctx context.Context, limit int) ([]model.Article, error) {
	var articles []model.Article
	err := r.DB.WithContext(ctx).
		Preload("Source").
		Order("published_at DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}
