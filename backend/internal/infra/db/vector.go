package db

import (
	"backend/internal/infra/llm"
	"backend/internal/model"
	"context"
	"fmt"
	"strings"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type VectorDB struct {
	db       *gorm.DB
	embedder llm.Embedder
}

func NewVectorDB(db *gorm.DB, embedder llm.Embedder) *VectorDB {
	return &VectorDB{db: db, embedder: embedder}
}

func (v *VectorDB) SearchRelevantChunks(ctx context.Context, query string, topK int) (string, error) {
	// 1. 将用户的查询字符串变成向量
	queryVector, err := v.embedder.EmbedString(ctx, query)
	if err != nil {
		return "", fmt.Errorf("生成查询向量失败: %w", err)
	}

	var docs []model.KnowledgeDoc

	// 2. 使用 pgvector 的 <=> 符号计算余弦距离
	// 按相似度从高到低排序，取前 TopK 个
	err = v.db.WithContext(ctx).
		Order(gorm.Expr("vector <=> ?", pgvector.NewVector(queryVector))).
		Limit(topK).
		Find(&docs).Error

	if err != nil {
		return "", fmt.Errorf("数据库检索失败: %w", err)
	}

	// 3. 把查出来的文本拼接成一段大文本返回
	if len(docs) == 0 {
		return "未在知识库中找到相关信息。", nil
	}

	var builder strings.Builder
	for i, doc := range docs {
		fmt.Fprintf(&builder, "内容片段 %d\n%s \n\n", i+1, doc.Content)
	}

	return builder.String(), nil
}

func (v *VectorDB) SaveChunks(ctx context.Context, docs []*model.KnowledgeDoc) error {
	return v.db.WithContext(ctx).Create(&docs).Error
}
