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
	if err := r.db.WithContext(ctx).First(&friend, friendID).Error; err != nil {
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
		Offset(int(offset)).
		Limit(limit).
		Find(&friends).Error

	return friends, err
}

func (r *friendRepository) DeleteFriend(ctx context.Context, friendID uint) error {
	return r.db.WithContext(ctx).Delete(&model.Friend{}, friendID).Error
}

func (r *friendRepository) GetFriendWithDeleted(ctx context.Context, charID, userID uint) (*model.Friend, error) {
	var friend model.Friend
	err := r.db.WithContext(ctx).Unscoped().
		Preload("Character.Author").
		Where("character_id = ? AND me_id = ?", charID, userID).
		First(&friend).Error
	if err != nil {
		return nil, err
	}
	return &friend, nil
}

func (r *friendRepository) RestoreFriend(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().
		Model(&model.Friend{}).Where("id = ?", id).
		Update("deleted_at", nil).Error
}
