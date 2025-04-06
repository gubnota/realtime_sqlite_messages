package model

type Device struct {
	ID        string `gorm:"primaryKey;size:36"`
	UserID    string `gorm:"index;not null;size:36"`
	LastSeen  int64  `gorm:"not null"`                 // Unix timestamp
	Status    string `gorm:"type:CHAR(1);default:'F'"` // O - online, F - offline
	UserAgent string `gorm:"size:500"`
}

func (Device) TableName() string {
	return "devices"
}
