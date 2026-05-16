package repository

import (
	"context"
	"time"

	"github.com/hafidluqman50/maoi/src/model"
	"gorm.io/gorm"
)

type CoordinationClusterRepository struct {
	DB *gorm.DB
}

type CoordinationClusterCreateInput struct {
	WindowStart    time.Time
	WindowEnd      time.Time
	ArticleCount   int
	AvgSimilarity  float64
	Topic          string
	ArticleIDsJSON string
}

func (r *CoordinationClusterRepository) Create(ctx context.Context, input CoordinationClusterCreateInput) (int64, error) {
	cluster := model.CoordinationCluster{
		WindowStart:    input.WindowStart,
		WindowEnd:      input.WindowEnd,
		ArticleCount:   input.ArticleCount,
		AvgSimilarity:  input.AvgSimilarity,
		Topic:          input.Topic,
		ArticleIDsJSON: input.ArticleIDsJSON,
	}
	if err := r.DB.WithContext(ctx).Create(&cluster).Error; err != nil {
		return 0, err
	}
	return cluster.ID, nil
}

func (r *CoordinationClusterRepository) FindRecent(ctx context.Context, limit int) ([]model.CoordinationCluster, error) {
	var clusters []model.CoordinationCluster
	err := r.DB.WithContext(ctx).
		Order("detected_at DESC").
		Limit(limit).
		Find(&clusters).Error
	return clusters, err
}

func (r *CoordinationClusterRepository) FindInDateRange(ctx context.Context, from, to time.Time) ([]model.CoordinationCluster, error) {
	var clusters []model.CoordinationCluster
	err := r.DB.WithContext(ctx).
		Where("window_start >= ? AND window_end <= ?", from, to).
		Order("window_start ASC").
		Find(&clusters).Error
	return clusters, err
}

func (r *CoordinationClusterRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&model.CoordinationCluster{}).Count(&count).Error
	return count, err
}
