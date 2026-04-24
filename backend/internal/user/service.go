package user

import (
	"backend/internal/model"
	"backend/pkg/constants"
	"backend/pkg/utils"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"unicode/utf8"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 接口定义（业务层契约）
type UserService interface {
	Register(ctx context.Context, username, password string) (*model.User, error)
	Login(ctx context.Context, username, password string) (*model.User, error)
	GetUserInfo(ctx context.Context, userID uint) (*model.User, error)
	UpdateProfile(ctx context.Context, userID uint, username, profile string, photo *multipart.FileHeader) (*model.User, error)
}

type userService struct {
	repo UserRepository
}

// NewUserService 构造函数
func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func validateUsername(username string) error {
	length := utf8.RuneCountInString(username)
	if length < constants.MinUsernameLen || length > constants.MaxUsernameLen {
		return fmt.Errorf("用户名长度需在 %d-%d 个字符之间", constants.MinUsernameLen, constants.MaxUsernameLen)
	}
	return nil
}

func validatePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < constants.MinPasswordLen || length > constants.MaxPasswordLen {
		return fmt.Errorf("密码长度需在 %d-%d 个字符之间", constants.MinPasswordLen, constants.MaxPasswordLen)
	}
	return nil
}

func (s *userService) Register(ctx context.Context, username, password string) (*model.User, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	existingUser, err := s.repo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		zap.L().Error("[user service] GetByUsername error: ", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		zap.L().Error("[user service] bcrypt error: ", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}

	user := &model.User{
		Username: username,
		Password: string(hashedPass),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		zap.L().Error("[user service] Create error: ", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, username, password string) (*model.User, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.L().Error("[user service] GetByUsername error: ", zap.Error(err))
			return nil, errors.New("用户名或密码错误")
		}
		zap.L().Error("[user service] GetByUsername error: ", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	return user, nil
}

func (s *userService) GetUserInfo(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		zap.L().Error("[user service] GetUserInfo error: ", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}

	if user == nil {
		return nil, errors.New("用户不存在")
	}

	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uint, username, profile string, photo *multipart.FileHeader) (*model.User, error) {
	username = strings.TrimSpace(username)

	// 1. 用户名长度校验
	uLen := utf8.RuneCountInString(username)
	if uLen < constants.MinUsernameLen || uLen > constants.MaxUsernameLen {
		return nil, fmt.Errorf("用户名长度需在 %d-%d 个字符之间", constants.MinUsernameLen, constants.MaxUsernameLen)
	}

	// 2. 简介长度校验
	pLen := utf8.RuneCountInString(profile)
	if pLen > constants.MaxUserProfileLen {
		return nil, fmt.Errorf("简介太长了，最多支持 %d 个字符", constants.MaxUserProfileLen)
	}

	// 3. 获取用户信息
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		zap.L().Error("[user service] GetByID error: ", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}

	// 4. 用户名改动查重
	if username != user.Username {
		userExisting, err := s.repo.GetByUsername(ctx, username)
		if err == nil && userExisting != nil {
			return nil, errors.New("用户名已存在")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			zap.L().Error("[user service] GetByUsername error: ", zap.Error(err))
			return nil, errors.New("系统繁忙，请稍后再试")
		}
	}

	// 5. 更新基本信息
	user.Username = username
	user.Profile = profile

	oldPhotoURL := user.Photo
	var newPhotoURL string
	// 6. 图片处理
	if photo != nil {
		url, err := utils.UploadFile(userID, photo, constants.DirUserPhoto)
		if err != nil {
			return nil, err
		}
		newPhotoURL = url
		user.Photo = newPhotoURL
	}

	// 7. 写入数据库
	if err := s.repo.Update(ctx, user); err != nil {
		if newPhotoURL != "" {
			_ = utils.RemoveFile(newPhotoURL) // 数据库没存上，刚才上传的新图就没用了，删掉它
		}
		zap.L().Error("[user service] Update error: ", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}

	if oldPhotoURL != "" && newPhotoURL != "" {
		_ = utils.RemoveFile(oldPhotoURL)
	}
	return user, nil
}
