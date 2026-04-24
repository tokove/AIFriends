# AIFriends

AIFriends 是一个基于 Go 和 Vue 的 AI 陪伴聊天项目，提供用户系统、角色创建、好友会话、流式聊天和长期记忆能力。

## 项目结构

```text
AIFriends/
├── backend/                # Go 后端
│   ├── cmd/                # 程序入口
│   ├── configs/            # 配置文件
│   ├── documents/          # 知识库数据
│   ├── internal/           # 核心业务代码
│   ├── media/              # 上传与静态资源
│   ├── static/frontend/    # 前端打包产物
│   └── pkg/                # 公共工具与常量
└── frontend/               # Vue 3 + Vite 前端
```

## 技术栈

- 后端：Gin、GORM、PostgreSQL、Redis、Zap
- 前端：Vue 3、Vue Router、Pinia、Vite、Tailwind CSS、daisyUI
- AI 能力：Agent Graph、Tool 调用、Vector 检索、长期记忆

## 核心能力

- 用户注册、登录、刷新令牌、资料编辑
- 角色创建、编辑、删除、列表查询
- 好友会话创建与删除
- 流式 AI 对话与历史消息查询
- 长期记忆摘要与知识检索增强

## API 概览

### 用户模块

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| POST | `/api/user/account/register` | 公开 | 注册 |
| POST | `/api/user/account/login` | 公开 | 登录 |
| POST | `/api/user/account/refresh_token` | 公开 | 刷新访问令牌 |
| POST | `/api/user/account/logout` | 私有 | 退出登录 |
| GET | `/api/user/account/get_user_info` | 私有 | 获取当前用户信息 |
| POST | `/api/user/profile/update/` | 私有 | 更新用户资料 |

### 角色模块

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/create/character/get_list` | 公开 | 角色列表 |
| GET | `/api/homepage/index` | 公开 | 首页与搜索 |
| POST | `/api/create/character/create` | 私有 | 创建角色 |
| POST | `/api/create/character/update` | 私有 | 更新角色 |
| GET | `/api/create/character/get_single` | 私有 | 角色详情 |
| POST | `/api/create/character/remove` | 私有 | 删除角色 |

### 好友与聊天模块

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| POST | `/api/friend/get_or_create` | 私有 | 创建或获取会话 |
| GET | `/api/friend/get_list` | 私有 | 会话列表，游标分页 |
| POST | `/api/friend/remove` | 私有 | 删除会话 |
| POST | `/api/friend/message/chat` | 私有 | 流式聊天 |
| GET | `/api/friend/message/get_history` | 私有 | 历史消息 |

## AI 架构

核心组件：

- `ChatModel`：对话模型
- `Embedding`：向量化模型
- `VectorDB`：语义检索
- `Tool`：工具调用
- `ChatGraph`：对话编排
- `MemoryGraph`：长期记忆编排

执行链路：

```text
用户输入 -> Graph -> Tool 调用 -> Vector 检索 -> LLM 生成 -> 流式返回
```

## 环境要求

- Go
- Node.js 20 或更高版本
- PostgreSQL
- Redis

## 后端配置

后端配置文件位于 `backend/configs/config.yaml`，默认服务端口为 `8080`。

首次运行前，请在 `backend` 目录创建 `.env` 文件，并按 `.env.example` 填写必要配置。

## 本地开发

### 启动后端

在 `backend` 目录执行：

```bash
go mod tidy
go run cmd/importer/main.go
go run cmd/server/main.go
```

### 启动前端

在 `frontend` 目录执行：

```bash
npm install
npm run dev
```

前端开发服务器默认运行在 `http://localhost:5173`，并代理 `/api` 到 `http://127.0.0.1:8080`。

## 前后端一体化部署

在 `frontend` 目录执行：

```bash
npm install
npm run build
```

构建产物会输出到 `backend/static/frontend`。完成后启动后端，直接访问 `http://localhost:8080` 即可加载前端页面。

## 资源路径

| 类型 | 路径 |
| --- | --- |
| 用户头像 | `backend/media/user/photos` |
| 角色图片 | `backend/media/character/photos` |
| 聊天背景图 | `backend/media/character/background_images` |
| 前端打包产物 | `backend/static/frontend` |

## 关键设计

- H-S-R 分层架构，便于拆分职责
- 统一 `model` 层，减少循环依赖
- 中间件处理鉴权、跨域和限流
- 通过 Agent Graph 组织对话与记忆流程
- 支持流式输出和长期记忆更新

## 后续规划

- 完善首页召回与推荐能力
- 补充测试与部署文档
- 优化会话列表和角色广场体验
