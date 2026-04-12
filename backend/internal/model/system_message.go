package model

import (
	"time"

	"gorm.io/gorm"
)

type SystemPrompt struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"type:varchar(100);not null" json:"title"`
	OrderNumber int            `gorm:"default:0" json:"order_number"`
	Prompt      string         `gorm:"type:text;not null" json:"prompt"`
	CreatedAt   time.Time      `json:"create_time"`
	UpdatedAt   time.Time      `json:"update_time"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
