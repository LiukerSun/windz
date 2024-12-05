# Windz Backend

这是Windz项目的后端服务，使用Go语言开发的现代化Web后端服务。

## 项目架构

本项目采用清晰的分层架构设计：

```
backend/
├── cmd/           # 主程序入口
├── config/        # 配置文件
├── internal/      # 内部包
├── pkg/          # 可重用的库代码
└── logs/         # 日志文件
```

## 技术栈

- 开发语言：Go 1.23
- Web框架：Gin
- 数据库支持：
  - MySQL
  - PostgreSQL
  - SQLite
- ORM：GORM
- 配置管理：Viper
- 日志系统：Zap
- 认证：JWT

## 主要特性

- RESTful API设计
- JWT用户认证
- 多数据库支持
- 结构化日志
- 配置热重载
- 模块化架构

## 快速开始

1. 确保已安装Go 1.23或以上版本
2. 克隆项目
3. 安装依赖：
```bash
go mod download
```
4. 运行项目：
```bash
go run cmd/main.go
```

## 项目结构说明

- `cmd/`: 包含主程序入口文件
- `config/`: 配置文件目录
- `internal/`: 内部应用代码
  - `api/`: API处理器
  - `middleware/`: 中间件
  - `model/`: 数据模型
  - `service/`: 业务逻辑
  - `repository/`: 数据访问层
- `pkg/`: 可重用的公共包
- `logs/`: 日志文件目录

## API 文档

项目提供了Swagger API文档，您可以通过访问以下地址查看和测试API：

```
http://localhost:8080/swagger/index.html
```

## 默认密码

在首次运行时，系统会创建一个超级管理员账号，默认密码为 `admin123`。您可以在配置文件中修改此密码：

```yaml
app:
  default_password: "admin123" # 默认密码，用于初始化超级管理员账号
```

## 开源协议

MIT License

## 贡献指南

欢迎提交Issue和Pull Request来帮助改进项目。
