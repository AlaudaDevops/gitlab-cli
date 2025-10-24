# GitLab CLI SDK

基于官方 GitLab Go SDK 的用户和项目自动化管理工具。

## 特性

- ✅ **纯 Go 实现** - 使用官方 GitLab Go SDK，无需外部依赖（如 glab）
- ✅ **类型安全** - 完整的类型定义和编译时检查
- ✅ **高性能** - 直接 HTTP API 调用，无进程调用开销
- ✅ **易于维护** - 结构化错误处理和清晰的代码架构
- ✅ **功能完整** - 支持用户、组、项目的批量创建和清理

## 技术栈

- **Go** 1.23.0+
- **GitLab SDK**: `gitlab.com/gitlab-org/api/client-go` v0.157.0
- **CLI 框架**: `github.com/spf13/cobra` v1.8.0
- **配置解析**: `gopkg.in/yaml.v3` v3.0.1

## 快速开始

### 1. 构建工具

```bash
# 下载依赖并构建
make build

# 或者使用 Go 命令
go build -o bin/gitlab-cli ./cmd/gitlab-cli
```

### 2. 配置环境变量

```bash
export GITLAB_HOST=https://gitlab.example.com
export GITLAB_TOKEN=glpat-your-token-here
```

**Token 要求：**
- 必须拥有 `api` 和 `sudo` scopes
- Token 所属用户必须是 GitLab 管理员

### 3. 准备配置文件

创建 `users.yaml`:

```yaml
users:
  - username: testuser1
    email: testuser1@example.com
    name: "Test User 1"
    password: "SecurePass123!"
    groups:
      - name: test-group
        path: test-group
        visibility: private
        projects:
          - name: test-project
            path: test-project
            description: "测试项目"
            visibility: private
```

**注意**：每个用户可以创建多个组（`groups` 是数组），每个组下可以有多个项目。

### 4. 创建用户和项目

```bash
./bin/gitlab-cli user create --config users.yaml
```

### 5. 清理用户

```bash
./bin/gitlab-cli user cleanup --config users.yaml
```

## 命令参考

### user create

创建用户、组和项目：

```bash
gitlab-cli-sdk user create [flags]

Flags:
  -f, --config string   配置文件路径 (默认 "../test-users.yaml")
      --host string     GitLab 主机地址
      --token string    GitLab Personal Access Token
  -h, --help           帮助信息
```

### user cleanup

清理（删除）用户：

```bash
gitlab-cli-sdk user cleanup [flags]

Flags:
  -f, --config string   配置文件路径 (默认 "../test-users.yaml")
      --host string     GitLab 主机地址
      --token string    GitLab Personal Access Token
  -h, --help           帮助信息
```

## Makefile 命令

```bash
make help            # 显示所有可用命令
make deps            # 下载 Go 依赖
make build           # 构建当前平台二进制
make build-all       # 构建所有平台二进制
make install         # 安装到 /usr/local/bin
make clean           # 清理构建文件
make test            # 运行测试
make fmt             # 格式化代码
make lint            # 代码检查
make run             # 运行示例
make release         # 创建发布包
```

## 项目结构

```
gitlab-cli-sdk/
├── cmd/
│   └── gitlab-cli/          # 命令行程序入口
│       └── main.go          # main 函数
├── internal/                # 内部包（仅供本项目使用）
│   ├── cli/                 # CLI 命令定义
│   ├── config/              # 配置管理
│   ├── processor/           # 业务逻辑处理
│   └── utils/               # 工具函数
├── pkg/                     # 公共包（可被外部使用）
│   ├── client/              # GitLab API 客户端封装
│   └── types/               # 数据类型定义
├── docs/                    # 文档
│   ├── ARCHITECTURE.md      # 架构文档
│   ├── QUICKSTART.md        # 快速开始
│   └── README.md            # 详细说明（本文件）
├── bin/                     # 编译输出目录
├── go.mod                   # Go 模块依赖
├── go.sum                   # 依赖校验和
├── Makefile                 # 构建脚本
└── README.md                # 项目说明
```

📖 **详细架构说明请参考** [ARCHITECTURE.md](ARCHITECTURE.md)

## 常见问题

### Q: 认证失败怎么办？

**A:** 检查以下几点：
1. Token 是否正确设置
2. Token 是否包含 `api` 和 `sudo` scopes
3. Token 所属用户是否有管理员权限
4. GitLab 主机地址格式是否正确（如 `https://gitlab.example.com`）

### Q: 用户创建失败？

**A:** 常见原因：
1. 邮箱地址已被使用 - 修改邮箱地址
2. 用户名已存在 - 修改用户名或使用 cleanup 清理
3. 密码不符合要求 - 使用更强的密码（大小写+数字+特殊字符）

### Q: 配置文件格式？

**A:** 配置文件使用 YAML 格式，支持为每个用户创建多个组和项目。详见上面的示例。

### Q: 如何在 CI/CD 中使用？

**A:** 示例脚本：

```bash
#!/bin/bash

# 设置环境变量
export GITLAB_HOST="${CI_GITLAB_HOST}"
export GITLAB_TOKEN="${CI_GITLAB_ADMIN_TOKEN}"

# 构建工具
make build

# 创建测试用户
./bin/gitlab-cli user create --config ci-users.yaml

# 运行测试
./run-e2e-tests.sh

# 清理（无论测试是否成功）
./bin/gitlab-cli user cleanup --config ci-users.yaml || true
```

## 开发

### 项目架构

本项目采用标准的 Go 项目布局：
- **cmd/** - 程序入口
- **internal/** - 内部包（不对外暴露）
- **pkg/** - 公共库（可被外部复用）

详细架构说明请参考 [ARCHITECTURE.md](ARCHITECTURE.md)

### 添加新功能

1. 在 `pkg/types/` 中定义数据结构
2. 在 `pkg/client/` 中添加 GitLab API 方法
3. 在 `internal/processor/` 中添加业务逻辑
4. 在 `internal/cli/` 中添加 CLI 命令
5. 更新文档

### 运行测试

```bash
make test
```

### 代码格式化

```bash
make fmt
```

### 添加新命令示例

```go
// internal/cli/cmd.go
func buildXxxCommand(cfg *config.CLIConfig) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "xxx",
        Short: "xxx 命令描述",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runXxx(cfg)
        },
    }
    return cmd
}
```

## 许可证

MIT License

## 相关链接

- [GitLab Go SDK 文档](https://pkg.go.dev/gitlab.com/gitlab-org/api/client-go)
- [GitLab API 文档](https://docs.gitlab.com/ee/api/)
- [Cobra CLI 框架](https://github.com/spf13/cobra)

## 贡献

欢迎提交 Issue 和 Pull Request！
