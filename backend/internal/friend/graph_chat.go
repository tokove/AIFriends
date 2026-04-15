package friend

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type ChatState struct {
	Messages []*schema.Message
}

func NewChatGraph(ctx context.Context, llm model.ToolCallingChatModel) (compose.Runnable[ChatState, *schema.Message], error) {
	g := compose.NewGraph[ChatState, *schema.Message]()

	if err := g.AddLambdaNode("extract_messages", compose.InvokableLambda(
		func(ctx context.Context, input ChatState) ([]*schema.Message, error) {
			return input.Messages, nil
		},
	)); err != nil {
		return nil, fmt.Errorf("添加 extract_messages 节点失败: %w", err)
	}

	if err := g.AddChatModelNode("llm", llm); err != nil {
		return nil, fmt.Errorf("添加 llm 节点失败: %w", err)
	}
	if err := g.AddEdge(compose.START, "extract_messages"); err != nil {
		return nil, fmt.Errorf("添加 START->extract 边失败: %w", err)
	}
	if err := g.AddEdge("extract_messages", "llm"); err != nil {
		return nil, fmt.Errorf("添加 extract->llm 边失败: %w", err)
	}
	if err := g.AddEdge("llm", compose.END); err != nil {
		return nil, fmt.Errorf("添加 llm->END 边失败: %w", err)
	}

	runnable, err := g.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译 MemoryGraph 失败: %w", err)
	}

	return runnable, nil
}
