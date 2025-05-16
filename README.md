# CineLight API

CineLight API 是一个用 Go 语言开发的电影信息管理系统后端服务，提供完整的 RESTful API 接口，允许用户创建、查询、更新和删除电影信息，并支持用户认证和权限管理。

## 功能特点

- **电影管理**：CRUD 操作，支持电影信息的增删改查
- **用户系统**：用户注册、激活和认证
- **权限控制**：基于令牌的认证和细粒度的权限管理
- **速率限制**：防止 API 滥用
- **JSON 日志**：结构化日志输出
- **CORS 支持**：配置跨源资源共享
- **数据库迁移**：使用 migrate 工具管理数据库版本
- **Docker 支持**：容器化部署

## 技术栈

- **语言**：Go
- **Web 框架**：Gin
- **数据库**：PostgreSQL
- **认证**：JWT（JSON Web Token）
- **容器化**：Docker & Docker Compose
- **配置管理**：环境变量 & 命令行参数

## 目录结构

```
.
├── bin/                  # 编译生成的二进制文件
├── cmd/                  # 应用程序入口
│   ├── api/              # API 服务主要代码
│   └── examples/         # 示例代码
├── internal/             # 私有应用程序代码
│   ├── data/             # 数据模型和数据库交互
│   ├── jsonlog/          # JSON 日志工具
│   ├── mailer/           # 邮件服务
│   └── validator/        # 输入验证
├── migrations/           # 数据库迁移文件
├── vendor/               # 依赖包
├── .envrc                # 环境变量
├── docker-compose.yml    # Docker 配置
├── Dockerfile            # Docker 镜像构建
├── go.mod                # Go 模块依赖
├── go.sum                # Go 依赖校验和
└── Makefile              # 构建和管理工具
```

## 安装和运行

### 先决条件

- Go 1.16+
- PostgreSQL
- Docker & Docker Compose（可选，用于容器化部署）

### 本地开发环境设置

1. 克隆仓库

```bash
git clone https://your-repository-url/CineLight-API.git
cd CineLight-API
```

2. 创建 `.envrc` 文件并设置环境变量

```
export CINELIGHT_DB_DSN="postgres://postgres:123456@localhost:5432/cinelight?sslmode=disable"
```

3. 启动数据库（使用 Docker Compose）

```bash
docker-compose up -d postgres
```

4. 运行数据库迁移

```bash
make db/migrations/up
```

5. 启动 API 服务

```bash
make run/api
```

### 使用 Docker Compose 部署

```bash
docker-compose up -d
```

## API 端点

### 健康检查

- `GET /v1/healthcheck` - 检查 API 服务状态

### 用户管理

- `POST /v1/users` - 注册新用户
- `PUT /v1/users/activated` - 激活用户账户

### 认证

- `POST /v1/tokens/authentication` - 创建认证令牌（登录）

### 电影管理（需要认证）

- `GET /v1/movies` - 获取电影列表（支持过滤、分页和排序）
- `POST /v1/movies` - 创建新电影（需要写权限）
- `GET /v1/movies/:id` - 获取特定电影详情（需要读权限）
- `PATCH /v1/movies/:id` - 更新电影信息（需要写权限）
- `DELETE /v1/movies/:id` - 删除电影（需要写权限）

## 常用命令

CineLight API 使用 Makefile 简化常见操作：

- `make help` - 显示帮助信息
- `make run/api` - 启动 API 服务
- `make db/psql` - 连接到 PostgreSQL 数据库
- `make db/migration/new name=<迁移名称>` - 创建新的数据库迁移文件
- `make db/migrations/up` - 执行数据库迁移
- `make audit` - 格式化代码并运行静态检查和测试
- `make vendor` - 整理和备份依赖包
- `make build/api` - 构建 API 二进制文件

## 配置选项

通过命令行参数或环境变量配置服务：

- `-port` - API 服务端口（默认：3939）
- `-env` - 运行环境（development|staging|production）
- `-db-dsn` - PostgreSQL 数据源名称
- `-db-max-open-conns` - 最大数据库连接数
- `-db-max-idle-conns` - 最大空闲连接数
- `-db-max-idle-time` - 连接最大空闲时间
- `-limiter-rps` - 速率限制（每秒请求数）
- `-limiter-burst` - 速率限制突发值
- `-limiter-enabled` - 是否启用速率限制
- `-smtp-*` - SMTP 服务器配置
- `-cors-trusted-origins` - 受信任的 CORS 来源

## 许可证

[许可证信息] 
