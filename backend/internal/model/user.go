package model


import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"type:varchar(32);not null;index:idx_username_active,unique,where:deleted_at IS NULL" json:"username"`
	Password  string         `gorm:"type:varchar(128);not null" json:"-"`
	Photo     string         `gorm:"type:varchar(255);not null;default:'user/photos/default.jpg'" json:"photo"`
	Profile   string         `gorm:"type:text;default:'这个用户很懒，什么也没留下。'" json:"profile"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
