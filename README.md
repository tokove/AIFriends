# 🤖 AIFriends - AI 陪伴聊天平台

AIFriends 是一款基于 Go 语言构建的高性能 AI 陪伴聊天平台后端。项目采用工业标准的 **H-S-R (Handler-Service-Repository)** 分层架构，针对 **PostgreSQL** 数据持久化与静态资源管理进行了深度优化，确保业务逻辑清晰、资源处理闭环。

---

## 🏗️ 目录结构

```text
backend/
├── cmd/server          # 🚀 程序入口 (main.go)
├── configs/            # ⚙️ 配置文件 (.yaml, .env.example)
├── data/               # 📂 静态资源物理存储
│   └── user/photos     # 👤 用户头像上传目录
├── internal/           # 🔒 核心私有业务逻辑
│   ├── config/         # Viper + godotenv 配置解析
│   ├── infra/          # 🔌 基础设施 (PostgreSQL 驱动、Zap Logger)
│   ├── middleware/     # 🛡️ 中间件 (JWT 鉴权、CORS、日志回收)
│   ├── router/         # 🛣️ 路由分发 (SetupRouter)
│   └── user/           # 👤 用户领域模块 (Model/Handler/Service/Repo)
├── pkg/                # 📦 公共工具包
│   ├── constants/      # 📏 全局业务常量 (长度限制、文件路径)
│   └── utils/          # 🛠️ 常用工具 (JWT 处理、文件置换逻辑)
└── log/                # 📝 系统运行日志
```

---

## 🛠️ 技术栈

* **核心框架**: `Gin` (高性能 Web 路由)
* **持久层**: `GORM` + `PostgreSQL` (严谨的 ACID 事务支持)
* **配置管理**: `Viper` + `godotenv` (支持 .env 敏感信息与 .yaml 结构化配置)
* **安全认证**: `JWT` (双 Token 无感刷新机制) + `Bcrypt` (密码哈希加密)
* **日志系统**: `Zap` (结构化高性能日志)

---

## 📡 API 接口预览

### 🔓 用户账密 (Public)
| 方法 | 路径 | 说明 |
| :--- | :--- | :--- |
| POST | `/api/user/account/register` | 用户注册 (用户名 2-32位, 密码 8-72位) |
| POST | `/api/user/account/login` | 用户登录 (下发 AccessToken 并设置 Refresh Cookie) |
| POST | `/api/user/account/refresh_token`| 利用刷新令牌重续 AccessToken |

### 🔐 用户资料 (Protected)
| 方法 | 路径 | 说明 |
| :--- | :--- | :--- |
| POST | `/api/user/account/logout` | 退出登录 (清除鉴权状态) |
| GET  | `/api/user/account/get_user_info` | 获取当前登录用户的详细资料 |
| POST | `/api/user/profile/update/` | 更新个人资料 (支持头像 Form-Data 上传) |


---

## 🚀 快速启动

1.  **环境配置**:
    ```bash
    cp configs/.env.example configs/.env
    # 根据本地 PostgreSQL 连接信息修改 configs/.env
    ```

2.  **安装依赖**:
    ```bash
    go mod tidy
    ```

3.  **运行项目**:
    ```bash
    go run cmd/server/main.go
    ```