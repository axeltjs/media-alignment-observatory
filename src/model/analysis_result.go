package model

import "time"

type AnalysisResult struct {
	ID                  int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ArticleID           int64     `gorm:"column:article_id;not null;index"`
	GovernmentContentID int64     `gorm:"column:government_content_id;not null;index"`
	SimilarityScore     float64   `gorm:"column:similarity_score;not null"`
	AlignmentLabel      string    `gorm:"column:alignment_label"` // "high", "medium", "low"
	AnalyzedAt          time.Time `gorm:"column:analyzed_at;autoCreateTime"`
	Article             Article             `gorm:"foreignKey:ArticleID"`
	GovernmentContent   GovernmentContent   `gorm:"foreignKey:GovernmentContentID"`
}

func (AnalysisResult) TableName() string {
	return "analysis_results"
}
