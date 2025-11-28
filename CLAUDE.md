# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a GitLab CLI tool built with the official GitLab Go SDK for automating user, group, and project management. The tool uses a layered architecture with clear separation between CLI commands, business logic, and API client operations.

## Build and Development Commands

### Building
```bash
# Build for current platform
make build

# Build for all platforms (Linux, macOS, ARM64/AMD64)
make build-all

# Install to /usr/local/bin
make install
```

### Testing and Quality
```bash
# Run tests
make test
# Or: go test -v -race -coverprofile=coverage.out ./...

# Format code
make fmt
# Or: go fmt ./...

# Run linter (if golangci-lint is installed)
make lint
```

### Running
```bash
# Show help
./bin/gitlab-cli --help

# Create users/groups/projects from config
./bin/gitlab-cli user create -f config.yaml --host https://gitlab.example.com --token <token>

# With output file
./bin/gitlab-cli user create -f config.yaml -o output.yaml

# With custom template
./bin/gitlab-cli user create -f config.yaml -o output.yaml -t template.yaml

# Cleanup users (default: only delete users created 2+ days ago)
./bin/gitlab-cli user cleanup -f config.yaml
./bin/gitlab-cli user cleanup -f config.yaml --days-old 7   # Delete users 7+ days old
./bin/gitlab-cli user cleanup -f config.yaml --days-old 0   # Delete all (ignore date)

# Delete specific users
./bin/gitlab-cli user delete --username user1,user2

# List users by prefix
./bin/gitlab-cli user list --prefix tektoncd

# Delete users by prefix (default: only delete users created 2+ days ago)
./bin/gitlab-cli user delete-by-prefix --prefix tektoncd --dry-run           # Preview
./bin/gitlab-cli user delete-by-prefix --prefix tektoncd --days-old 7        # Delete 7+ days old
./bin/gitlab-cli user delete-by-prefix --prefix tektoncd --days-old 0        # Delete all
```

## Architecture

### Layered Structure

The codebase follows a strict layered architecture (top to bottom):

1. **Entry Layer** (`cmd/gitlab-cli/main.go`): Application entry point, minimal logic
2. **Command Layer** (`internal/cli/cmd.go`): Cobra command definitions and orchestration
3. **Business Logic Layer** (`internal/processor/processor.go`): Core resource processing logic
4. **API Client Layer** (`pkg/client/client.go`): GitLab API abstraction
5. **External Service**: GitLab REST API

### Package Visibility

- **`internal/`**: Private to this project only
  - `internal/cli`: CLI command builders and flow orchestration
  - `internal/config`: Configuration loading and credential management
  - `internal/processor`: Business logic for creating/deleting resources
  - `internal/template`: Go template rendering for custom output
  - `internal/utils`: Utility functions (timestamp generation, visibility conversion)

- **`pkg/`**: Public packages that can be imported by external projects
  - `pkg/client`: GitLab API client wrapper (built on official SDK)
  - `pkg/types`: Data structure definitions (UserSpec, GroupSpec, ProjectSpec, etc.)

### Key Design Patterns

1. **Dependency Injection**: `ResourceProcessor` receives `GitLabClient` as a field
2. **Idempotent Operations**: `ensure*` methods check existence before creating
3. **Single Responsibility**: Each function does one thing, typically under 50 lines
4. **Builder Pattern**: Commands are built using `build*Command()` functions

### Data Flow

1. User runs CLI command → `cmd/gitlab-cli/main.go`
2. Cobra routes to command handler → `internal/cli/cmd.go`
3. Command handler initializes client and loads config
4. Business logic processes resources → `internal/processor/processor.go`
5. API client makes GitLab API calls → `pkg/client/client.go`
6. Results are collected and optionally saved to YAML/template output

## Important Implementation Details

### Naming Modes

The tool supports two naming modes for resources:

- **`prefix` mode (default)**: Appends timestamp to username/email/group/project paths
  - Example: `tektoncd` → `tektoncd-20251030150000`
  - Used for test environments and creating multiple similar resources
  - **Important**: Cleanup must use the output file from creation (not config file)

- **`name` mode**: Uses names directly from config file without modification
  - Example: `test-user-001` → `test-user-001`
  - Used for production or fixed-name resources
  - Can use config file directly for cleanup

### Token Management

- Personal Access Tokens are created with configurable scopes
- If `expires_at` is not specified in config, defaults to 2 days from today
- Token values are included in output files for automation purposes
- The admin token (used for API calls) requires `api` + `sudo` scopes

### Resource Hierarchy

Resources are created in this order:
1. User
2. Personal Access Token (if configured)
3. Groups (with nested projects)
4. User-level projects (projects not belonging to any group)

Deletion happens in reverse order to avoid dependency issues.

### User Age Filtering (Cleanup/Delete Safety Feature)

Both `user cleanup` and `user delete-by-prefix` commands include a safety mechanism to prevent accidental deletion of recently created users:

- **Default behavior**: Only delete users created 2 or more days ago
- **Configurable**: Use `--days-old` flag to adjust the threshold
  - `--days-old 7`: Only delete users 7+ days old
  - `--days-old 0`: Disable age filtering, delete all matching users
- **Implementation**: Compares user's `CreatedAt` field with current time
- **Logging**: Shows creation date and days since creation for each user

This prevents accidental deletion of users created in recent test runs or CI/CD pipelines.

### Error Handling

- 404 errors from GitLab API are handled specially (resource doesn't exist)
- Each operation includes comprehensive logging with emoji indicators (✓, ⚠)
- Errors are wrapped with context using `fmt.Errorf`
- Failed operations log warnings but don't always stop the entire batch

### Environment Variables

- `GITLAB_URL`: GitLab instance URL (can be overridden by `--host`)
- `GITLAB_TOKEN`: Personal Access Token (can be overridden by `--token`)
- Command-line flags take precedence over environment variables

## Configuration File Structure

YAML files define users with nested groups and projects:
- `nameMode`: "prefix" or "name" (default: "prefix")
- `token`: Optional PAT configuration with scopes and expiration
- `groups[]`: Array of groups, each with their own projects
- `projects[]`: User-level projects (not under any group)

See `examples/user.yaml` for a complete example.

## Template System

The tool supports Go template rendering for custom output formats:
- Template file specified with `-t` flag
- Has access to `OutputConfig` struct with all created resource data
- Dynamic fields available: `$.Endpoint`, `$.Host`, `$.Scheme`, `$.Port`
- User data: `.Username`, `.UserID`, `.Token.Value`, `.Groups`, etc.

See `template-example.yaml` for template syntax.

## CI/CD

GitHub Actions workflow (`.github/workflows/ci.yml`) runs:
- Lint (golangci-lint)
- Test (with race detection and coverage)
- Build (verifies binary creation)

All checks must pass on Go 1.23.

## Adding New Features

### To add a new command:
1. Create `build*Command()` function in `internal/cli/cmd.go`
2. Create `run*()` orchestration function
3. Add business logic method to `ResourceProcessor` in `internal/processor/processor.go`
4. If needed, add new API methods to `pkg/client/client.go`

### To add a new resource type:
1. Define data structures in `pkg/types/types.go`
2. Add API methods in `pkg/client/client.go`
3. Add processing logic in `internal/processor/processor.go`
4. Wire up commands in `internal/cli/cmd.go`

## Code Style

- Package names: lowercase singular (`config`, not `configs`)
- Exported types/functions: PascalCase (`GitLabClient`, `NewGitLabClient`)
- Private functions: camelCase (`ensureUser`, `deleteProjects`)
- Chinese comments are used throughout for team readability
- Log messages use structured format with emoji indicators
