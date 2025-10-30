# GitLab CLI

GitLab 用户和项目自动化管理工具，基于官方 GitLab Go SDK 开发。

## ✨ 特性

- ✅ **官方 SDK**: 使用 GitLab 官方 Go SDK (`gitlab.com/gitlab-org/api/client-go`)
- ✅ **纯 Go 实现**: 无需外部依赖，类型安全的 API 调用
- ✅ **批量管理**: 支持批量创建和管理 GitLab 用户、组和项目
- ✅ **Token 自动创建**: 为用户自动创建 Personal Access Token，支持自定义权限和过期时间
- ✅ **智能默认值**: Token 过期时间默认为第2天（从当天算起）
- ✅ **灵活输出**: 支持默认 YAML 格式和自定义 Go Template 模板输出
- ✅ **完整结果**: 输出包含 Token 值、用户 ID、组 ID、项目 ID、Web URL 等完整信息
- ✅ **模块化设计**: 易于维护和扩展

## 🚀 快速开始

### 前置要求

- Go 1.23.0 或更高版本
- GitLab 管理员权限的 Personal Access Token (需要 `api` + `sudo` scopes)

### 安装

```bash
# 克隆仓库
git clone <repository-url>
cd gitlab-cli

# 构建
make build

# 或者直接安装到系统
make install
```

### 基本用法

```bash
# 设置环境变量（可选）
export GITLAB_HOST=https://your-gitlab-instance.com
export GITLAB_TOKEN=your-personal-access-token

# 创建用户、组和项目
./bin/gitlab-cli user create \
  --host https://your-gitlab.com \
  --token your-token \
  -f config.yaml

# 输出结果到文件
./bin/gitlab-cli user create \
  --host https://your-gitlab.com \
  --token your-token \
  -f config.yaml \
  -o output.yaml

# 使用自定义模板输出
./bin/gitlab-cli user create \
  --host https://your-gitlab.com \
  --token your-token \
  -f config.yaml \
  -o output.yaml \
  -t template.yaml

# 清理用户及其资源
./bin/gitlab-cli user cleanup \
  --host https://your-gitlab.com \
  --token your-token \
  -f config.yaml

# ⚠️ 注意：prefix 模式下的清理
# 如果使用 nameMode: prefix（添加时间戳），清理时需要使用创建时输出的文件
# 因为实际的用户名、组名、项目名都带有时间戳

# 1. 创建时保存输出文件
./bin/gitlab-cli user create \
  -f config.yaml \
  -o output.yaml

# 2. 清理时使用输出文件
./bin/gitlab-cli user cleanup \
  -f output.yaml
```

## 📖 配置文件示例

### 命名模式说明

配置文件支持两种命名模式：

**1. prefix 模式（默认）**
- 自动在 username、email、group path、project path 后添加时间戳
- 示例：`tektoncd` → `tektoncd-20251030150000`
- 适用场景：测试环境、需要创建多个相似资源
- ⚠️ 清理时必须使用创建时输出的文件

**2. name 模式**
- 不添加时间戳，直接使用配置文件中的名称
- 示例：`test-user-001` → `test-user-001`（不变）
- 适用场景：生产环境、固定名称的资源
- 可直接使用配置文件清理

### 基本配置

```yaml
# test-users.yaml
users:
  # 使用 prefix 模式（默认）
  - nameMode: prefix  # 可选，默认为 prefix
    username: tektoncd
    email: tektoncd001@test.example.com
    name: tektoncd-test
    password: "MyStr0ng!Pass2024"

    # Personal Access Token 配置（可选）
    token:
      scope:
        - api
        - read_user
        - read_repository
        - write_repository
        - read_api
        - create_runner
      # expires_at: 2026-01-01  # 可选，不指定则默认为第2天

    # 组和项目配置
    groups:
      - name: tektoncd-frontend-group
        path: tektoncd-frontend-group
        visibility: private
        projects:
          - name: test-e2e-demo
            path: test-e2e-demo
            description: 测试前端应用
            visibility: private
          - name: test-vue-app
            path: test-vue-app
            description: Vue.js 测试应用
            visibility: private
      - name: tektoncd-backend-group
        path: tektoncd-backend-group
        visibility: private
        projects:
          - name: test-java
            path: test-java-e2e-demo
            description: 测试后端 API
            visibility: public
          - name: test-go-api
            path: test-go-api
            description: Go API 服务
            visibility: private
```

### Token 配置说明

#### 支持的权限范围 (scope)

- `api` - 完整的 API 访问权限
- `read_user` - 读取用户信息
- `read_repository` - 读取仓库
- `write_repository` - 写入仓库
- `read_api` - 只读 API 访问
- `create_runner` - 创建 Runner
- `sudo` - 管理员权限

#### 过期时间

- **指定过期时间**: `expires_at: 2026-01-01` (格式: YYYY-MM-DD)
- **不指定**: 自动设置为第2天过期（从当天算起，即今天 + 2 天）

**示例**:
```yaml
# 方式 1: 指定过期时间
token:
  scope:
    - api
  expires_at: 2026-01-01

# 方式 2: 使用默认过期时间（第2天）
token:
  scope:
    - api
  # 不指定 expires_at，系统自动设为第2天

```

**默认过期时间说明**:
- 如果今天是 2025-10-27，默认过期时间为 2025-10-29
- Token 会在过期时间当天结束时失效
- 日志会显示: `未指定过期时间，使用默认值: 2025-10-29 (第2天)`

## 📤 输出功能

### 默认 YAML 输出

```bash
./bin/gitlab-cli user create -f config.yaml -o output.yaml
```

输出格式：

```yaml
users:
  - username: tektoncd
    email: tektoncd001@test.example.com
    name: tektoncd-test
    user_id: 24
    token:
      value: glpat-TXLgrsMwyVt5obFqkDny
      scope:
        - api
        - read_user
        - read_repository
        - write_repository
        - read_api
        - create_runner
      expires_at: "2025-10-29"
    groups:
      - name: tektoncd-frontend-group
        path: tektoncd-frontend-group
        group_id: 1506
        visibility: private
        projects:
          - name: test-e2e-demo
            path: tektoncd-frontend-group/test-e2e-demo
            project_id: 1434
            description: 测试前端应用
            visibility: private
            web_url: https://devops-gitlab.alaudatech.net/tektoncd-frontend-group/test-e2e-demo
          - name: test-vue-app
            path: tektoncd-frontend-group/test-vue-app
            project_id: 1435
            description: Vue.js 测试应用
            visibility: private
            web_url: https://devops-gitlab.alaudatech.net/tektoncd-frontend-group/test-vue-app
      - name: tektoncd-backend-group
        path: tektoncd-backend-group
        group_id: 1507
        visibility: private
        projects:
          - name: test-java
            path: tektoncd-backend-group/test-java-e2e-demo
            project_id: 1436
            description: 测试后端 API
            visibility: public
            web_url: https://devops-gitlab.alaudatech.net/tektoncd-backend-group/test-java-e2e-demo
          - name: test-go-api
            path: tektoncd-backend-group/test-go-api
            project_id: 1437
            description: Go API 服务
            visibility: private
            web_url: https://devops-gitlab.alaudatech.net/tektoncd-backend-group/test-go-api
```

### 自定义模板输出

项目提供了模板示例文件 **template-example.yaml**，展示了如何使用 Go template 语法自定义输出格式。

使用模板：

```yaml
# 使用 Go template 语法，支持动态渲染 GitLab 服务器信息
{{- range .Users }}
toolchains:
  gitlab:
    # 动态渲染服务器配置（根据 --host 参数自动适配）
    endpoint: {{ $.Endpoint }}
    host: {{ $.Host }}
    scheme: {{ $.Scheme }}
    # 用户信息
    username: {{ .Username }}
    email: {{ .Email }}
    user_id: {{ .UserID }}
    {{- if .Token }}
    token:
      value: {{ .Token.Value }}
      scope: {{ range $i, $s := .Token.Scope }}{{ if $i }}, {{ end }}{{ $s }}{{ end }}
      expires_at: {{ .Token.ExpiresAt }}
    {{- end }}
    {{- if .Groups }}
    groups:
      default: {{ .Username }}
      {{- range .Groups }}
      - name: {{ .Name }}
        group_id: {{ .GroupID }}
      {{- end }}
    {{- end }}
{{- end }}
```

**模板说明：**
- `default: {{ .Username }}` - 指定默认组，新创建的项目将默认使用此用户名作为命名空间

使用模板：

```bash
./bin/gitlab-cli user create -f config.yaml -o output.yaml -t template.yaml
```

详细的模板文档请参考 [模板使用指南](docs/TEMPLATE.md)。

## 📁 项目结构

```
gitlab-cli/
├── cmd/
│   └── gitlab-cli/        # 命令行入口
├── internal/              # 内部包（不对外暴露）
│   ├── cli/               # CLI 命令定义
│   ├── config/            # 配置管理
│   ├── processor/         # 业务逻辑处理
│   ├── template/          # 模板渲染
│   └── utils/             # 工具函数
├── pkg/                   # 公共包（可被外部使用）
│   ├── client/            # GitLab 客户端
│   └── types/             # 数据类型定义
├── docs/                  # 文档
│   ├── ARCHITECTURE.md    # 架构设计
│   ├── QUICKSTART.md      # 快速开始
│   ├── TEMPLATE.md        # 模板使用指南
│   └── README.md          # 详细说明
├── bin/                   # 编译输出
├── template-example.yaml  # 模板示例
└── Makefile               # 构建脚本
```

## 📚 文档

- [快速开始指南](docs/QUICKSTART.md) - 快速入门教程
- [架构设计文档](docs/ARCHITECTURE.md) - 详细的代码架构说明
- [模板使用指南](docs/TEMPLATE.md) - 自定义输出模板
- [详细使用文档](docs/README.md) - 完整功能说明
