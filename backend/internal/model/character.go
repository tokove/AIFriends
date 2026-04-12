package model

import (
	"time"

	"gorm.io/gorm"
)

type Character struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	AuthorID        uint           `gorm:"index;not null" json:"author_id"`
	LikesCount      int            `gorm:"default:0;index" json:"likes_count"`
	Name            string         `gorm:"type:varchar(50);not null;index" json:"name"`
	Photo           string         `gorm:"type:varchar(255);not null" json:"photo"`
	Profile         string         `gorm:"type:text;not null" json:"profile"`
	BackgroundImage string         `gorm:"type:varchar(255);not null" json:"background_image"`
	CreatedAt       time.Time      `json:"create_at"`
	UpdatedAt       time.Time      `json:"update_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	
	Author          *User          `gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE;" json:"author,omitempty"`
}
