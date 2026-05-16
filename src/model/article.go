package model

import "time"

type Article struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	SourceID        int64     `gorm:"column:source_id;not null;index"`
	Title           string    `gorm:"column:title;not null"`
	URL             string    `gorm:"column:url;not null;uniqueIndex"`
	Content         string    `gorm:"column:content;type:text"`
	CleanContent    string    `gorm:"column:clean_content;type:text"`
	EmbeddingJSON   string    `gorm:"column:embedding_json;type:text"`
	PublishedAt     time.Time `gorm:"column:published_at;index"`
	FetchedAt       time.Time `gorm:"column:fetched_at;autoCreateTime"`
	Source          Source    `gorm:"foreignKey:SourceID"`
}

func (Article) TableName() string {
	return "articles"
}
