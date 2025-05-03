package model

type Recipe struct {
    ID      string `gorm:"primaryKey;size:36"`
    SpaceID string `gorm:"index;not null"`
    Name    string `gorm:"not null"`
}

type RecipeIngredient struct {
    ID          string `gorm:"primaryKey;size:36"`
    RecipeID    string `gorm:"index;not null"`
    IngredientID string `gorm:"index;not null"`
    Quantity    int64  `gorm:"not null"`
}