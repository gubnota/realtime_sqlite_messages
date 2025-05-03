package model

type Department struct {
    ID      string `gorm:"primaryKey;size:36"`
    SpaceID string `gorm:"index;not null"`
    Name    string `gorm:"not null;size:255"`
}