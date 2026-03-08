import os
from typing import TypedDict, Annotated, Sequence

from langchain_core.messages import BaseMessage
from langchain_openai import ChatOpenAI
from langgraph.constants import START, END
from langgraph.graph import add_messages, StateGraph

class ChatGraph:
    @staticmethod
    def create_app():
        llm = ChatOpenAI(
            model="deepseek-v3.2",
            openai_api_key=os.getenv("API_KEY"),
            openai_api_base=os.getenv("API_BASE"),
            streaming=True, # 流式输出
            model_kwargs={
                "stream_options": {
                    "include_usage": True,  # 输出token消耗数量
                }
            }
        )

        class AgentState(TypedDict):
            messages: Annotated[Sequence[BaseMessage], add_messages]

        def model_call(state: AgentState) -> AgentState:
            res = llm.invoke(state["messages"]) # 调用大模型
            return {"messages": [res]}

        # 创建图，类型为AgentState
        graph = StateGraph(AgentState)
        # 添加新节点agent，start->agent->end
        graph.add_node('agent', model_call)
        # 添加边
        graph.add_edge(START, "agent")
        graph.add_edge("agent", END)
        # 编译返回
        return graph.compile()