package model

import "time"

type Result struct {
	UserID      string    `gorm:"primaryKey"`
	Score       int       `gorm:"default:0;not null"`
	LastUpdated time.Time `gorm:"index;autoUpdateTime"`
	game        uint      `gorm:"default:0"`
}

func (Result) TableName() string {
	return "results"
}
