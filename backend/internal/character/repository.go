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
	GetList(ctx context.Context, authorID uint) ([]*model.Character, error)
	Delete(ctx context.Context, id uint) error
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

func (r *charRepository) GetList(ctx context.Context, authorID uint) ([]*model.Character, error) {
	var chars []*model.Character
	if err := r.db.WithContext(ctx).Preload("Author").Where("author_id = ?", authorID).Order("created_at DESC").Find(&chars).Error; err != nil {
		return nil, err
	}
	return chars, nil
}

func (r *charRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Character{}, id).Error
}
