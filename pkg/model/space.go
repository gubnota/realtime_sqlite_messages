package model

type Space struct {
    ID        string   `gorm:"primaryKey;size:36"`
    Name      string   `gorm:"not null;size:255"`
    OwnerID   string   `gorm:"index;not null"`
    CreatedAt int64    `gorm:"not null"`
}