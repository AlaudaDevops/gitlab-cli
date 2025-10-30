# 模板输出功能

GitLab CLI 支持使用自定义模板来格式化输出结果，让你可以按照自己的需求定制输出格式。

## 快速开始

### 基本用法

```bash
# 使用模板输出
./bin/gitlab-cli user create \
  --host https://your-gitlab.com \
  --token your-token \
  -f config.yaml \
  -o output.yaml \
  -t template.yaml
```

### 参数说明

- `-f, --config`: 输入配置文件（用户、组、项目定义）
- `-o, --output`: 输出文件路径
- `-t, --template`: 模板文件路径（可选）

**注意**:
- 如果不指定 `--template`，将使用默认的 YAML 格式输出
- `--output` 和 `--template` 通常一起使用

## 模板语法

模板使用 Go 的 `text/template` 语法，支持访问以下数据结构：

### 可用数据

```go
.Users[0]
  ├── .Username      // 用户名
  ├── .Email         // 邮箱
  ├── .Name          // 姓名
  ├── .UserID        // 用户 ID
  ├── .Password      // 用户密码
  ├── .Token         // Personal Access Token (可能为空)
  │   ├── .Value     // Token 值
  │   ├── .Scope     // 权限范围数组
  │   └── .ExpiresAt // 过期时间
  ├── .Groups        // 组数组
  │   ├── .Name       // 组名
  │   ├── .Path       // 组路径
  │   ├── .GroupID    // 组 ID
  │   ├── .Visibility // 可见性
  │   └── .Projects   // 组下的项目数组
  │       ├── .Name        // 项目名
  │       ├── .Path        // 项目路径
  │       ├── .ProjectID   // 项目 ID
  │       ├── .Description // 描述
  │       ├── .Visibility  // 可见性
  │       └── .WebURL      // Web URL
  └── .Projects      // 用户级项目数组（不属于任何组）
      ├── .Name        // 项目名
      ├── .Path        // 项目路径
      ├── .ProjectID   // 项目 ID
      ├── .Description // 描述
      ├── .Visibility  // 可见性
      └── .WebURL      // Web URL
```

### 基础语法

#### 1. 变量替换

```yaml
username: {{ .Username }}
email: {{ .Email }}
user_id: {{ .UserID }}
{{- if .Password }}
password: {{ .Password }}
{{- end }}
```

#### 2. 遍历用户列表

```yaml
{{- range .Users }}
user:
  username: {{ .Username }}
  email: {{ .Email }}
{{- end }}
```

#### 3. 条件判断

```yaml
{{- if .Token }}
token:
  value: {{ .Token.Value }}
  expires_at: {{ .Token.ExpiresAt }}
{{- end }}
```

#### 4. 遍历数组

```yaml
scopes:
{{- range .Token.Scope }}
  - {{ . }}
{{- end }}
```

或者内联格式：

```yaml
scope: {{ range $i, $s := .Token.Scope }}{{ if $i }}, {{ end }}{{ $s }}{{ end }}
```

#### 5. 去除空白

使用 `{{-` 和 `-}}` 来去除前后的空白字符：

```yaml
{{- if .Token }}
  # 这一行不会有额外的空行
{{- end }}
```

## 模板示例

### 示例 1: 简单格式

```yaml
# template-simple.yaml
{{- range .Users }}
username: {{ .Username }}
email: {{ .Email }}
user_id: {{ .UserID }}
{{- if .Token }}
token: {{ .Token.Value }}
{{- end }}
{{- end }}
```

### 示例 2: 完整格式（包含在 template-example.yaml）

```yaml
{{- range .Users }}
# ========================================
# 用户: {{ .Username }}
# ========================================

toolchains:
  gitlab:
    endpoint: https://devops-gitlab.alaudatech.net
    host: devops-gitlab.alaudatech.net
    port: 443
    scheme: https
    username: {{ .Username }}
    email: {{ .Email }}
    user_id: {{ .UserID }}
    {{- if .Password }}
    password: {{ .Password }}
    {{- end }}
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
        path: {{ .Path }}
        group_id: {{ .GroupID }}
        visibility: {{ .Visibility }}
        {{- if .Projects }}
        projects:
          {{- range .Projects }}
          - name: {{ .Name }}
            project_id: {{ .ProjectID }}
            web_url: {{ .WebURL }}
          {{- end }}
        {{- end }}
      {{- end }}
    {{- end }}
{{- end }}
```

### 示例 3: 多用户格式

```yaml
# template-multi-user.yaml
users:
{{- range .Users }}
  - username: {{ .Username }}
    email: {{ .Email }}
    user_id: {{ .UserID }}
    {{- if .Token }}
    token: {{ .Token.Value }}
    token_expires: {{ .Token.ExpiresAt }}
    {{- end }}
    groups_count: {{ len .Groups }}
{{- end }}
```

## 使用场景

### 场景 1: 生成 CI/CD 配置

适用于生成 GitLab CI、Jenkins 或其他 CI/CD 工具的配置文件。

### 场景 2: 生成测试配置

为自动化测试生成包含 GitLab 凭证的配置文件。

### 场景 3: 生成文档

自动生成包含项目信息的文档。

### 场景 4: 集成到其他系统

生成符合其他系统要求的配置格式。

## 完整示例

### 输入配置 (test-users.yaml)

```yaml
users:
  - username: tektoncd
    email: tektoncd001@test.example.com
    name: tektoncd-test
    password: "MyStr0ng!Pass2024"
    token:
      scope:
        - api
        - read_user
      expires_at: 2026-01-01
    groups:
      - name: test-group
        path: test-group
        visibility: private
        projects:
          - name: test-project
            path: test-project
            description: 测试项目
            visibility: private
```

### 执行命令

```bash
./bin/gitlab-cli user create \
  --host https://devops-gitlab.alaudatech.net \
  --token glpat-xxx \
  -f test-users.yaml \
  -o result.yaml \
  -t template-example.yaml
```

### 输出结果 (result.yaml)

```yaml
# ========================================
# 用户: tektoncd
# ========================================

toolchains:
  gitlab:
    endpoint: https://devops-gitlab.alaudatech.net
    username: tektoncd
    email: tektoncd001@test.example.com
    user_id: 24
    token:
      value: glpat-5mtyG_ftYFvh7pNKRGXd
      scope: api, read_user
      expires_at: 2026-01-01
    groups:
      - name: test-group
        group_id: 1496
        projects:
          - name: test-project
            project_id: 1434
            web_url: https://devops-gitlab.alaudatech.net/test-group/test-project
```

## 最佳实践

1. **保持模板简洁**: 不要在模板中包含过于复杂的逻辑
2. **使用注释**: 在模板中添加注释说明数据结构
3. **测试模板**: 先用小数据集测试模板是否正确
4. **版本控制**: 将常用模板纳入版本控制
5. **模板复用**: 为不同场景创建多个模板文件

## 故障排查

### 模板解析错误

如果遇到模板解析错误，检查：
- 所有的 `{{` 都有对应的 `}}`
- 变量名是否正确（区分大小写）
- 是否正确使用了 `range`、`if` 等语句

### 空白问题

如果输出有多余的空行：
- 使用 `{{-` 和 `-}}` 去除空白
- 检查模板文件末尾是否有多余空行

### 数据不显示

如果某些数据不显示：
- 检查数据是否实际存在（使用默认格式先确认）
- 使用 `{{ if .Field }}` 检查字段是否存在
- 确认变量路径是否正确

## 参考资源

- [Go text/template 文档](https://pkg.go.dev/text/template)
- [GitLab CLI 项目仓库](https://github.com/your-repo)
