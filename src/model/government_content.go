package model

import "time"

type GovernmentContent struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Title         string    `gorm:"column:title;not null"`
	URL           string    `gorm:"column:url;not null;uniqueIndex"`
	Content       string    `gorm:"column:content;type:text"`
	CleanContent  string    `gorm:"column:clean_content;type:text"`
	EmbeddingJSON string    `gorm:"column:embedding_json;type:text"`
	Agency        string    `gorm:"column:agency"`
	PublishedAt   time.Time `gorm:"column:published_at;index"`
	FetchedAt     time.Time `gorm:"column:fetched_at;autoCreateTime"`
}

func (GovernmentContent) TableName() string {
	return "government_contents"
}
