package model

import (
	"time"

	"gorm.io/gorm"
)

type Friend struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	MeID        uint           `gorm:"uniqueIndex:idx_me_char;not null" json:"me_id"`
	CharacterID uint           `gorm:"uniqueIndex:idx_me_char;not null" json:"character_id"`
	Memory      string         `gorm:"type:text" json:"memory"`
	ChatCount   int64          `gorm:"default:0" json:"chat_count"`
	LastMessage string         `gorm:"type:text" json:"last_message"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Character *Character `gorm:"foreignKey:CharacterID;constraint:OnDelete:CASCADE;" json:"character,omitempty"`
}
