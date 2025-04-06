// pkg/model/user.go
package model

type User struct {
	ID        string `gorm:"primaryKey;size:36"`
	Email     string `gorm:"unique;size:255"`
	Password  string `gorm:"size:255"`
	CreatedAt int64  `gorm:"not null"` // time.Time
	LastSeen  int64  `gorm:"not null"`
	Score     int    `gorm:"default:0"`
}
