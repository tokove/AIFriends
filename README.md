# 🤖✨ AIFriends - AI 陪伴聊天平台

AIFriends 是一个基于 Go 构建的高性能 AI 陪伴聊天后端服务 💬
采用 H-S-R 分层架构，融合 PostgreSQL + Redis + AI Agent，实现稳定、高扩展的智能聊天系统 🚀

---

## 🏗️📦 项目结构

```bash
backend/
├── cmd/                    # 🚀 程序入口
│   ├── importer/           # 📥 数据导入工具
│   └── server/             # 🌐 主服务入口 (main.go)
│
├── configs/                # ⚙️ 配置文件 (.env / yaml)
│
├── documents/              # 📚 知识库数据
│   └── knowledge/
│
├── internal/               # 🔒 核心业务
│   ├── config/             # 🧩 配置加载 (Viper + godotenv)
│   ├── infra/              # 🔌 基础设施层
│   │   ├── db/             # 🗄️ PostgreSQL + VectorDB
│   │   ├── redis/          # ⚡ Redis
│   │   ├── logger/         # 📝 Zap 日志
│   │   └── llm/            # 🤖 大模型封装
│   │
│   ├── middleware/         # 🛡️ JWT / CORS / 限流
│   ├── model/              # 📦 数据模型
│   ├── router/             # 🛣️ 路由注册
│   │
│   ├── user/               # 👤 用户模块
│   ├── character/          # 🎭 角色模块
│   ├── friend/             # 💬 好友、聊天模块
│   │   └── agent/          # 🧠 AI Agent
│   │       ├── graph/      # 🔗 对话、记忆编排
│   │       └── tool/       # 🧰 工具系统（RAG等）
│   │
│   └── task/               # ⏱️ 定时任务
│
├── media/                  # 📂 静态资源
│   ├── user/photos/        # 👤 用户头像
│   └── character/
│       ├── photos/         # 🎭 角色图片
│       └── background_images/ # 🖼️ 背景图
│
├── log/                    # 📝 日志
│
└── pkg/                    # 📦 工具库
    ├── constants/
    └── utils/
```

---

## 🛠️⚙️ 技术栈

* 🌐 Gin：高性能 Web 框架
* 🗄️ GORM：ORM
* 🐘 PostgreSQL：主数据库
* ⚡ Redis：缓存 / 状态管理
* 🧩 Viper + godotenv：配置管理
* 🔐 JWT：鉴权系统
* 📝 Zap：高性能日志
* 🤖 AI Agent：Graph + Tool 执行架构
* 🔍 VectorDB：语义检索（RAG）

---

## 📡📌 API 接口

### 👤 用户模块

| 方法   | 路径                              | 权限    | 说明       |
| ---- | ------------------------------- | ----- | -------- |
| POST | /api/user/account/register      | 🌍 公开 | 注册       |
| POST | /api/user/account/login         | 🌍 公开 | 登录       |
| POST | /api/user/account/refresh_token | 🌍 公开 | 刷新 Token |
| POST | /api/user/account/logout        | 🔒 私有 | 退出       |
| GET  | /api/user/account/get_user_info | 🔒 私有 | 用户信息     |
| POST | /api/user/profile/update/       | 🔒 私有 | 更新资料     |

---

### 🎭 角色模块

| 方法   | 路径                               | 权限    | 说明    |
| ---- | -------------------------------- | ----- | ----- |
| GET  | /api/create/character/get_list   | 🌍 公开 | 角色列表  |
| GET  | /api/homepage/index              | 🌍 公开 | 首页/搜索 |
| POST | /api/create/character/create     | 🔒 私有 | 创建角色  |
| POST | /api/create/character/update     | 🔒 私有 | 更新角色  |
| GET  | /api/create/character/get_single | 🔒 私有 | 单个角色  |
| POST | /api/create/character/remove     | 🔒 私有 | 删除角色  |

---

### 💬 好友、聊天模块

| 方法   | 路径                              | 权限    | 说明           |
| ---- | ------------------------------- | ----- | ------------ |
| POST | /api/friend/get_or_create       | 🔒 私有 | 创建/获取会话      |
| GET  | /api/friend/get_list            | 🔒 私有 | 会话列表         |
| POST | /api/friend/remove              | 🔒 私有 | 删除会话         |
| POST | /api/friend/message/chat        | 🔒 私有 | 🤖 AI 对话（流式） |
| GET  | /api/friend/message/get_history | 🔒 私有 | 📜 历史消息      |

---

## 🧠 AI 架构

核心组件：

* 🤖 ChatModel：对话模型
* 🧬 Embedding：向量化
* 📊 VectorDB：语义检索
* 🧰 Tool：工具调用
* 🔗 ChatGraph：对话流程编排
* 🧠 MemoryGraph：长期记忆编排

执行链路：

```
用户输入 → Graph → Tool调用 → Vector检索 → LLM生成 → 流式返回
```

---

## ⚙️🔧 环境配置

```bash
在backend目录下新建.env文件，按照.env.example填写
```

---

## 🚀 启动项目

在 `backend` 目录执行：

```bash
go mod tidy
go run cmd/importer/main.go
go run cmd/server/main.go 
```

---

## 📌 关键设计

* 🧱 H-S-R 分层架构（解耦清晰）
* 📦 Model 统一层（解决循环依赖）
* 🔌 Infra 层（统一外部依赖）
* 🛡️ 中间件体系（鉴权 + 限流）
* 🤖 Agent 架构（复杂 AI 行为编排）
* 🔍 RAG 能力（知识增强）
* ⚡ 流式输出（更自然的对话体验）

---

## 📂 资源路径

| 类型      | 路径                                |
| ------- | --------------------------------- |
| 👤 用户头像 | media/user/photos                 |
| 🎭 角色图片 | media/character/photos            |
| 🖼️ 背景图 | media/character/background_images |

---

## 📈 后续规划

*  📊 首页召回式

     
