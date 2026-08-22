package model

import "time"

// CoordinationCluster captures a group of articles from different sources
// that publish highly similar content within a short time window — a signal
// of possible coordinated narrative amplification.
type CoordinationCluster struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	WindowStart     time.Time `gorm:"column:window_start;not null;index"`
	WindowEnd       time.Time `gorm:"column:window_end;not null"`
	ArticleCount    int       `gorm:"column:article_count;not null"`
	AvgSimilarity   float64   `gorm:"column:avg_similarity;not null"`
	Topic           string    `gorm:"column:topic;type:text"`
	ArticleIDsJSON  string    `gorm:"column:article_ids_json;type:text"`
	DetectedAt      time.Time `gorm:"column:detected_at;autoCreateTime"`
}

func (CoordinationCluster) TableName() string {
	return "coordination_clusters"
}
