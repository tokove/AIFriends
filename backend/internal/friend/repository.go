package friend

import (
	"backend/internal/model"
	"context"

	"gorm.io/gorm"
)

type FriendRepository interface {
	GetFriend(ctx context.Context, charID, userID uint) (*model.Friend, error)
	GetByID(ctx context.Context, friendID uint) (*model.Friend, error)
	CreateFriend(ctx context.Context, friend *model.Friend) error
	GetList(ctx context.Context, userID uint, offset int, limit int) ([]*model.Friend, error)
	DeleteFriend(ctx context.Context, friendID uint) error
	GetFriendWithDeleted(ctx context.Context, charID, userID uint) (*model.Friend, error)
	RestoreFriend(ctx context.Context, id uint) error
	GetSystemPrompts(ctx context.Context, title string) ([]*model.SystemPrompt, error)
	GetRecentMessages(ctx context.Context, friendID uint, limit int) ([]*model.Message, error)
	CreateMessage(ctx context.Context, msg *model.Message) error
	GetMessageCount(ctx context.Context, friendID uint) (int64, error)
	UpdateFriendMemory(ctx context.Context, friendID uint, newMemory string) error
	UpdateFriendActiveStatus(ctx context.Context, friendID uint, lastMsg string) error
	GetMessageHistory(ctx context.Context, friendID uint, lastMsgID uint, limit int) ([]*model.Message, error)
}

type friendRepository struct {
	db *gorm.DB
}

func NewFriendRepository(db *gorm.DB) FriendRepository {
	return &friendRepository{db: db}
}

func (r *friendRepository) GetFriend(ctx context.Context, charID, userID uint) (*model.Friend, error) {
	var friend model.Friend
	err := r.db.WithContext(ctx).
		Preload("Character.Author").
		Where("character_id = ? AND me_id = ?", charID, userID).
		First(&friend).Error
	if err != nil {
		return nil, err
	}
	return &friend, nil
}

func (r *friendRepository) GetByID(ctx context.Context, friendID uint) (*model.Friend, error) {
	var friend model.Friend
	if err := r.db.WithContext(ctx).
		Preload("Character.Author").
		First(&friend, friendID).Error; err != nil {
		return nil, err
	}
	return &friend, nil
}

func (r *friendRepository) CreateFriend(ctx context.Context, friend *model.Friend) error {
	return r.db.WithContext(ctx).Create(friend).Error
}

func (r *friendRepository) GetList(ctx context.Context, userID uint, offset int, limit int) ([]*model.Friend, error) {
	var friends []*model.Friend
	err := r.db.WithContext(ctx).
		Preload("Character.Author").
		Where("me_id = ?", userID).
		Order("updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&friends).Error

	return friends, err
}

func (r *friendRepository) DeleteFriend(ctx context.Context, friendID uint) error {
	return r.db.WithContext(ctx).Delete(&model.Friend{}, friendID).Error
}

func (r *friendRepository) GetFriendWithDeleted(ctx context.Context, charID, userID uint) (*model.Friend, error) {
	var friend model.Friend
	if err := r.db.WithContext(ctx).Unscoped().
		Preload("Character.Author").
		Where("character_id = ? AND me_id = ?", charID, userID).
		First(&friend).Error; err != nil {
		return nil, err
	}
	return &friend, nil
}

func (r *friendRepository) RestoreFriend(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().
		Model(&model.Friend{}).Where("id = ?", id).
		Update("deleted_at", nil).Error
}

func (r *friendRepository) GetSystemPrompts(ctx context.Context, title string) ([]*model.SystemPrompt, error) {
	var prompts []*model.SystemPrompt
	if err := r.db.WithContext(ctx).
		Where("title = ?", title).
		Order("order_number ASC").
		Find(&prompts).Error; err != nil {
		return nil, err
	}
	return prompts, nil
}

func (r *friendRepository) GetRecentMessages(ctx context.Context, friendID uint, limit int) ([]*model.Message, error) {
	var msgs []*model.Message
	if err := r.db.WithContext(ctx).
		Where("friend_id = ?", friendID).
		Order("id DESC").
		Limit(limit).
		Find(&msgs).Error; err != nil {
		return nil, err
	}
	// reverse
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (r *friendRepository) CreateMessage(ctx context.Context, msg *model.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *friendRepository) GetMessageCount(ctx context.Context, friendID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Friend{}).
		Where("id = ?", friendID).
		Pluck("chat_count", &count).Error
	return count, err
}

func (r *friendRepository) UpdateFriendMemory(ctx context.Context, friendID uint, newMemory string) error {
	return r.db.WithContext(ctx).Model(&model.Friend{}).Where("id = ?", friendID).Update("memory", newMemory).Error
}

func (r *friendRepository) UpdateFriendActiveStatus(ctx context.Context, friendID uint, lastMsg string) error {
	return r.db.WithContext(ctx).
		Model(&model.Friend{}).
		Where("id = ?", friendID).
		Updates(map[string]any{
			"chat_count":   gorm.Expr("chat_count + 1"),
			"last_message": lastMsg,
		}).Error
}

func (r *friendRepository) GetMessageHistory(ctx context.Context, friendID uint, lastMsgID uint, limit int) ([]*model.Message, error) {
	var msgs []*model.Message

	query := r.db.WithContext(ctx).Where("friend_id = ?", friendID)
	if lastMsgID > 0 {
		query = query.Where("id < ?", lastMsgID)
	}
	err := query.Order("id DESC").Limit(limit).Find(&msgs).Error
	return msgs, err
}
