package friend

import (
	"backend/internal/model"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type FriendRepository interface {
	GetFriend(ctx context.Context, charID, userID uint) (*model.Friend, error)
	GetByID(ctx context.Context, friendID uint) (*model.Friend, error)
	AddFriend(ctx context.Context, friend *model.Friend) error
	RemoveFriend(ctx context.Context, friendID uint) error
	GetList(ctx context.Context, userID uint, cursorUpdatedAt *time.Time, cursorID uint, limit int) ([]*model.Friend, error)
	GetSystemPrompts(ctx context.Context, title string) ([]*model.SystemPrompt, error)
	GetRecentMessages(ctx context.Context, friendID uint, limit int) ([]*model.Message, error)
	SaveMessageTx(ctx context.Context, msg *model.Message) error
	GetMessageCount(ctx context.Context, friendID uint) (int64, error)
	UpdateFriendMemory(ctx context.Context, friendID uint, newMemory string) error
	GetMessageHistory(ctx context.Context, friendID uint, cursor uint, limit int) ([]*model.Message, error)
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

func (r *friendRepository) AddFriend(ctx context.Context, f *model.Friend) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.Friend

		err := tx.Unscoped().
			Where("me_id = ? AND character_id = ?", f.MeID, f.CharacterID).
			First(&existing).Error

		if err == nil {
			if existing.DeletedAt.Valid {
				if err := tx.Unscoped().Model(&model.Friend{}).Where("id = ?", existing.ID).Update("deleted_at", nil).Error; err != nil {
					return err
				}

				return tx.Model(&model.Character{}).
					Where("id = ?", existing.CharacterID).
					Update("friend_count", gorm.Expr("friend_count + ?", 1)).Error
			}
			return nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := tx.Create(f).Error; err != nil {
			return err
		}

		return tx.Model(&model.Character{}).
			Where("id = ?", f.CharacterID).
			Update("friend_count", gorm.Expr("friend_count + ?", 1)).Error
	})
}

func (r *friendRepository) RemoveFriend(ctx context.Context, fid uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var f model.Friend
		if err := tx.Select("character_id").First(&f, fid).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.Friend{}, fid).Error; err != nil {
			return err
		}

		return tx.Model(&model.Character{}).
			Where("id = ?", f.CharacterID).
			Update("friend_count", gorm.Expr("friend_count - ?", 1)).Error
	})
}

func (r *friendRepository) GetList(ctx context.Context, userID uint, cursorUpdatedAt *time.Time, cursorID uint, limit int) ([]*model.Friend, error) {
	var friends []*model.Friend
	query := r.db.WithContext(ctx).
		Preload("Character.Author").
		Where("me_id = ?", userID)

	if cursorUpdatedAt != nil {
		query = query.Where(
			"(updated_at < ?) OR (updated_at = ? AND id < ?)",
			*cursorUpdatedAt,
			*cursorUpdatedAt,
			cursorID,
		)
	}

	err := query.
		Order("updated_at DESC, id DESC").
		Limit(limit).
		Find(&friends).Error

	return friends, err
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

func (r *friendRepository) SaveMessageTx(ctx context.Context, msg *model.Message) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(msg).Error; err != nil {
			return err
		}

		preview := []rune(msg.Output)
		previewStr := string(preview)
		if len(preview) > 50 {
			previewStr = string(preview[:50]) + "..."
		}
		if err := tx.Model(&model.Friend{}).Where("id = ?", msg.FriendID).Updates(map[string]any{
			"chat_count":   gorm.Expr("chat_count + 1"),
			"last_message": previewStr,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Character{}).
			Where("id = (SELECT character_id FROM friends WHERE id = ?)", msg.FriendID).
			Updates(map[string]any{
				"total_chat_count":  gorm.Expr("total_chat_count + 1"),
				"recent_chat_count": gorm.Expr("recent_chat_count + 1"),
			}).Error; err != nil {
			return err
		}

		return nil
	})
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

func (r *friendRepository) GetMessageHistory(ctx context.Context, friendID uint, cursor uint, limit int) ([]*model.Message, error) {
	var msgs []*model.Message
	q := r.db.WithContext(ctx).Where("friend_id = ?", friendID)
	if cursor > 0 {
		q = q.Where("id < ?", cursor)
	}
	if err := q.Order("id DESC").Limit(limit).Find(&msgs).Error; err != nil {
		return nil, err
	}
	return msgs, nil
}
