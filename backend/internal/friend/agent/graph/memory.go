package graph

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// MemoryState 记忆提炼图的输入状态
type MemoryState struct {
	Messages []*schema.Message
}

// NewMemoryGraph 编译并返回记忆专用的执行图
func NewMemoryGraph(ctx context.Context, llm model.ToolCallingChatModel) (compose.Runnable[MemoryState, *schema.Message], error) {
	g := compose.NewGraph[MemoryState, *schema.Message]()

	if err := g.AddLambdaNode("extract_messages", compose.InvokableLambda(
		func(ctx context.Context, input MemoryState) ([]*schema.Message, error) {
			return input.Messages, nil
		},
	)); err != nil {
		return nil, fmt.Errorf("添加 extract_messages 节点失败: %w", err)
	}

	if err := g.AddChatModelNode("agent", llm); err != nil {
		return nil, fmt.Errorf("添加 agent 节点失败: %w", err)
	}
	if err := g.AddEdge(compose.START, "extract_messages"); err != nil {
		return nil, fmt.Errorf("添加 START->extract 边失败: %w", err)
	}
	if err := g.AddEdge("extract_messages", "agent"); err != nil {
		return nil, fmt.Errorf("添加 extract->agent 边失败: %w", err)
	}
	if err := g.AddEdge("agent", compose.END); err != nil {
		return nil, fmt.Errorf("添加 agent->END 边失败: %w", err)
	}

	runnable, err := g.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译 MemoryGraph 失败: %w", err)
	}

	return runnable, nil
}
