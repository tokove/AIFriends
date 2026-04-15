package friend

type GetOrCreateReq struct {
	CharacterID uint `json:"character_id" binding:"required"`
}

type FriendResp struct {
	ID        uint          `json:"id"`
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
	FriendID uint   `json:"friend_id" binding:"required"`
	Message  string `json:"message" binding:"required"`
}

type MessageResp struct {
	ID          uint   `json:"id"`
	UserMessage string `json:"user_message"`
	Output      string `json:"output"`
}