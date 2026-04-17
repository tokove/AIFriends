package character

import (
	"backend/internal/model"
	"context"

	"gorm.io/gorm"
)

type CharRepository interface {
	Create(ctx context.Context, char *model.Character) error
	Update(ctx context.Context, char *model.Character) error
	GetByID(ctx context.Context, id uint) (*model.Character, error)
	GetList(ctx context.Context, authorID uint, offset int, limit int) ([]*model.Character, error)
	Delete(ctx context.Context, id uint) error
	SearchRecall(ctx context.Context, query string, limit int) ([]*model.SearchCandidate, error)
	RecallTotal(ctx context.Context, limit int) ([]*model.Character, error)
	RecallRecent(ctx context.Context, limit int) ([]*model.Character, error)
	RecallNew(ctx context.Context, limit int) ([]*model.Character, error)
	RecallSocial(ctx context.Context, limit int) ([]*model.Character, error)
}

type charRepository struct {
	db *gorm.DB
}

func NewCharRepository(db *gorm.DB) CharRepository {
	return &charRepository{db: db}
}

func (r *charRepository) Create(ctx context.Context, char *model.Character) error {
	return r.db.WithContext(ctx).Create(char).Error
}

func (r *charRepository) Update(ctx context.Context, char *model.Character) error {
	return r.db.WithContext(ctx).Save(char).Error
}

func (r *charRepository) GetByID(ctx context.Context, id uint) (*model.Character, error) {
	var char model.Character
	if err := r.db.WithContext(ctx).First(&char, id).Error; err != nil {
		return nil, err
	}
	return &char, nil
}

func (r *charRepository) GetList(ctx context.Context, authorID uint, offset int, limit int) ([]*model.Character, error) {
	var chars []*model.Character
	if err := r.db.WithContext(ctx).Preload("Author").
		Where("author_id = ?", authorID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&chars).Error; err != nil {
		return nil, err
	}
	return chars, nil
}

func (r *charRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.Character{}, id).Error; err != nil {
			return err
		}
		if err := tx.Where("character_id = ?", id).Delete(&model.Friend{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *charRepository) SearchRecall(ctx context.Context, query string, limit int) ([]*model.SearchCandidate, error) {
	var candidates []*model.SearchCandidate
	err := r.db.WithContext(ctx).
		Table("characters").
		Preload("Author").
		Select("*, similarity(name, ?) as text_score", query).
		Where("name ILIKE ? OR profile ILIKE ?", "%"+query+"%", "%"+query+"%").
		Order("text_score DESC").
		Limit(limit).
		Find(&candidates).Error
	return candidates, err
}

func (r *charRepository) RecallTotal(ctx context.Context, limit int) ([]*model.Character, error) {
	var chars []*model.Character
	err := r.db.WithContext(ctx).Preload("Author").Order("total_chat_count DESC").Limit(limit).Find(&chars).Error
	return chars, err
}

func (r *charRepository) RecallRecent(ctx context.Context, limit int) ([]*model.Character, error) {
	var chars []*model.Character
	err := r.db.WithContext(ctx).Preload("Author").Order("recent_chat_count DESC").Limit(limit).Find(&chars).Error
	return chars, err
}

func (r *charRepository) RecallNew(ctx context.Context, limit int) ([]*model.Character, error) {
	var chars []*model.Character
	err := r.db.WithContext(ctx).Preload("Author").Order("updated_at DESC").Limit(limit).Find(&chars).Error
	return chars, err
}

func (r *charRepository) RecallSocial(ctx context.Context, limit int) ([]*model.Character, error) {
	var chars []*model.Character
	if err := r.db.WithContext(ctx).Preload("Author").Order("friend_count DESC").Limit(limit).Find(&chars).Error; err != nil {
		return nil, err
	}
	return chars, nil
}
