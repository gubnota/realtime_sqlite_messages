package model

import (
	"time"
)

type Message struct {
	ID        uint      `gorm:"primaryKey"`
	Sender    string    `gorm:"index;not null"`
	Receiver  string    `gorm:"index;not null"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null"`
	Delivered bool      `gorm:"default:false;not null"`
}

func (Message) TableName() string {
	return "messages"
}
