# 发布指南

本项目使用 GitHub Actions 自动化构建和发布流程。

## 自动化工作流

### 1. CI Workflow (`ci.yml`)

**触发条件**：
- 推送代码到 `main` 或 `develop` 分支
- 创建 Pull Request 到 `main` 或 `develop`

**执行任务**：
- ✅ 代码检查 (golangci-lint)
- ✅ 运行测试和生成覆盖率报告
- ✅ 构建验证

### 2. Release Workflow (`release.yml`)

**触发条件**：
- 推送 tag (如 `v0.2.0`)
- 推送代码到 `main` 或 `develop` 分支（仅构建，不发布）

**支持的平台**：
| 操作系统 | 架构 | 文件名 |
|---------|------|--------|
| Linux | amd64 | `gitlab-cli-linux-amd64` |
| Linux | arm64 | `gitlab-cli-linux-arm64` |
| macOS | amd64 (Intel) | `gitlab-cli-darwin-amd64` |
| macOS | arm64 (Apple Silicon) | `gitlab-cli-darwin-arm64` |
| Windows | amd64 | `gitlab-cli-windows-amd64.exe` |

**生成的文件**：
- 二进制文件
- SHA256 校验和文件

## 发布新版本

### 步骤 1: 更新版本号

编辑 `cmd/gitlab-cli/main.go`，更新版本常量：

```go
const Version = "0.3.0"  // 修改为新版本号
```

### 步骤 2: 提交更改

```bash
git add cmd/gitlab-cli/main.go
git commit -m "chore: bump version to v0.3.0"
git push origin main
```

### 步骤 3: 创建并推送 tag

```bash
# 创建 tag
git tag -a v0.3.0 -m "Release v0.3.0

主要更新：
- 新增功能 A
- 修复 Bug B
- 优化性能 C
"

# 推送 tag
git push origin v0.3.0
```

### 步骤 4: 等待自动构建

推送 tag 后，GitHub Actions 会自动：

1. ⏳ 构建所有平台的二进制文件 (约 5-10 分钟)
2. ✅ 生成 SHA256 校验和
3. 📦 创建 GitHub Release
4. 📝 自动生成 Release Notes

### 步骤 5: 验证发布

访问 https://github.com/yhuan123/gitlab-cli/releases 查看发布的版本。

## 版本号规范

遵循 [Semantic Versioning 2.0.0](https://semver.org/):

- **主版本号 (MAJOR)**: 不兼容的 API 变更
- **次版本号 (MINOR)**: 向后兼容的功能新增
- **修订号 (PATCH)**: 向后兼容的问题修复

**示例**：
- `v1.0.0` - 第一个稳定版本
- `v1.1.0` - 新增功能
- `v1.1.1` - Bug 修复
- `v2.0.0` - 破坏性变更

## 手动构建（本地测试）

如果需要本地测试构建：

```bash
# 构建当前平台
make build

# 构建所有平台
make build-all

# 查看构建产物
ls -lh bin/
```

## Release Checklist

发布前检查清单：

- [ ] 所有测试通过
- [ ] 文档已更新
- [ ] CHANGELOG.md 已更新（如果有）
- [ ] 版本号已更新
- [ ] 本地构建测试通过
- [ ] CI 工作流通过
- [ ] Tag 已创建并推送
- [ ] GitHub Release 自动创建成功
- [ ] 二进制文件可下载
- [ ] SHA256 校验和正确

## 常见问题

### Q: 如何取消/删除一个发布？

**A**: 在 GitHub 上：
1. 进入 Releases 页面
2. 点击要删除的 release
3. 点击 "Delete" 按钮
4. 删除对应的 tag:
   ```bash
   git tag -d v0.3.0
   git push origin :refs/tags/v0.3.0
   ```

### Q: 构建失败怎么办？

**A**:
1. 查看 Actions 页面的构建日志
2. 修复问题后重新推送
3. 如果是 tag 触发的，需要删除 tag 后重新创建

### Q: 如何创建预发布版本？

**A**: 使用带有 `-rc` 或 `-beta` 后缀的 tag:
```bash
git tag -a v0.3.0-rc1 -m "Release Candidate 1"
git push origin v0.3.0-rc1
```

### Q: 如何手动上传额外的文件到 Release？

**A**:
1. 等待自动 Release 创建完成
2. 进入 Release 页面点击 "Edit"
3. 拖拽文件到 "Attach binaries" 区域
4. 保存更改

## 相关链接

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [GitHub Releases 文档](https://docs.github.com/en/repositories/releasing-projects-on-github)
- [Semantic Versioning](https://semver.org/)
