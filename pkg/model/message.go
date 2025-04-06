package model

type Message struct {
	ID        uint   `gorm:"primaryKey"`
	Sender    string `gorm:"index;not null;index:idx_sender"`
	Receiver  string `gorm:"index;not null;index:idx_receiver"`
	Content   string `gorm:"type:text;not null"`
	CreatedAt int64  `gorm:"not null"` //time.Time `gorm:"not null"`
	Delivered bool   `gorm:"default:false;not null"`
}

func (Message) TableName() string {
	return "messages"
}
