package utils

import (
	"backend/pkg/constants"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// CheckImage 深度安全检查：大小和真实内容类型
func CheckImage(file *multipart.FileHeader) error {
	if file.Size > constants.MaxFileSize {
		return fmt.Errorf("图片大小不能超过 %.fMB", float64(constants.MaxFileSize)/1024/1024)
	}

	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	// 读取前512字节探测真实MIME类型
	buffer := make([]byte, 512)
	_, _ = f.Read(buffer)
	contentType := http.DetectContentType(buffer)

	if !strings.HasPrefix(contentType, "image/") {
		return errors.New("文件格式非法，只允许上传图片")
	}
	return nil
}

func UploadFile(userID uint, file *multipart.FileHeader, subDir string) (string, error) {
	if file == nil {
		return "", nil
	}

	// 1. 生成文件名: {userID}_{10位随机串}{后缀}
	ext := filepath.Ext(file.Filename)
	randomStr := strings.ReplaceAll(uuid.New().String(), "-", "")[:10]
	var fileName, relPath string

	switch subDir {
	case constants.DirUserPhoto:
		fileName = fmt.Sprintf("%d_%s%s", userID, randomStr, ext)
		relPath = filepath.Join(subDir, fileName)
	default:
		fileName = fmt.Sprintf("%s%s", randomStr, ext)
		relPath = filepath.Join(subDir, fmt.Sprintf("%d", userID), fileName)
	}

	// 2. 构造存储路径
	// 数据库存这个: user/photos/1_abc1234567.jpg
	// 物理路径: media/user/photos/1_abc1234567.jpg
	fullPath := filepath.Join("media", relPath)

	// 3. 确保目录存在 (mkdir -p ./media/user/photos)
	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return "", err
	}

	// 4. 写入文件
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return relPath, nil
}

func RemoveFile(relPath string) error {
	if relPath == "" {
		return nil
	}

	// 逻辑 A: 包含 "default" 关键字的都不删 (最省心)
	if strings.Contains(strings.ToLower(relPath), "default") {
		return nil
	}

	fullPath := filepath.Join("media", relPath)

	// 2. 检查并删除
	if _, err := os.Stat(fullPath); err == nil {
		return os.Remove(fullPath)
	}
	return nil
}

func SplitText(text string, chunkSize int, overlap int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	if chunkSize <= 0 {
		return []string{text}
	}
	if overlap >= chunkSize {
		overlap = chunkSize - 1
	}

	runes := []rune(text)
	totalLen := len(runes)
	var chunks []string

	if totalLen <= chunkSize {
		return []string{text}
	}

	step := chunkSize - overlap

	for i := 0; i < totalLen; i += step {
		end := i + chunkSize
		end = min(end, totalLen)

		chunks = append(chunks, string(runes[i:end]))

		if end == totalLen {
			break
		}
	}

	return chunks
}
