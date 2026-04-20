package model

import (
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// KnowledgeDoc 知识库文档分块表
type KnowledgeDoc struct {
	gorm.Model
	Content string          `gorm:"type:text;not null"`
	Source  string          `gorm:"type:varchar(255)"`
	Vector  pgvector.Vector `gorm:"type:vector(1024)"` // 1536 是 OpenAI 向量的默认维度，1024 是 阿里云的，如果你用别的模型请修改这个数字
}
