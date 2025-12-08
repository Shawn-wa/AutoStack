# ⚡ AutoStack

**低代码快捷部署平台** - 一键部署您的应用程序

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D?style=flat&logo=vue.js)](https://vuejs.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://docker.com)

## 📖 简介

AutoStack 是一个基于 Go + Vue 构建的低代码快捷部署平台，旨在简化应用程序的部署流程。通过可视化界面和预配置模板，让部署变得简单高效。

### ✨ 特性

- 🚀 **一键部署** - 选择模板，填写配置，一键启动
- 📦 **模板市场** - 丰富的预配置部署模板
- 🎨 **低代码配置** - 可视化配置界面，无需编写复杂配置
- 🔄 **多环境支持** - 开发、测试、生产环境一键切换
- 📊 **实时监控** - 查看部署状态、资源使用和日志
- 🔐 **安全可靠** - JWT 认证，权限控制

## 🏗️ 技术栈

### 后端
- **Go 1.21+** - 高性能后端语言
- **Gin** - 轻量级 Web 框架
- **GORM** - ORM 框架
- **Viper** - 配置管理
- **JWT** - 身份认证

### 前端
- **Vue 3** - 渐进式 JavaScript 框架
- **TypeScript** - 类型安全
- **Vite** - 下一代构建工具
- **Pinia** - 状态管理
- **Vue Router** - 路由管理

### 部署
- **Docker** - 容器化
- **Docker Compose** - 编排工具
- **Nginx** - 反向代理

## 📁 项目结构

```
AutoStack/
├── backend/                 # Go 后端
│   ├── cmd/server/         # 入口文件
│   ├── internal/           # 内部包
│   │   ├── api/           # API 服务
│   │   ├── config/        # 配置
│   │   ├── handler/       # 处理器
│   │   ├── middleware/    # 中间件
│   │   ├── model/         # 数据模型
│   │   ├── repository/    # 数据仓库
│   │   └── service/       # 业务逻辑
│   ├── pkg/               # 公共包
│   ├── config.yaml        # 配置文件
│   ├── Dockerfile
│   └── go.mod
├── frontend/               # Vue 前端
│   ├── src/
│   │   ├── api/           # API 请求
│   │   ├── components/    # 组件
│   │   ├── composables/   # 组合式函数
│   │   ├── layouts/       # 布局
│   │   ├── pages/         # 页面
│   │   ├── stores/        # 状态管理
│   │   ├── styles/        # 样式
│   │   └── types/         # 类型定义
│   ├── Dockerfile
│   ├── nginx.conf
│   └── package.json
├── docker-compose.yml      # Docker 编排
├── Makefile               # 常用命令
└── README.md
```

## 🚀 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose (可选)

### 本地开发

**1. 克隆项目**
```bash
git clone https://github.com/your-username/autostack.git
cd autostack
```

**2. 启动后端**
```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

**3. 启动前端**
```bash
cd frontend
npm install
npm run dev
```

**4. 访问应用**
- 前端: http://localhost:3000
- 后端 API: http://localhost:8080

### Docker 部署

**一键启动**
```bash
docker-compose up -d
```

**查看日志**
```bash
docker-compose logs -f
```

**停止服务**
```bash
docker-compose down
```

## 📚 API 文档

### 认证接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/register` | 用户注册 |

### 项目管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/projects` | 项目列表 |
| POST | `/api/v1/projects` | 创建项目 |
| GET | `/api/v1/projects/:id` | 项目详情 |
| PUT | `/api/v1/projects/:id` | 更新项目 |
| DELETE | `/api/v1/projects/:id` | 删除项目 |

### 部署管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/deployments` | 部署列表 |
| POST | `/api/v1/deployments` | 创建部署 |
| GET | `/api/v1/deployments/:id` | 部署详情 |
| POST | `/api/v1/deployments/:id/start` | 启动部署 |
| POST | `/api/v1/deployments/:id/stop` | 停止部署 |

### 模板管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/templates` | 模板列表 |
| POST | `/api/v1/templates` | 创建模板 |
| GET | `/api/v1/templates/:id` | 模板详情 |

## ⚙️ 配置说明

### 后端配置 (config.yaml)

```yaml
server:
  port: "8080"
  mode: "debug"  # debug, release, test

database:
  driver: "sqlite"  # sqlite, mysql
  dsn: "autostack.db"

jwt:
  secret: "your-secret-key"
  expire_hour: 24
```

### 环境变量

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `SERVER_PORT` | 服务端口 | 8080 |
| `SERVER_MODE` | 运行模式 | debug |
| `DATABASE_DRIVER` | 数据库驱动 | sqlite |
| `DATABASE_DSN` | 数据库连接 | autostack.db |
| `JWT_SECRET` | JWT 密钥 | - |

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

