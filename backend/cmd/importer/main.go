package main

import (
	"context"
	"fmt"
	"os"

	"backend/internal/config"
	"backend/internal/infra/db"
	"backend/internal/infra/llm"
	"backend/internal/infra/logger"
	"backend/internal/model"
	"backend/pkg/constants"
	"backend/pkg/utils"

	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	// init config
	cfg := config.LoadConfig(constants.DefaultPath)

	// init logger
	logger.InitLogger(cfg)

	// init db
	db.InitDB(cfg)

	// 确保数据库安装了 vector 插件
	if err := db.DB.Exec("CREATE EXTENSION IF NOT EXISTS vector;").Error; err != nil {
		zap.L().Fatal("install vector extension failed", zap.Error(err))
	}

	// _ = db.DB.Migrator().DropTable(&model.KnowledgeDoc{})
	// 自动迁移知识库表结构
	if err := db.DB.AutoMigrate(&model.KnowledgeDoc{}); err != nil {
		zap.L().Fatal("AutoMigrate failed", zap.Error(err))
	}

	embedder, err := llm.NewDefaultEmbedder(cfg.Agent)
	if err != nil {
		zap.L().Fatal("init embed failed", zap.Error(err))
	}

	// 初始化 VectorDB 操作封装
	vectorDB := db.NewVectorDB(db.DB, embedder)

	filePath := "./documents/knowledge/data.txt"
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		zap.L().Fatal("ReadFile failed", zap.Error(err))
	}
	content := string(contentBytes)

	// 调用 utils.SplitText (每块 500 字，前后重叠 50 字)
	chunks := utils.SplitText(content, 500, 50)
	fmt.Printf("文件读取成功，共切分为 %d 个文本块。开始请求大模型向量化...\n", len(chunks))

	// 分批向量化并入库
	batchSize := 10
	var allDocs []*model.KnowledgeDoc

	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		end = min(end, len(chunks))

		batchChunks := chunks[i:end]
		fmt.Printf("正在调用 API 处理第 %d 到 %d 个块...\n", i+1, end)

		// 批量生成向量
		vectors, err := embedder.EmbedStrings(ctx, batchChunks)
		if err != nil {
			zap.L().Fatal("EmbedStrings failed", zap.Error(err))
		}

		// 将文本和对应的向量组装成数据库模型
		for j, vec := range vectors {
			allDocs = append(allDocs, &model.KnowledgeDoc{
				Content: batchChunks[j],
				Source:  filePath,
				Vector:  pgvector.NewVector(vec),
			})
		}
	}

	fmt.Println("向量化全部完成，准备写入数据库...")
	if err := vectorDB.SaveChunks(ctx, allDocs); err != nil {
		zap.L().Fatal("SaveChunks failed", zap.Error(err))
	}

	fmt.Printf("成功将 %d 条知识库片段灌入 PostgreVector \n", len(allDocs))
}
