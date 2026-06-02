package model

import "time"

type Conversion struct {
	ID        string `gorm:"primaryKey"`
	Platform  string
	Format    string
	Filename  string
	FilePath  string
	CreatedAt time.Time
	ExpiresAt time.Time
}
