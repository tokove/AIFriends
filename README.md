# 🤖 AIFriends - AI 陪伴聊天平台

AIFriends 是一款基于 Go 语言构建的高性能 AI 陪伴聊天平台后端。项目采用工业标准的 H-S-R (Handler-Service-Repository) 分层架构，并引入 Redis 与 PostgreSQL 双引擎优化，确保业务逻辑清晰、资源处理闭环。

---

## 🏗️ 目录结构
```bash
backend/
├── cmd/server          # 🚀 程序入口 (main.go)
├── configs/            # ⚙️ 配置文件 (.yaml, .env.example)
├── data/               # 📂 静态资源物理存储
│   ├── user/           # 👤 用户头像上传目录
│   └── character/      # 🎭 角色资源存储 (photos/background_images)
├── internal/           # 🔒 核心私有业务逻辑
│   ├── config/         # Viper + godotenv 配置解析
│   ├── infra/          # 🔌 基础设施 (PostgreSQL, Redis, Zap Logger)
│   ├── middleware/     # 🛡️ 中间件 (JWT 鉴权、CORS、Auth)
│   ├── model/          # 📦 统一数据模型 (解决循环依赖的关键)
│   ├── router/         # 🛣️ 路由分发 (SetupRouter)
│   ├── user/           # 👤 用户领域模块 (Handler/Service/Repo)
│   └── character/      # 🎭 角色领域模块 (Handler/Service/Repo)
├── pkg/                # 📦 公共工具包
│   ├── constants/      # 📏 全局业务常量 (长度限制、文件路径)
│   └── utils/          # 🛠️ 常用工具 (JWT 处理、文件补偿逻辑)
└── log/                # 📝 系统运行日志
```
---

## 🛠️ 技术栈

* 核心框架: Gin (高性能 Web 路由)
* 持久层: GORM + PostgreSQL (严谨的 ACID 事务支持)
* 缓存层: Redis (用于状态管理、验证码存储及高性能缓存)
* 配置管理: Viper + godotenv (支持多环境配置)
* 安全认证: JWT (双 Token 无感刷新机制) + Bcrypt (密码哈希加密)
* 日志系统: Zap (结构化高性能日志)

---

## 📡 API 接口预览

### 👤 用户模块 (User)
| 方法 | 路径 | 权限 | 说明 |
| :--- | :--- | :--- | :--- |
| POST | /api/user/account/register | 公开 | 用户注册 (用户名 2-32位, 密码 8-72位) |
| POST | /api/user/account/login | 公开 | 用户登录 (下发 AccessToken) |
| POST | /api/user/account/refresh_token| 公开 | 利用刷新令牌重续 AccessToken |
| POST | /api/user/account/logout | 私有 | 退出登录 (清除鉴权状态) |
| GET  | /api/user/account/get_user_info | 私有 | 获取当前登录用户的详细资料 |
| POST | /api/user/profile/update/ | 私有 | 更新个人资料 (支持头像上传) |

### 🎭 角色模块 (Character)
| 方法 | 路径 | 权限 | 说明 |
| :--- | :--- | :--- | :--- |
| POST | /api/create/character/create | 私有 | 创建 AI 角色 (头像/背景图双上传校验) |
| POST | /api/create/character/update | 私有 | 更新角色信息 (具备文件一致性补偿逻辑) |
| POST | /api/create/character/remove | 私有 | 软删除角色 (保留物理文件以备审计) |
| GET  | /api/create/character/get_single| 私有 | 获取单个角色详情 |
| GET  | /api/create/character/get_list | 私有 | 获取当前用户的角色列表 (按时间倒序) |

---


## 🚀 快速启动

1. 环境配置:
```
   cp configs/.env.example configs/.env
   # 根据本地 PostgreSQL 和 Redis 连接信息修改 configs/.env
```

2. 安装依赖:
```
   go mod tidy
```
3. 运行项目:
```
   go run cmd/server/main.go
```