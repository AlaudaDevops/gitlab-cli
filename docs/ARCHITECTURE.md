# GitLab CLI SDK 代码架构

## 项目概述

本项目采用标准的 Go 项目布局，遵循 Go 社区最佳实践，实现**模块化设计**、**单一职责原则**和**高内聚、低耦合**。

## 目录结构

```
gitlab-cli-sdk/
├── cmd/
│   └── gitlab-cli/              # 命令行程序入口
│       └── main.go              # main 函数 (19 行)
├── internal/                    # 内部包（仅供本项目使用）
│   ├── cli/
│   │   └── cmd.go              # CLI 命令定义和编排 (181 行)
│   ├── config/
│   │   └── config.go           # 配置加载和凭证管理 (49 行)
│   ├── processor/
│   │   └── processor.go        # 业务逻辑处理器 (329 行)
│   └── utils/
│       └── utils.go            # 工具函数 (8 行)
├── pkg/                         # 公共包（可被外部项目使用）
│   ├── client/
│   │   └── client.go           # GitLab API 客户端封装 (187 行)
│   └── types/
│       └── types.go            # 数据结构定义 (32 行)
├── docs/                        # 文档
│   ├── ARCHITECTURE.md          # 架构文档（本文件）
│   ├── QUICKSTART.md            # 快速开始指南
│   └── README.md                # 详细使用说明
├── bin/                         # 编译输出目录
├── go.mod                       # Go 模块定义
├── go.sum                       # Go 依赖锁定
├── Makefile                     # 构建脚本
└── README.md                    # 项目说明
```

## 架构设计

### 分层架构

项目采用经典的分层架构，从上到下依次为：

```
┌─────────────────────────────────────┐
│   cmd/gitlab-cli (入口层)            │  ← main 函数
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   internal/cli (命令层)              │  ← Cobra 命令定义
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   internal/processor (业务逻辑层)    │  ← 核心业务逻辑
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   pkg/client (API 客户端层)          │  ← GitLab API 封装
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   GitLab REST API                   │  ← 外部服务
└─────────────────────────────────────┘
```

### 模块职责

#### 1. cmd/gitlab-cli - 程序入口层

**文件**: `cmd/gitlab-cli/main.go`

**职责**: 应用程序启动入口

```go
package main

import (
    "gitlab-cli-sdk/internal/cli"
    "gitlab-cli-sdk/internal/config"
)

func main() {
    cfg := &config.CLIConfig{}
    rootCmd := cli.BuildRootCommand(cfg, Version)
    rootCmd.Execute()
}
```

**特点**:
- 极简设计，只包含启动逻辑
- 定义应用版本号
- 遵循 Go 标准项目布局的 `cmd/` 目录约定

---

#### 2. internal/cli - 命令定义层

**文件**: `internal/cli/cmd.go`

**职责**: CLI 命令构建和业务流程编排

**主要函数**:
- `BuildRootCommand()` - 构建根命令
- `buildUserCommand()` - 构建用户管理命令组
- `buildUserCreateCommand()` - 用户创建命令
- `buildUserCleanupCommand()` - 用户清理命令
- `runUserCreate()` - 执行用户创建流程
- `runUserCleanup()` - 执行用户清理流程
- `initializeClient()` - 初始化 GitLab 客户端

**设计特点**:
- 使用构建器模式组织命令树
- 命令定义与业务逻辑分离
- 流程编排清晰，易于理解
- 将具体业务逻辑委托给 `processor`

---

#### 3. internal/config - 配置管理层

**文件**: `internal/config/config.go`

**职责**: 配置加载和凭证管理

**主要类型和函数**:
```go
type CLIConfig struct {
    ConfigFile  string
    GitLabHost  string
    GitLabToken string
}

func LoadGitLabCredentials(cfg *CLIConfig) error
func Load(configFile string) (*types.UserConfig, error)
```

**设计特点**:
- 支持环境变量和命令行参数
- 环境变量优先级处理
- YAML 配置文件解析
- 统一的错误处理

---

#### 4. internal/processor - 业务逻辑层

**文件**: `internal/processor/processor.go`

**职责**: 封装资源操作的核心业务逻辑

**主要类型**:
```go
type ResourceProcessor struct {
    Client *client.GitLabClient
}
```

**创建流程方法**:
- `ProcessUserCreation()` - 用户创建总协调
- `ensureUser()` - 确保用户存在（幂等操作）
- `createGroups()` - 批量创建组
- `ensureGroup()` - 确保组存在（幂等操作）
- `createProjects()` - 批量创建项目

**清理流程方法**:
- `ProcessUserCleanup()` - 用户清理总协调
- `deleteConfiguredGroups()` - 删除配置的组
- `deleteProjects()` - 删除项目
- `verifyGroupsDeletion()` - 验证组删除
- `deleteUserOwnedGroups()` - 删除用户拥有的其他组
- `verifyUserGroupsDeletion()` - 验证用户组删除
- `deleteUser()` - 删除用户并验证

**设计特点**:
- 使用依赖注入（Client 字段）
- 每个方法职责单一，不超过 50 行
- 完善的日志输出
- 幂等性设计（ensure* 方法）
- 异步验证机制

---

#### 5. internal/utils - 工具函数层

**文件**: `internal/utils/utils.go`

**职责**: 通用工具函数

```go
func GetVisibility(v string) string
```

**设计特点**:
- 无状态纯函数
- 可独立测试
- 易于复用

---

#### 6. pkg/client - API 客户端层

**文件**: `pkg/client/client.go`

**职责**: GitLab API 客户端封装

**主要类型**:
```go
type GitLabClient struct {
    client *gitlab.Client
}
```

**API 方法分类**:

**用户操作**:
- `NewGitLabClient()` - 创建客户端
- `CheckAuth()` - 检查认证和权限
- `GetUser()` - 获取用户
- `CreateUser()` - 创建用户
- `DeleteUser()` - 删除用户

**组操作**:
- `GetGroup()` - 获取组
- `CreateGroup()` - 创建组
- `DeleteGroup()` - 删除组
- `ListUserGroups()` - 列出用户拥有的组

**项目操作**:
- `GetProject()` - 获取项目
- `CreateProject()` - 创建项目
- `DeleteProject()` - 删除项目

**设计特点**:
- 封装官方 GitLab SDK
- 统一的错误处理
- 404 状态码特殊处理
- 类型安全
- **作为 `pkg/` 包，可被外部项目复用**

---

#### 7. pkg/types - 数据定义层

**文件**: `pkg/types/types.go`

**职责**: 数据结构定义

```go
type UserConfig struct {
    Users []UserSpec `yaml:"users"`
}

type UserSpec struct {
    Username string      `yaml:"username"`
    Email    string      `yaml:"email"`
    Name     string      `yaml:"name"`
    Password string      `yaml:"password"`
    Groups   []GroupSpec `yaml:"groups"`
}

type GroupSpec struct {
    Name       string        `yaml:"name"`
    Path       string        `yaml:"path"`
    Visibility string        `yaml:"visibility"`
    Projects   []ProjectSpec `yaml:"projects"`
}

type ProjectSpec struct {
    Name        string `yaml:"name"`
    Path        string `yaml:"path"`
    Description string `yaml:"description"`
    Visibility  string `yaml:"visibility"`
}
```

**设计特点**:
- 清晰的数据结构
- YAML 映射标签
- **作为 `pkg/` 包，可被外部项目复用**

---

## 包的可见性设计

### internal/ vs pkg/

本项目严格遵循 Go 的包可见性约定：

#### internal/ - 内部包
- **特点**: 仅供本项目内部使用，不能被外部项目导入
- **包含**: `cli`, `config`, `processor`, `utils`
- **原因**: 这些包包含项目特定的业务逻辑和配置，不应被外部复用

#### pkg/ - 公共包
- **特点**: 可以被其他项目导入和复用
- **包含**: `client`, `types`
- **原因**:
  - `client` 提供了通用的 GitLab API 封装，其他项目可能需要类似功能
  - `types` 定义了 GitLab 资源的数据结构，可供其他工具使用

## 依赖关系图

```
cmd/gitlab-cli/main.go
  ├── internal/cli
  │    ├── internal/config
  │    │    └── pkg/types
  │    └── internal/processor
  │         ├── pkg/client
  │         ├── pkg/types
  │         └── internal/utils
  └── internal/config
       └── pkg/types

pkg/client (独立，可被外部使用)
pkg/types (独立，可被外部使用)
```

## 设计原则

### 1. 单一职责原则（SRP）
- 每个包只负责一个功能域
- 每个文件只包含相关的功能
- 每个函数只做一件事

### 2. 依赖倒置原则（DIP）
- `processor` 依赖于 `client` 接口，而不是具体实现
- 使用依赖注入便于测试

### 3. 开闭原则（OCP）
- 新增命令：在 `cli` 包中添加新的命令构建函数
- 新增资源类型：在 `processor` 中添加新的处理方法
- 不需要修改现有代码

### 4. 接口隔离原则（ISP）
- Client 提供明确的 API 方法
- 每个方法功能单一

### 5. 高内聚、低耦合
- 相关功能集中在同一包中
- 包之间通过清晰的接口交互
- 单向依赖关系

## 代码指标对比

| 指标 | 重构前 | 重构后 | 改进 |
|------|--------|--------|------|
| 文件数 | 3 个 | 7 个包 + 7 个文件 | 模块化 ✅ |
| main.go | 535 行 | 19 行 | ↓ 96% |
| 最大文件 | 535 行 | 329 行 | ↓ 39% |
| 最大函数 | ~200 行 | ~50 行 | ↓ 75% |
| 包可见性 | 无 | internal/pkg 分离 | ✅ |
| 可复用性 | 低 | 高 (pkg/) | ✅ |

## 扩展指南

### 添加新命令

1. 在 `internal/cli/cmd.go` 中添加命令构建函数：
```go
func buildXxxCommand(cfg *config.CLIConfig) *cobra.Command {
    // 命令定义
}
```

2. 添加执行函数：
```go
func runXxx(cfg *config.CLIConfig) error {
    // 业务流程编排
}
```

3. 在 `internal/processor/processor.go` 中添加业务逻辑：
```go
func (p *ResourceProcessor) ProcessXxx(...) error {
    // 具体实现
}
```

### 添加新的资源类型

1. 在 `pkg/types/types.go` 中定义数据结构
2. 在 `pkg/client/client.go` 中添加 API 方法
3. 在 `internal/processor/processor.go` 中添加处理逻辑
4. 在 `internal/cli/cmd.go` 中添加命令

### 编写测试

```go
// internal/processor/processor_test.go
package processor_test

import (
    "testing"
    "gitlab-cli-sdk/internal/processor"
    "gitlab-cli-sdk/pkg/client"
)

type mockClient struct {
    client.GitLabClient
}

func TestProcessUserCreation(t *testing.T) {
    mockClient := &mockClient{}
    proc := &processor.ResourceProcessor{Client: mockClient}
    // 测试逻辑
}
```

## 最佳实践

### 命名规范
- **包名**: 小写单数（`config`, `client`, 不是 `configs`, `clients`）
- **导出类型**: 大驼峰（`GitLabClient`, `UserConfig`）
- **导出函数**: 大驼峰（`NewGitLabClient`, `LoadConfig`）
- **私有函数**: 小驼峰（`ensureUser`, `deleteProjects`）
- **接口**: 以 `er` 结尾（如 `Reader`, `Writer`）

### 注释规范
- 每个导出的类型、函数都有注释
- 注释以类型/函数名开头
- 说明功能，不说明实现
- 使用中文提高团队可读性

### 错误处理
- 使用 `fmt.Errorf` 包装错误并添加上下文
- 在调用链中逐层添加信息
- 不吞噬错误
- 记录关键操作的错误日志

### 日志输出
- 使用 `log.Printf` 输出结构化日志
- 使用 emoji 增强可读性（✓, ⚠）
- 记录操作进度和结果

## 性能考虑

- 使用 GitLab SDK 的批量 API（如 `ListGroups`）
- 实现资源验证的重试机制
- 合理的超时和等待时间

## 安全考虑

- Token 通过环境变量传递
- 不在日志中输出敏感信息
- 使用 sudo 模式确保权限隔离

## 参考资料

- [Go 项目标准布局](https://github.com/golang-standards/project-layout)
- 《Code Complete》第 7 章 - 高质量的子程序
- 《Clean Code》第 3 章 - 函数
- 《重构：改善既有代码的设计》
- [Effective Go](https://golang.org/doc/effective_go.html)
