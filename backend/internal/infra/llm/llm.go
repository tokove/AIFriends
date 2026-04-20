package llm

import (
	"context"

	"backend/internal/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// InitChatModel 使用阿里云百炼的 OpenAI 兼容模式连接千问
func InitChatModel(ctx context.Context, cfg config.AgentConfig) (model.ToolCallingChatModel, error) {
	// 划重点：这里是 ChatModelConfig，不是 Config！
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   cfg.LLMModel, // 对应你的 MODEL
		APIKey:  cfg.APIKey,   // 对应你的 API_KEY
		BaseURL: cfg.BaseURL,  // 对应你的 BASE_URL
	})
}
