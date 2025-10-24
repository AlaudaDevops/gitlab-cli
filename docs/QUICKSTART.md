# 快速入门指南

本指南帮助你快速开始使用 GitLab CLI SDK。

## 5 分钟快速开始

### 1. 构建工具

```bash
# 下载 Go 依赖并构建
make build
```

### 2. 配置 GitLab Token

```bash
# 设置环境变量
export GITLAB_HOST=https://gitlab.example.com
export GITLAB_TOKEN=glpat-your-token-here
```

**Token 要求：**
- 必须包含 `api` 和 `sudo` scopes
- Token 所属用户必须是 GitLab 管理员

### 3. 创建配置文件

创建 `my-users.yaml`:
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

**注意**：配置文件支持为每个用户创建多个组，每个组下可以有多个项目。

### 4. 创建用户

```bash
./bin/gitlab-cli user create --config my-users.yaml
```

### 5. 清理用户

测试完成后清理：
```bash
./bin/gitlab-cli user cleanup --config my-users.yaml
```

## 常见问题

### Q: 认证失败

**A:** 检查以下几点：
1. Token 是否正确设置
2. Token 是否包含 `api` 和 `sudo` scopes
3. Token 所属用户是否有管理员权限
4. GitLab 主机地址是否正确

### Q: 用户创建失败

**A:** 常见原因：
1. 邮箱地址已被使用 - 修改邮箱地址
2. 用户名已存在 - 修改用户名
3. 密码不符合要求 - 使用更强的密码

## 完整工作流程示例

### 在 CI/CD 中使用

```bash
#!/bin/bash

# 1. 设置环境变量
export GITLAB_HOST="${CI_GITLAB_HOST}"
export GITLAB_TOKEN="${CI_GITLAB_ADMIN_TOKEN}"

# 2. 构建工具
make build

# 3. 创建测试用户
./bin/gitlab-cli user create --config ci-users.yaml

# 4. 运行测试
./run-e2e-tests.sh

# 5. 清理（无论测试是否成功）
./bin/gitlab-cli user cleanup --config ci-users.yaml || true
```

### 本地开发测试

```bash
#!/bin/bash

# 开发环境配置
export GITLAB_HOST=https://dev-gitlab.company.com
export GITLAB_TOKEN=$(cat ~/.gitlab-admin-token)

# 构建最新版本
make build

# 创建临时用户进行测试
./bin/gitlab-cli user create --config dev-users.yaml

echo "测试用户已创建，可以开始测试了"
echo "测试完成后运行: ./bin/gitlab-cli user cleanup --config dev-users.yaml"
```

## Makefile 速查表

```bash
# 依赖管理
make deps            # 下载依赖

# 构建
make build           # 构建当前平台
make build-all       # 构建所有平台
make clean           # 清理

# 开发
make fmt             # 格式化代码
make test            # 运行测试
make run             # 运行示例

# 帮助
make help            # 查看所有命令
```

## 下一步

- 阅读 [README.md](README.md) 了解更多功能
- 查看 [ARCHITECTURE.md](ARCHITECTURE.md) 了解代码架构
- 参考根目录的 `test-users.yaml` 作为配置文件示例

## 获取帮助

```bash
# 查看命令帮助
./bin/gitlab-cli --help
./bin/gitlab-cli user --help
./bin/gitlab-cli user create --help
./bin/gitlab-cli user cleanup --help
```

开始使用吧！
