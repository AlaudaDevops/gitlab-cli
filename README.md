# GitLab CLI SDK

GitLab 用户和项目自动化管理工具，基于官方 GitLab Go SDK 开发。

## 特性

- ✅ 使用官方 GitLab Go SDK (`gitlab.com/gitlab-org/api/client-go`)
- ✅ 无需外部依赖，纯 Go 实现
- ✅ 类型安全的 API 调用
- ✅ 模块化设计，易于维护和扩展
- ✅ 支持批量创建和管理 GitLab 用户、组和项目
- ✅ 支持通过 YAML 配置文件批量操作

## 快速开始

### 前置要求

- Go 1.23.0 或更高版本
- GitLab 管理员权限的 Personal Access Token (需要 `api` + `sudo` scopes)

### 安装

```bash
# 克隆仓库
git clone <repository-url>
cd gitlab-cli-sdk

# 构建
make build

# 或者直接安装到系统
make install
```

### 使用

```bash
# 设置环境变量
export GITLAB_HOST=https://your-gitlab-instance.com
export GITLAB_TOKEN=your-personal-access-token

# 创建用户、组和项目
./bin/gitlab-cli user create -f config.yaml

# 清理用户及其资源
./bin/gitlab-cli user cleanup -f config.yaml
```

## 项目结构

```
gitlab-cli-sdk/
├── cmd/
│   └── gitlab-cli/        # 命令行入口
├── internal/              # 内部包（不对外暴露）
│   ├── cli/               # CLI 命令定义
│   ├── config/            # 配置管理
│   ├── processor/         # 业务逻辑处理
│   └── utils/             # 工具函数
├── pkg/                   # 公共包（可被外部使用）
│   ├── client/            # GitLab 客户端
│   └── types/             # 数据类型定义
├── docs/                  # 文档
├── bin/                   # 编译输出
└── Makefile               # 构建脚本
```

## 文档

- [架构设计](docs/ARCHITECTURE.md) - 详细的代码架构说明
- [快速开始](docs/QUICKSTART.md) - 快速开始指南
- [详细说明](docs/README.md) - 完整使用文档

## 开发

```bash
# 格式化代码
make fmt

# 运行代码检查
make lint

# 运行测试
make test

# 构建所有平台
make build-all

# 创建发布包
make release
```

## Makefile 命令

运行 `make help` 查看所有可用命令：

```bash
make help
```

## 许可证

[许可证类型]

## 贡献

欢迎贡献代码！请阅读 [贡献指南](CONTRIBUTING.md) 了解详情。
