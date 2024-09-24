package models

import "time"

type UrlModel struct {
	ID       int64  `gorm:"column:id;type:BIGSERIAL;primaryKey"`
	ShortUrl string `gorm:"column:short_url;type:VARCHAR(7);not null;uniqueIndex"`
	LongUrl  string `gorm:"column:long_url;type:VARCHAR(400);not null;index"`
	// UserId    string    `gorm:"column:user_id;type:VARCHAR(50);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for this model
func (UrlModel) TableName() string {
	return "tinyurl"
}
