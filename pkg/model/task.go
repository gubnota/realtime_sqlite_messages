package model

import "time"

type Task struct {
    ID          string    `gorm:"primaryKey;size:36"`
    SpaceID     string    `gorm:"index;not null"`
    RecipeID    string    `gorm:"index;not null"`
    AssignedTo  string    `gorm:"index"`
    Department  string    `gorm:"index"`
    Comment     string
    Quantity    int       `gorm:"not null"`
    Status      string    `gorm:"default:'new'"` // new, in_progress, done, cancelled
    ETA         int       // in minutes
    ScheduledAt time.Time
    CreatedAt   time.Time
}