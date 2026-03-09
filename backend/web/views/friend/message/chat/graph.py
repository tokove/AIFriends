import os
from pprint import pprint
from typing import TypedDict, Annotated, Sequence

from django.utils.timezone import now, localtime
from langchain_core.messages import BaseMessage
from langchain_core.tools import tool
from langchain_openai import ChatOpenAI
from langgraph.constants import START, END
from langgraph.graph import add_messages, StateGraph
from langgraph.prebuilt import ToolNode


class ChatGraph:
    @staticmethod
    def create_app():
        @tool
        def get_time() -> str:
            """当需要查询精确时间时，调用此函数。返回格式为： [年-月-日 时:分:秒]"""
            return localtime(now()).strftime("%Y-%m-%d %H:%M:%S")

        tools = [get_time]

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
        ).bind_tools(tools)

        class AgentState(TypedDict):
            messages: Annotated[Sequence[BaseMessage], add_messages]

        def model_call(state: AgentState) -> AgentState:
            pprint(state)
            res = llm.invoke(state["messages"]) # 调用大模型
            return {"messages": [res]}

        def should_continue(state: AgentState) -> str:
            last_message = state["messages"][-1]
            if last_message.tool_calls:
                return "tools"
            return "end"

        tool_node = ToolNode(tools)
        # 创建图，类型为AgentState
        graph = StateGraph(AgentState)
        # 添加新节点agent，start->agent->end
        graph.add_node('agent', model_call)
        graph.add_node('tools', tool_node)
        # 添加边
        graph.add_edge(START, "agent")
        graph.add_conditional_edges(
            "agent",
            should_continue,
            {
                "tools": "tools",
                "end": END,
            }
        )
        graph.add_edge("tools", "agent")
        # 编译返回
        return graph.compile()