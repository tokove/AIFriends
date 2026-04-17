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
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
