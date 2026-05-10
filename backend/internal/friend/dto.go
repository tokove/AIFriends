package friend

type GetOrCreateReq struct {
	CharacterID uint `json:"character_id" binding:"required"`
}

type FriendResp struct {
	ID        uint          `json:"id"`
	UpdatedAt int64         `json:"updated_at"`
	Character CharacterResp `json:"character"`
}

type CharacterResp struct {
	ID      uint       `json:"id"`
	Name    string     `json:"name"`
	Profile string     `json:"profile"`
	Photo   string     `json:"photo"`
	BgImage string     `json:"background_image"`
	Author  AuthorResp `json:"author"`
}

type AuthorResp struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Photo    string `json:"photo"`
}

type RemoveReq struct {
	FriendID uint `json:"friend_id" binding:"required"`
}

type ChatReq struct {
	FriendID            uint   `json:"friend_id" binding:"required"`
	Message             string `json:"message" binding:"required"`
	UserMessageType     string `json:"user_message_type"`
	UserAudio           string `json:"user_audio"`
	UserASRText         string `json:"user_asr_text"`
	UserAudioDurationMS int    `json:"user_audio_duration_ms"`
	EnableTTS           bool   `json:"enable_tts"`
}

type streamEvent struct {
	Content      string
	AudioBase64  string
	Err          error
	Interrupted  bool
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

type MessageResp struct {
	ID                  uint   `json:"id"`
	UserMessage         string `json:"user_message"`
	UserMessageType     string `json:"user_message_type"`
	UserAudio           string `json:"user_audio"`
	UserASRText         string `json:"user_asr_text"`
	UserAudioDurationMS int    `json:"user_audio_duration_ms"`
	Output              string `json:"output"`
}

type TTSReq struct {
	FriendID  uint `json:"friend_id" binding:"required"`
	MessageID uint `json:"message_id" binding:"required"`
}