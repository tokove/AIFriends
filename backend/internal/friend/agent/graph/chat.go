package graph

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

type ChatState struct {
	Messages []*schema.Message
}

func NewChatGraph(ctx context.Context, llm model.ToolCallingChatModel, tools []tool.BaseTool) (compose.Runnable[ChatState, *schema.Message], error) {
	g := compose.NewGraph[ChatState, *schema.Message]()

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: llm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: tools,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("初始化 react agent 失败: %w", err)
	}

	if err := g.AddLambdaNode("agent", compose.StreamableLambda(
		func(ctx context.Context, msgs []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
			return agent.Stream(ctx, msgs)
		},
	)); err != nil {
		return nil, fmt.Errorf("添加 agent 节点失败: %w", err)
	}

	if err := g.AddLambdaNode("extract_messages", compose.InvokableLambda(
		func(ctx context.Context, input ChatState) ([]*schema.Message, error) {
			return input.Messages, nil
		},
	)); err != nil {
		return nil, fmt.Errorf("添加 extract_messages 节点失败: %w", err)
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
		return nil, fmt.Errorf("编译 ChatGraph 失败: %w", err)
	}

	return runnable, nil
}
