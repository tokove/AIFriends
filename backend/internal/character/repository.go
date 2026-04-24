package character

import (
	"backend/internal/model"
	"context"
	"time"

	"gorm.io/gorm"
)

type CharRepository interface {
	Create(ctx context.Context, char *model.Character) error
	Update(ctx context.Context, char *model.Character) error
	GetByID(ctx context.Context, id uint) (*model.Character, error)
	GetUserByID(ctx context.Context, id uint) (*model.User, error)
	GetList(ctx context.Context, authorID uint, cursorUpdatedAt *time.Time, cursorID uint, limit int) ([]*model.Character, error)
	Delete(ctx context.Context, id uint) error
	HomeOrSearch(ctx context.Context, query string, cursorTime int64, cursorID uint, limit int) ([]*model.Character, error)
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
	return r.db.WithContext(ctx).Updates(char).Error
}

func (r *charRepository) GetByID(ctx context.Context, id uint) (*model.Character, error) {
	var char model.Character
	if err := r.db.WithContext(ctx).First(&char, id).Error; err != nil {
		return nil, err
	}
	return &char, nil
}

func (r *charRepository) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *charRepository) GetList(ctx context.Context, authorID uint, cursorUpdatedAt *time.Time, cursorID uint, limit int) ([]*model.Character, error) {
	var chars []*model.Character
	query := r.db.WithContext(ctx).Preload("Author").
		Where("author_id = ?", authorID)

	if cursorUpdatedAt != nil {
		query = query.Where(
			"(updated_at < ?) OR (updated_at = ? AND id < ?)",
			*cursorUpdatedAt,
			*cursorUpdatedAt,
			cursorID,
		)
	}

	if err := query.
		Order("updated_at DESC, id DESC").
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

func (r *charRepository) HomeOrSearch(ctx context.Context, query string, cursorTime int64, cursorID uint, limit int) ([]*model.Character, error) {
	var chars []*model.Character
	q := r.db.WithContext(ctx).Preload("Author")

	if query != "" {
		like := "%" + query + "%"
		q = q.Where(
			r.db.Where("name ILIKE ?", like).Or("profile ILIKE ?", like),
		)
	}

	if cursorTime > 0 {
		q = q.Where(
			"updated_at < ? OR (updated_at = ? AND id < ?)",
			time.Unix(cursorTime, 0),
			time.Unix(cursorTime, 0),
			cursorID,
		)
	}

	if err := q.Order("updated_at DESC, id DESC").Limit(limit).Find(&chars).Error; err != nil {
		return nil, err
	}

	return chars, nil
}
