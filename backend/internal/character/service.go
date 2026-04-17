package character

import (
	"backend/internal/model"
	"backend/pkg/constants"
	"backend/pkg/utils"
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"mime/multipart"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CharService interface {
	CreateChar(ctx context.Context, authorID uint, name, profile string, photo, bg *multipart.FileHeader) error
	UpdateChar(ctx context.Context, authorID, charID uint, name, profile string, photo, bg *multipart.FileHeader) error
	GetCharSingle(ctx context.Context, charID uint) (*GetSingleResp, error)
	GetUserChars(ctx context.Context, authorID uint, itemsCount int) ([]*model.Character, error)
	DeleteChar(ctx context.Context, authorID, charID uint) error
	GetFeedOrSearch(ctx context.Context, query string, cursor string, limit int) ([]*model.Character, string, error)
}

type charService struct {
	repo CharRepository
}

func NewCharService(repo CharRepository) CharService {
	return &charService{repo: repo}
}

func (s *charService) CreateChar(ctx context.Context, authorID uint, name, profile string, photo, bg *multipart.FileHeader) error {
	name = strings.TrimSpace(name)
	profile = strings.TrimSpace(profile)

	nLen := utf8.RuneCountInString(name)
	if nLen < constants.MinCharNameLen || nLen > constants.MaxCharNameLen {
		return fmt.Errorf("名字长度需在 %d-%d 个字符之间", constants.MinCharNameLen, constants.MaxCharNameLen)
	}
	pLen := utf8.RuneCountInString(profile)
	if pLen == 0 {
		return errors.New("介绍不能为空")
	}
	if pLen > constants.MaxCharProfileLen {
		return fmt.Errorf("介绍太长了，最多支持 %d 个字符", constants.MaxCharProfileLen)
	}

	if photo == nil {
		return errors.New("头像不能为空")
	}
	if bg == nil {
		return errors.New("背景图片不能为空")
	}

	photoURL, err := utils.UploadFile(authorID, photo, constants.DirCharacterPhoto)
	if err != nil {
		zap.L().Error("[char service] Upload photo error", zap.Uint("authorID", authorID), zap.Error(err))
		return errors.New("头像上传失败")
	}

	bgURL, err := utils.UploadFile(authorID, bg, constants.DirCharacterBackgroundImage)
	if err != nil {
		_ = utils.RemoveFile(photoURL)
		zap.L().Error("[char service] Upload background error", zap.Uint("authorID", authorID), zap.Error(err))
		return errors.New("背景图片上传失败")
	}

	char := &model.Character{
		AuthorID:        authorID,
		Name:            name,
		Profile:         profile,
		Photo:           photoURL,
		BackgroundImage: bgURL,
	}

	if err := s.repo.Create(ctx, char); err != nil {
		_ = utils.RemoveFile(photoURL)
		_ = utils.RemoveFile(bgURL)
		zap.L().Error("[char service] CreateChar DB error", zap.Uint("authorID", authorID), zap.Error(err))
		return errors.New("系统繁忙，请稍后再试")
	}
	return nil
}

func (s *charService) UpdateChar(ctx context.Context, authorID, charID uint, name, profile string, photo, bg *multipart.FileHeader) error {
	// 1. 先查询旧数据进行鉴权
	oldChar, err := s.repo.GetByID(ctx, charID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		zap.L().Error("[char service] UpdateChar find char error", zap.Uint("charID", charID), zap.Error(err))
		return errors.New("系统繁忙，请稍后再试")
	}

	// 2. 权限校验：只有作者本人可以修改
	if oldChar.AuthorID != authorID {
		zap.L().Warn("[char service] UpdateChar permission denied", zap.Uint("userID", authorID), zap.Uint("charID", charID))
		return errors.New("角色不存在")
	}

	// 3. 准备更新字段
	name = strings.TrimSpace(name)
	profile = strings.TrimSpace(profile)

	nLen := utf8.RuneCountInString(name)
	if nLen < constants.MinCharNameLen || nLen > constants.MaxCharNameLen {

		return fmt.Errorf("名字长度需在 %d-%d 个字符之间", constants.MinCharNameLen, constants.MaxCharNameLen)
	}
	oldChar.Name = name

	pLen := utf8.RuneCountInString(profile)
	if pLen == 0 {
		return errors.New("介绍不能为空")
	}
	if pLen > constants.MaxCharProfileLen {
		return fmt.Errorf("介绍太长了，最多支持 %d 个字符", constants.MaxCharProfileLen)
	}
	oldChar.Profile = profile

	// 4. 处理图片：如果有新上传则更新，否则保持原样

	oldPhotoURL, oldBgURL := oldChar.Photo, oldChar.BackgroundImage
	var newPhotoURL, newBgURL string

	success := false
	cleanup := func() {
		if !success {
			if newPhotoURL != "" {
				_ = utils.RemoveFile(newPhotoURL)
			}
			if newBgURL != "" {
				_ = utils.RemoveFile(newBgURL)
			}
		}
	}
	defer cleanup()

	if photo != nil {
		url, err := utils.UploadFile(authorID, photo, constants.DirCharacterPhoto)
		if err != nil {
			zap.L().Error("[char service] Update photo error", zap.Error(err))
			return errors.New("新头像上传失败")
		}
		newPhotoURL = url
		oldChar.Photo = newPhotoURL
	}

	if bg != nil {
		url, err := utils.UploadFile(authorID, bg, constants.DirCharacterBackgroundImage)
		if err != nil {
			zap.L().Error("[char service] Update background error", zap.Error(err))
			return errors.New("新背景图片上传失败")
		}
		newBgURL = url
		oldChar.BackgroundImage = newBgURL
	}

	// 5. 执行更新
	if err := s.repo.Update(ctx, oldChar); err != nil {
		zap.L().Error("[char service] UpdateChar DB error", zap.Error(err))
		return errors.New("更新失败，请稍后再试")
	}
	success = true

	if oldPhotoURL != "" && newPhotoURL != "" {
		_ = utils.RemoveFile(oldPhotoURL)
	}
	if oldBgURL != "" && newBgURL != "" {
		_ = utils.RemoveFile(oldBgURL)
	}

	return nil
}

func (s *charService) GetCharSingle(ctx context.Context, charID uint) (*GetSingleResp, error) {
	char, err := s.repo.GetByID(ctx, charID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		zap.L().Error("[char service] GetByID db error", zap.Uint("charID", charID), zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}
	return &GetSingleResp{
		ID:              charID,
		Name:            char.Name,
		Profile:         char.Profile,
		Photo:           constants.StaticBaseURL + char.Photo,
		BackgroundImage: constants.StaticBaseURL + char.BackgroundImage,
	}, nil
}

func (s *charService) GetUserChars(ctx context.Context, authorID uint, itemsCount int) ([]*model.Character, error) {
	chars, err := s.repo.GetList(ctx, authorID, itemsCount, constants.DefaultLimit)
	if err != nil {
		zap.L().Error("[char service] GetList db error", zap.Uint("userID", authorID), zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}
	return chars, nil
}

func (s *charService) DeleteChar(ctx context.Context, authorID, charID uint) error {
	char, err := s.repo.GetByID(ctx, charID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		zap.L().Error("[char service] GetByID db error", zap.Uint("userID", authorID), zap.Uint("charID", charID), zap.Error(err))
		return errors.New("系统繁忙，请稍后再试")
	}

	if authorID != char.AuthorID {
		zap.L().Error("[char service] DeleteChar permission denied", zap.Uint("userID", authorID), zap.Uint("charID", charID))
		return errors.New("角色不存在")
	}

	if err := s.repo.Delete(ctx, charID); err != nil {
		zap.L().Error("[char service] Delete db error", zap.Uint("userID", authorID), zap.Uint("charID", charID), zap.Error(err))
		return errors.New("系统繁忙，请稍后再试")
	}

	return nil
}

func (s *charService) GetFeedOrSearch(ctx context.Context, query string, cursor string, limit int) ([]*model.Character, string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	cmap := make(map[uint]*model.Character)
	textScoreMap := make(map[uint]float64)

	recallLimit := constants.DefaultRecallLimit

	if query != "" {
		candidates, err := s.repo.SearchRecall(ctx, query, recallLimit)
		if err == nil {
			for _, c := range candidates {
				cmap[c.ID] = &c.Character
				textScoreMap[c.ID] = c.TextScore
			}
		}
	} else {
		add := func(cs []*model.Character) {
			mu.Lock()
			defer mu.Unlock()
			for _, c := range cs {
				if _, ok := cmap[c.ID]; !ok {
					cmap[c.ID] = c
					textScoreMap[c.ID] = 0.0
				}
			}
		}

		wg.Add(4)
		go func() { defer wg.Done(); cs, _ := s.repo.RecallTotal(ctx, recallLimit); add(cs) }()
		go func() { defer wg.Done(); cs, _ := s.repo.RecallRecent(ctx, recallLimit); add(cs) }()
		go func() { defer wg.Done(); cs, _ := s.repo.RecallNew(ctx, recallLimit); add(cs) }()
		go func() { defer wg.Done(); cs, _ := s.repo.RecallSocial(ctx, recallLimit); add(cs) }()
		wg.Wait()
	}

	// 2. 算分融合层 (Rerank)
	const tau = 172800.0
	now := time.Now()

	TimeDecay := func(ut time.Time) float64 {
		dt := now.Sub(ut).Seconds()
		if dt < 0 {
			dt = 0
		}
		return math.Exp(-dt / tau)
	}

	CalScore := func(c *model.Character, tScore float64) float64 {
		st := 0.3 * math.Log1p(float64(c.TotalChatCount))
		sr := 0.4 * math.Log1p(float64(c.RecentChatCount))
		sd := 0.2 * TimeDecay(c.UpdatedAt)
		sf := 0.1 * math.Log1p(float64(c.FriendCount))
		baseScore := st + sr + sd + sf + (rand.Float64() * 0.01)

		weightText := 5.0
		weightBase := 1.0
		if query == "" {
			weightText = 0.0
		}
		return (tScore * weightText) + (baseScore * weightBase)
	}

	res := make([]*model.RankedCharacter, 0, len(cmap))
	for _, c := range cmap {
		res = append(res, &model.RankedCharacter{
			Character: c,
			Score:     CalScore(c, textScoreMap[c.ID]),
		})
	}

	// 3. 排序 (Sort) - 分数降序，ID降序防碰撞
	sort.Slice(res, func(i, j int) bool {
		if res[i].Score == res[j].Score {
			return res[i].Character.ID > res[j].Character.ID
		}
		return res[i].Score > res[j].Score
	})

	// 4. 游标解析与截断 (Cursor Pagination)
	var lastScore float64 = math.MaxFloat64
	var lastID uint = 0
	if cursor != "" && cursor != "0" {
		parts := strings.Split(cursor, "_")
		if len(parts) == 2 {
			if s, err := strconv.ParseFloat(parts[0], 64); err == nil {
				lastScore = s
			}
			if id, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
				lastID = uint(id)
			}
		}
	}

	var finalRes []*model.RankedCharacter
	for _, item := range res {
		if item.Score < lastScore || (item.Score == lastScore && item.Character.ID < lastID) {
			finalRes = append(finalRes, item)
			if len(finalRes) == limit {
				break
			}
		}
	}

	// 5. 拼装结果与下一页游标
	nextCursor := ""
	out := make([]*model.Character, 0, len(finalRes))
	if len(finalRes) > 0 {
		lastItem := finalRes[len(finalRes)-1]
		nextCursor = fmt.Sprintf("%f_%d", lastItem.Score, lastItem.Character.ID)

		for _, r := range finalRes {
			out = append(out, r.Character)
		}
	}

	return out, nextCursor, nil
}
