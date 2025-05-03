package model

type Ingredient struct {
    ID          string `gorm:"primaryKey;size:36"`
    SpaceID     string `gorm:"index;not null"`
    Name        string `gorm:"not null;size:255"`
    Unit        string `gorm:"not null"` // "pcs", "ml", "g", "portion"
    TrackStock  bool   `gorm:"default:true"`
    Quantity    int64  `gorm:"not null"` // stored in base unit (e.g. grams)
}