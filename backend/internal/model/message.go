package model

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	FriendID            uint           `gorm:"index;not null" json:"friend_id"`
	UserMessage         string         `gorm:"type:text;not null" json:"user_message"`
	UserMessageType     string         `gorm:"type:varchar(16);not null;default:'text'" json:"user_message_type"`
	UserAudio           string         `gorm:"type:text;not null;default:''" json:"user_audio"`
	UserASRText         string         `gorm:"type:text;not null;default:''" json:"user_asr_text"`
	UserAudioDurationMS int            `gorm:"default:0" json:"user_audio_duration_ms"`
	Input               string         `gorm:"type:text;not null" json:"input"`
	Output              string         `gorm:"type:text;not null" json:"output"`
	InputTokens         int            `gorm:"default:0" json:"input_tokens"`
	OutputTokens        int            `gorm:"default:0" json:"output_tokens"`
	TotalTokens         int            `gorm:"default:0" json:"total_tokens"`
	CreatedAt           time.Time      `json:"created_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	Friend *Friend `gorm:"foreignKey:FriendID" json:"-"`
}
