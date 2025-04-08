package model

import "time"

// Model
type Game struct {
	ID       uint      `gorm:"primaryKey"`
	Sender   string    `gorm:"index;not null"`
	Receiver string    `gorm:"index;not null"`
	Created  time.Time `gorm:"index;not null"`
	Svote    int       `gorm:"default:0"`
	Rvote    int       `gorm:"default:0"`
	Status   string    `gorm:"default:'open';check:status IN ('open', 'closed')"`
}

func (Game) TableName() string {
	return "games"
}
