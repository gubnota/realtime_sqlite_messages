package model

import "time"

type Leaderboard struct {
	UserID      string    `gorm:"primaryKey"`
	Score       int       `gorm:"default:0;not null"`
	LastUpdated time.Time `gorm:"index;autoUpdateTime"`
}

func (Leaderboard) TableName() string {
	return "leaderboards"
}
