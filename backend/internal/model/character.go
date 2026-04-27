package model

import (
	"time"

	"gorm.io/gorm"
)

type Character struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	AuthorID        uint           `gorm:"index;not null" json:"author_id"`
	TotalChatCount  int64          `gorm:"default:0" json:"total_chat_count"`
	FriendCount     int64          `gorm:"default:0" json:"friend_count"`
	RecentChatCount int64          `gorm:"default:0" json:"recent_chat_count"`
	Name            string         `gorm:"type:varchar(50);not null;index" json:"name"`
	VoiceID         string         `gorm:"type:varchar(100);default:''" json:"voice_id"`
	Photo           string         `gorm:"type:varchar(255);not null" json:"photo"`
	Profile         string         `gorm:"type:text;not null" json:"profile"`
	BackgroundImage string         `gorm:"type:varchar(255);not null" json:"background_image"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	Author *User `gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE;" json:"author,omitempty"`
}
