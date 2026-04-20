package tool

import (
	"backend/internal/infra/db"
	"context"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"go.uber.org/zap"
)

func InitTools(vector *db.VectorDB) []tool.BaseTool {
	// InferTool 参数: 工具名称, 工具描述, 实际执行的函数
	// 1. 查时间工具
	timeTool, err := utils.InferTool(
		"get_current_time",
		"当用户询问现在几点了、今天是几号时调用此工具",
		func(ctx context.Context, req *GetTimeReq) (*GetTimeResp, error) {
			zap.L().Info("[tools] agent call timeTool")
			now := time.Now().Format("2006年01月02日 15:04:05")
			prefix := "当前系统时间是："
			if req.Location != "" {
				prefix = req.Location + "的时间是："
			}
			return &GetTimeResp{CurrentTime: prefix + now}, nil
		},
	)
	if err != nil {
		zap.L().Error("[agent tools] create get_current_time tool failed", zap.Error(err))
	}

	// 创建 RAG 工具
	ragTool, err := utils.InferTool(
		"search_knowledge",
		"当用户询问关于 AIFriends 设定、角色背景、幻梦之森相关知识时，调用此工具获取权威资料。",
		func(ctx context.Context, req *SearchKnowledgeReq) (*SearchKnowledgeResp, error) {
			// 每次召回相关性最高的前 3 个片段
			zap.L().Info("[tools] agent call ragTool")
			result, err := vector.SearchRelevantChunks(ctx, req.Query, 3)
			if err != nil {
				return nil, err
			}
			return &SearchKnowledgeResp{Context: result}, nil
		},
	)
	if err != nil {
		zap.L().Error("[agent tools] create rag tool failed", zap.Error(err))
	}

	return []tool.BaseTool{timeTool, ragTool}
}
