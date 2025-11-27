# GitLab CLI

A GitLab user and project automation management tool built on the official GitLab Go SDK.

## ✨ Features

- ✅ **Official SDK**: Built with GitLab's official Go SDK (`gitlab.com/gitlab-org/api/client-go`)
- ✅ **Pure Go Implementation**: No external dependencies, type-safe API calls
- ✅ **Batch Management**: Support for batch creation and management of GitLab users, groups, and projects
- ✅ **Auto Token Creation**: Automatically create Personal Access Tokens for users with customizable scopes and expiration dates
- ✅ **Smart Defaults**: Token expiration defaults to 2 days from today
- ✅ **Flexible Output**: Support for default YAML format and custom Go Template outputs
- ✅ **Complete Results**: Output includes token values, user IDs, group IDs, project IDs, web URLs, and more
- ✅ **Modular Design**: Easy to maintain and extend

## 🚀 Quick Start

### Prerequisites

- Go 1.23.0 or higher
- GitLab Personal Access Token with admin privileges (requires `api` + `sudo` scopes)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd gitlab-cli

# Build
make build

# Or install to system
make install
```

### Basic Usage

```bash
# Set environment variables (optional)
export GITLAB_URL=https://your-gitlab-instance.com
export GITLAB_TOKEN=your-personal-access-token

# Create user, groups, and projects
./bin/gitlab-cli user create \
  --host https://your-gitlab.com \
  --token your-token \
  -f config.yaml

# Output results to file
./bin/gitlab-cli user create \
  --host https://your-gitlab.com \
  --token your-token \
  -f config.yaml \
  -o output.yaml

# Use custom template for output
./bin/gitlab-cli user create \
  --host https://your-gitlab.com \
  --token your-token \
  -f config.yaml \
  -o output.yaml \
  -t template.yaml

# Clean up user and their resources
./bin/gitlab-cli user cleanup \
  --host https://your-gitlab.com \
  --token your-token \
  -f config.yaml

# ⚠️ Note: Cleanup with prefix mode
# When using nameMode: prefix (adds timestamp), cleanup requires the output file from creation
# Because actual usernames, group names, and project names all include timestamps

# 1. Save output file during creation
./bin/gitlab-cli user create \
  -f config.yaml \
  -o output.yaml

# 2. Use output file for cleanup
./bin/gitlab-cli user cleanup \
  -f output.yaml

# Delete user and all their resources (projects and groups) by username
./bin/gitlab-cli user delete \
  --host https://your-gitlab.com \
  --token your-token \
  --username user1

# Delete multiple users (comma-separated)
./bin/gitlab-cli user delete \
  --host https://your-gitlab.com \
  --token your-token \
  --username user1,user2,user3
```

## 📖 Configuration File Examples

### Naming Mode Description

The configuration file supports two naming modes:

**1. prefix mode (default)**
- Automatically appends timestamps to username, email, group path, and project path
- Example: `tektoncd` → `tektoncd-20251030150000`
- Use cases: Test environments, creating multiple similar resources
- ⚠️ Cleanup must use the output file from creation

**2. name mode**
- No timestamp added, uses names directly from configuration file
- Example: `test-user-001` → `test-user-001` (unchanged)
- Use cases: Production environments, fixed-name resources
- Can use configuration file directly for cleanup

### Basic Configuration

```yaml
# test-users.yaml
users:
  # Using prefix mode (default)
  - nameMode: prefix  # Optional, defaults to prefix
    username: tektoncd
    email: tektoncd001@test.example.com
    name: tektoncd-test
    password: "MyStr0ng!Pass2024"

    # Personal Access Token configuration (optional)
    token:
      scope:
        - api
        - read_user
        - read_repository
        - write_repository
        - read_api
        - create_runner
      # expires_at: 2026-01-01  # Optional, defaults to 2 days from today if not specified

    # Group and project configuration
    groups:
      - name: tektoncd-frontend-group
        path: tektoncd-frontend-group
        visibility: private
        projects:
          - name: test-e2e-demo
            path: test-e2e-demo
            description: Test frontend application
            visibility: private
          - name: test-vue-app
            path: test-vue-app
            description: Vue.js test application
            visibility: private
      - name: tektoncd-backend-group
        path: tektoncd-backend-group
        visibility: private
        projects:
          - name: test-java
            path: test-java-e2e-demo
            description: Test backend API
            visibility: public
          - name: test-go-api
            path: test-go-api
            description: Go API service
            visibility: private

    # User-level projects (not under any group, directly under user namespace)
    projects:
      - name: my-personal-project
        path: my-personal-project
        description: Personal project under user namespace
        visibility: private
```

### Token Configuration

#### Supported Scopes

- `api` - Full API access
- `read_user` - Read user information
- `read_repository` - Read repository
- `write_repository` - Write to repository
- `read_api` - Read-only API access
- `create_runner` - Create Runner
- `sudo` - Admin privileges

#### Expiration Time

- **Specify expiration**: `expires_at: 2026-01-01` (format: YYYY-MM-DD)
- **Not specified**: Automatically set to expire in 2 days (from today, i.e., today + 2 days)

**Examples**:
```yaml
# Method 1: Specify expiration time
token:
  scope:
    - api
  expires_at: 2026-01-01

# Method 2: Use default expiration (2 days)
token:
  scope:
    - api
  # Don't specify expires_at, system automatically sets to 2 days
```

**Default Expiration Explanation**:
- If today is 2025-10-27, default expiration is 2025-10-29
- Token expires at the end of the expiration date
- Log will show: `Expiration not specified, using default: 2025-10-29 (2 days)`

## 📤 Output Features

### Default YAML Output

```bash
./bin/gitlab-cli user create -f config.yaml -o output.yaml
```

Output contains all created resource information:
- User info: username, email, name, user_id, password
- Token info: value, scope, expires_at (if token is configured)
- Group info: name, path, group_id, visibility (if groups are configured)
- Project info: name, path, project_id, description, visibility, web_url (including projects under groups and user-level projects)

Output format:

```yaml
users:
  - username: tektoncd
    email: tektoncd001@test.example.com
    name: tektoncd-test
    user_id: 24
    password: MyStr0ng!Pass2024
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
            description: Test frontend application
            visibility: private
            web_url: https://devops-gitlab.alaudatech.net/tektoncd-frontend-group/test-e2e-demo
          - name: test-vue-app
            path: tektoncd-frontend-group/test-vue-app
            project_id: 1435
            description: Vue.js test application
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
            description: Test backend API
            visibility: public
            web_url: https://devops-gitlab.alaudatech.net/tektoncd-backend-group/test-java-e2e-demo
          - name: test-go-api
            path: tektoncd-backend-group/test-go-api
            project_id: 1437
            description: Go API service
            visibility: private
            web_url: https://devops-gitlab.alaudatech.net/tektoncd-backend-group/test-go-api
    projects:
      - name: my-personal-project
        path: tektoncd/my-personal-project
        project_id: 1438
        description: Personal project under user namespace
        visibility: private
        web_url: https://devops-gitlab.alaudatech.net/tektoncd/my-personal-project
```

### Custom Template Output

The project provides a template example file **template-example.yaml** that demonstrates how to use Go template syntax for custom output formats.

Using templates:

```yaml
# Using Go template syntax, supports dynamic rendering of GitLab server information
{{- range .Users }}
toolchains:
  gitlab:
    # Dynamic server configuration (automatically adapts based on --host parameter)
    endpoint: {{ $.Endpoint }}
    host: {{ $.Host }}
    scheme: {{ $.Scheme }}
    # User information
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
      default: {{ .Username }}
      {{- range .Groups }}
      - name: {{ .Name }}
        group_id: {{ .GroupID }}
      {{- end }}
    {{- end }}
{{- end }}
```

**Template Notes:**
- `default: {{ .Username }}` - Specifies the default group, newly created projects will use this username as the namespace by default

Using the template:

```bash
./bin/gitlab-cli user create -f config.yaml -o output.yaml -t template.yaml
```

For detailed template documentation, see [Template Usage Guide](docs/TEMPLATE.md).

## 📁 Project Structure

```
gitlab-cli/
├── cmd/
│   └── gitlab-cli/        # CLI entry point
├── internal/              # Internal packages (not exposed)
│   ├── cli/               # CLI command definitions
│   ├── config/            # Configuration management
│   ├── processor/         # Business logic processing
│   ├── template/          # Template rendering
│   └── utils/             # Utility functions
├── pkg/                   # Public packages (can be used externally)
│   ├── client/            # GitLab client
│   └── types/             # Data type definitions
├── docs/                  # Documentation
│   ├── ARCHITECTURE.md    # Architecture design
│   ├── QUICKSTART.md      # Quick start
│   ├── TEMPLATE.md        # Template usage guide
│   └── README.md          # Detailed description
├── bin/                   # Build output
├── template-example.yaml  # Template example
└── Makefile               # Build script
```

## 📚 Documentation

- [Quick Start Guide](docs/QUICKSTART.md) - Getting started tutorial
- [Architecture Documentation](docs/ARCHITECTURE.md) - Detailed code architecture description
- [Template Usage Guide](docs/TEMPLATE.md) - Custom output templates
- [Detailed Usage Documentation](docs/README.md) - Complete feature description

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
