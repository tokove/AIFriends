package friend

import (
	"backend/internal/model"
	"backend/pkg/constants"
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FriendService interface {
	GetOrCreate(ctx context.Context, charID, userID uint) (*model.Friend, error)
	GetList(ctx context.Context, userID uint, itemsCount int) ([]*model.Friend, error)
	DeleteFriend(ctx context.Context, friendID, userID uint) error
}

type friendService struct {
	repo FriendRepository
}

func NewFriendService(repo FriendRepository) FriendService {
	return &friendService{repo: repo}
}

func (s *friendService) GetOrCreate(ctx context.Context, charID, userID uint) (*model.Friend, error) {
	friend, err := s.repo.GetFriend(ctx, charID, userID)
	if err == nil {
		return friend, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		zap.L().Error("[friend service] GetFriend unexpected error", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}

	oldFriend, err := s.repo.GetFriendWithDeleted(ctx, charID, userID)
	if err == nil {
		if err := s.repo.RestoreFriend(ctx, oldFriend.ID); err != nil {
			zap.L().Error("[friend service] RestoreFriend error", zap.Error(err), zap.Uint("id", oldFriend.ID))
			return nil, errors.New("系统繁忙，请稍后再试")
		}
	} else {
		newFriend := &model.Friend{
			CharacterID: charID,
			MeID:        userID,
		}
		if err := s.repo.CreateFriend(ctx, newFriend); err != nil {
			zap.L().Warn("[friend service] CreateFriend conflict, retrying get", zap.Error(err))
		}
	}

	f, err := s.repo.GetFriend(ctx, charID, userID)
	if err != nil {
		zap.L().Error("[friend service] Final GetFriend error", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}
	return f, nil
}

func (s *friendService) GetList(ctx context.Context, userID uint, itemsCount int) ([]*model.Friend, error) {
	friends, err := s.repo.GetList(ctx, userID, itemsCount, constants.DefaultLimit)
	if err != nil {
		zap.L().Error("[friend service] GetList db error", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}
	return friends, nil
}

func (s *friendService) DeleteFriend(ctx context.Context, friendID, userID uint) error {
	friend, err := s.repo.GetByID(ctx, friendID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("好友不存在")
		}
		zap.L().Error("[friend service] GetByID db error", zap.Error(err))
		return errors.New("系统繁忙，请稍后再试")
	}

	if friend.MeID != userID {
		return errors.New("好友不存在")
	}

	if err := s.repo.DeleteFriend(ctx, friendID); err != nil {
		zap.L().Error("[friend service] DeleteFriend db error", zap.Error(err))
		return errors.New("系统繁忙，请稍后再试")
	}
	return nil
}
