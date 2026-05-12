package model

import (
	"time"
)

type Voice struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	VoiceID   string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"voice_id"`
	CreatedAt time.Time `json:"created_at"`
}
