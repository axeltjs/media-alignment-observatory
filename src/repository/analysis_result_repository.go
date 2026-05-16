package repository

import (
	"context"

	"github.com/hafidluqman50/maoi/src/model"
	"gorm.io/gorm"
)

type AnalysisResultRepository struct {
	DB *gorm.DB
}

type AnalysisResultCreateInput struct {
	ArticleID           int64
	GovernmentContentID int64
	SimilarityScore     float64
	AlignmentLabel      string
}

func (r *AnalysisResultRepository) Create(ctx context.Context, input AnalysisResultCreateInput) (int64, error) {
	result := model.AnalysisResult{
		ArticleID:           input.ArticleID,
		GovernmentContentID: input.GovernmentContentID,
		SimilarityScore:     input.SimilarityScore,
		AlignmentLabel:      input.AlignmentLabel,
	}
	if err := r.DB.WithContext(ctx).Create(&result).Error; err != nil {
		return 0, err
	}
	return result.ID, nil
}

func (r *AnalysisResultRepository) ExistsByArticleAndGovContent(ctx context.Context, articleID, govContentID int64) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&model.AnalysisResult{}).
		Where("article_id = ? AND government_content_id = ?", articleID, govContentID).
		Count(&count).Error
	return count > 0, err
}

func (r *AnalysisResultRepository) FindByArticleID(ctx context.Context, articleID int64) ([]model.AnalysisResult, error) {
	var results []model.AnalysisResult
	err := r.DB.WithContext(ctx).
		Preload("GovernmentContent").
		Where("article_id = ?", articleID).
		Order("similarity_score DESC").
		Find(&results).Error
	return results, err
}

func (r *AnalysisResultRepository) FindTopAligned(ctx context.Context, limit int) ([]model.AnalysisResult, error) {
	var results []model.AnalysisResult
	err := r.DB.WithContext(ctx).
		Preload("Article").
		Preload("Article.Source").
		Preload("GovernmentContent").
		Where("alignment_label = ?", "high").
		Order("similarity_score DESC").
		Limit(limit).
		Find(&results).Error
	return results, err
}

type AlignmentStats struct {
	Label string
	Count int64
	Avg   float64
}

func (r *AnalysisResultRepository) GetAlignmentStats(ctx context.Context) ([]AlignmentStats, error) {
	var stats []AlignmentStats
	err := r.DB.WithContext(ctx).Model(&model.AnalysisResult{}).
		Select("alignment_label as label, count(*) as count, avg(similarity_score) as avg").
		Group("alignment_label").
		Scan(&stats).Error
	return stats, err
}
