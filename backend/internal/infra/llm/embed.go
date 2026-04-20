package llm

import (
	"backend/internal/config"
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
)

// Embedder 定义了向量化接口, pgvector 数据库需要 float32 格式
type Embedder interface {
	EmbedString(ctx context.Context, text string) ([]float32, error)
	EmbedStrings(ctx context.Context, texts []string) ([][]float32, error)
}

// DefaultEmbedder 具体的实现类
type DefaultEmbedder struct {
	instance embedding.Embedder
}

// 生成embedder
func NewDefaultEmbedder(cfg config.AgentConfig) (*DefaultEmbedder, error) {
	embedder, err := openai.NewEmbedder(context.Background(), &openai.EmbeddingConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.EmbedModel,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化 Embedder 失败: %w", err)
	}

	return &DefaultEmbedder{instance: embedder}, nil
}

// EmbedString 获取单条文本的向量
func (e *DefaultEmbedder) EmbedString(ctx context.Context, text string) ([]float32, error) {
	// Eino 原生返回的是 [][]float64
	res, err := e.instance.EmbedStrings(ctx, []string{text})
	if err != nil || len(res) == 0 {
		return nil, err
	}

	// 转换 float64 为 float32 以适配 pgvector 数据库
	return convertToFloat32(res[0]), nil
}

// EmbedStrings 获取多条文本的向量
func (e *DefaultEmbedder) EmbedStrings(ctx context.Context, texts []string) ([][]float32, error) {
	res, err := e.instance.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, err
	}

	var result [][]float32
	for _, vec := range res {
		result = append(result, convertToFloat32(vec))
	}
	return result, nil
}

func convertToFloat32(input []float64) []float32 {
	output := make([]float32, len(input))
	for i, v := range input {
		output[i] = float32(v)
	}
	return output
}
