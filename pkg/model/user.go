// pkg/model/user.go
package model

import (
	"time"
)

type User struct {
	ID        string `gorm:"primaryKey;size:36"`
	Email     string `gorm:"unique;size:255"`
	Password  string `gorm:"size:255"`
	CreatedAt time.Time
}
