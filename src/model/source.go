package model

import "time"

type Source struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string     `gorm:"column:name;not null"`
	RSSUrl      string     `gorm:"column:rss_url;not null;uniqueIndex"`
	BaseUrl     string     `gorm:"column:base_url;not null"`
	Category    string     `gorm:"column:category"`
	IsActive    bool       `gorm:"column:is_active;default:true"`
	LastFetched *time.Time `gorm:"column:last_fetched"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (Source) TableName() string {
	return "sources"
}
