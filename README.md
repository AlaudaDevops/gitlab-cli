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
```

## 📖 配置文件示例

### 基本配置

```yaml
# test-users.yaml
users:
  - username: tektoncd
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

# 方式 3: 注释掉 expires_at（推荐用于测试）
token:
  scope:
    - api
    - read_user
  # expires_at: 2026-01-01  # 注释掉则使用默认值
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
# 使用 Go template 语法
{{- range .Users }}
toolchains:
  gitlab:
    endpoint: https://your-gitlab.com
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
      {{- range .Groups }}
      - name: {{ .Name }}
        group_id: {{ .GroupID }}
      {{- end }}
    {{- end }}
{{- end }}
```

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

## 🔧 命令行参数

### user create

创建用户、组、项目和 Token。

```bash
./bin/gitlab-cli user create [flags]
```

**参数**:

- `-f, --config` - 配置文件路径 (默认: test-users.yaml)
- `--host` - GitLab 主机地址
- `--token` - GitLab Personal Access Token
- `-o, --output` - 输出结果到 YAML 文件
- `-t, --template` - 使用模板文件格式化输出

**示例**:

```bash
# 基本用法
./bin/gitlab-cli user create -f config.yaml

# 输出到文件
./bin/gitlab-cli user create -f config.yaml -o output.yaml

# 使用模板输出
./bin/gitlab-cli user create -f config.yaml -o output.yaml -t template.yaml

# 指定 GitLab 地址和 Token
./bin/gitlab-cli user create \
  --host https://gitlab.example.com \
  --token glpat-xxxxxxxxxxxxxxxxxxxx \
  -f config.yaml \
  -o output.yaml
```

### user cleanup

清理配置文件中定义的用户及其所有资源。

```bash
./bin/gitlab-cli user cleanup [flags]
```

**参数**:

- `-f, --config` - 配置文件路径
- `--host` - GitLab 主机地址
- `--token` - GitLab Personal Access Token

**示例**:

```bash
./bin/gitlab-cli user cleanup -f config.yaml
```

## 🛠️ 开发

### 构建命令

```bash
# 下载依赖
make deps

# 格式化代码
make fmt

# 运行代码检查
make lint

# 运行测试
make test

# 构建当前平台
make build

# 构建所有平台
make build-all

# 创建发布包
make release

# 清理构建文件
make clean

# 查看帮助
make help
```

### 版本管理

版本号通过构建时注入：

```bash
go build -ldflags "-X main.Version=1.0.0" -o bin/gitlab-cli ./cmd/gitlab-cli
```

查看版本：

```bash
./bin/gitlab-cli --version
```

## 🎯 使用场景

### 场景 1: CI/CD 测试环境准备

为 CI/CD 流程自动创建测试用户、组和项目：

```yaml
users:
  - username: ci-test-user
    email: ci-test@example.com
    name: CI Test User
    password: "SecurePassword123!"
    token:
      scope:
        - api
        - read_repository
        - write_repository
      expires_at: 2025-12-31
    groups:
      - name: ci-test-group
        visibility: private
        projects:
          - name: test-project
            visibility: private
```

### 场景 2: 批量用户管理

为团队成员批量创建 GitLab 账户和项目空间：

```yaml
users:
  - username: developer1
    email: dev1@example.com
    token:
      scope: [api, read_user]
    groups:
      - name: dev1-workspace

  - username: developer2
    email: dev2@example.com
    token:
      scope: [api, read_user]
    groups:
      - name: dev2-workspace
```

### 场景 3: 生成自定义配置

使用模板为其他系统生成配置文件：

```bash
# 生成符合特定格式的配置
./bin/gitlab-cli user create \
  -f users.yaml \
  -o k8s-config.yaml \
  -t k8s-template.yaml
```

## ⚠️ 注意事项

1. **Token 安全**
   - Personal Access Token 只在创建时显示一次
   - 请妥善保存输出文件中的 Token 值
   - 不要将包含 Token 的输出文件提交到版本控制

2. **权限要求**
   - 需要 GitLab 管理员权限的 Token
   - Token 必须包含 `api` 和 `sudo` 权限范围

3. **清理操作**
   - `cleanup` 命令会删除用户及其所有关联资源
   - 删除操作不可逆，请谨慎使用
   - 建议在生产环境使用前先在测试环境验证

4. **过期时间**
   - Token 的过期时间默认为第2天
   - 建议根据实际需要设置合适的过期时间
   - 已过期的 Token 无法使用，需要重新创建

## 🐛 故障排查

### 认证失败

```
authentication failed: 401 Unauthorized
```

**解决方案**: 检查 Token 是否有效，是否有管理员权限。

### 权限不足

```
current user is not admin
```

**解决方案**: 确保使用的 Token 属于管理员账户。

### Token Scope 无效

```
scopes does not have a valid value
```

**解决方案**: 检查配置文件中的 scope 值是否为 GitLab 支持的权限范围。

### 模板渲染错误

```
parse template: template: output:1: unexpected "}"
```

**解决方案**: 检查模板语法，确保所有的 `{{` 都有对应的 `}}`。

## 📝 更新日志

### v0.2.0 (Latest)

**新功能**:
- ✨ 添加 Personal Access Token 自动创建功能
- ✨ 支持自定义 Token 权限范围和过期时间
- ✨ Token 过期时间默认值（第2天）
- ✨ 输出结果到 YAML 文件
- ✨ 自定义模板输出功能
- ✨ 完整的输出数据结构（包括 Token 值、项目 URL 等）

**改进**:
- 📝 完善文档和使用示例
- 🔧 优化错误处理和日志输出

### v0.1.0

**初始版本**:
- ✅ 基于 GitLab Go SDK 的基础实现
- ✅ 用户、组、项目的创建和管理
- ✅ 批量操作支持
- ✅ 清理功能

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

MIT License

## 🙏 致谢

- [GitLab Go SDK](https://gitlab.com/gitlab-org/api/client-go) - 官方 GitLab API 客户端
- [Cobra](https://github.com/spf13/cobra) - CLI 框架
- [YAML v3](https://github.com/go-yaml/yaml) - YAML 解析库
