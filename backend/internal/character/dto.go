package character

type GetSingleReq struct {
	CharID uint `json:"character_id"`
}

type DeleteCharReq struct {
	CharID uint `json:"character_id"`
}

type GetSingleResp struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Profile         string `json:"profile"`
	Photo           string `json:"photo"`
	BackgroundImage string `json:"background_image"`
}

type AuthorInfoResp struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Photo    string `json:"photo"`
}

type CharacterItemResp struct {
	ID              uint           `json:"id"`
	Name            string         `json:"name"`
	Profile         string         `json:"profile"`
	Photo           string         `json:"photo"`
	BackgroundImage string         `json:"background_image"`
	Author          AuthorInfoResp `json:"author"`
}

type UserProfileResp struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Profile  string `json:"profile"`
	Photo    string `json:"photo"`
}
