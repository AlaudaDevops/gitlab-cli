package processor

import (
	"fmt"
	"log"
	"time"

	"gitlab-cli-sdk/internal/utils"
	"gitlab-cli-sdk/pkg/client"
	"gitlab-cli-sdk/pkg/types"
)

// ResourceProcessor 封装资源创建和删除的业务逻辑
type ResourceProcessor struct {
	Client *client.GitLabClient
}

// ========================================
// 用户创建流程
// ========================================

// ProcessUserCreation 处理单个用户的创建流程
func (p *ResourceProcessor) ProcessUserCreation(userSpec types.UserSpec) (*types.UserOutput, error) {
	// 确定 nameMode
	nameMode := userSpec.NameMode
	if nameMode == "" {
		nameMode = "prefix" // 默认为 prefix 模式
	}

	// 根据 nameMode 生成实际的 username 和 email
	var actualUsername, actualEmail string
	if nameMode == "name" {
		// name 模式：直接使用配置文件中的名称
		actualUsername = userSpec.Username
		actualEmail = userSpec.Email
		log.Printf("  使用 name 模式（不添加时间戳）\n")
	} else {
		// prefix 模式：添加时间戳
		actualUsername = utils.GenerateUsernameWithTimestamp(userSpec.Username)
		actualEmail = utils.GenerateEmailWithTimestamp(userSpec.Email)
		log.Printf("  使用 prefix 模式（添加时间戳）\n")
	}

	log.Printf("  用户名: %s\n", actualUsername)
	log.Printf("  邮箱: %s\n", actualEmail)

	output := &types.UserOutput{
		Username: actualUsername,
		Email:    actualEmail,
		Name:     userSpec.Name,
	}

	// 1. 创建或获取用户
	userID, err := p.ensureUser(userSpec, actualUsername, actualEmail)
	if err != nil {
		return nil, err
	}
	output.UserID = userID

	// 2. 创建 Personal Access Token (如果配置了)
	if userSpec.Token != nil {
		log.Printf("  创建 Personal Access Token...\n")
		tokenValue, actualExpiresAt, err := p.createPersonalAccessToken(userID, actualUsername, userSpec.Token)
		if err != nil {
			log.Printf("  ⚠ 创建 Token 失败: %v\n", err)
		} else {
			log.Printf("  ✓ Token 创建成功\n")
			log.Printf("  Token Value: %s\n", tokenValue)

			// 保存 Token 信息到输出（使用实际的过期时间）
			output.Token = &types.TokenOutput{
				Value:     tokenValue,
				Scope:     userSpec.Token.Scope,
				ExpiresAt: actualExpiresAt,
			}
		}
	}

	// 3. 创建组和项目
	if len(userSpec.Groups) > 0 {
		log.Printf("  创建 %d 个组...\n", len(userSpec.Groups))
		groupOutputs, err := p.createGroupsWithOutput(actualUsername, userSpec.Groups, nameMode)
		if err != nil {
			return output, err
		}
		output.Groups = groupOutputs
	}

	// 4. 创建用户级项目（不属于任何组的项目）
	if len(userSpec.Projects) > 0 {
		log.Printf("  创建 %d 个用户级项目...\n", len(userSpec.Projects))
		projectOutputs, err := p.createUserProjectsWithOutput(actualUsername, userSpec.Projects, nameMode)
		if err != nil {
			log.Printf("  ⚠ 创建用户级项目失败: %v\n", err)
		} else {
			output.Projects = projectOutputs
		}
	}

	return output, nil
}

// createPersonalAccessToken 为用户创建 Personal Access Token，返回 token 值和实际使用的过期时间
func (p *ResourceProcessor) createPersonalAccessToken(userID int, username string, tokenSpec *types.TokenSpec) (string, string, error) {
	// 生成 token 名称，格式: username-token-timestamp
	tokenName := fmt.Sprintf("%s-token-%d", username, time.Now().Unix())

	// 设置过期时间：如果未指定，默认为第2天
	expiresAt := tokenSpec.ExpiresAt
	if expiresAt == "" {
		// 计算第2天的日期（格式: YYYY-MM-DD）
		tomorrow := time.Now().AddDate(0, 0, 2)
		expiresAt = tomorrow.Format("2006-01-02")
		log.Printf("    未指定过期时间，使用默认值: %s (第2天)\n", expiresAt)
	}

	// 调用客户端创建 token
	tokenValue, err := p.Client.CreatePersonalAccessToken(
		userID,
		tokenName,
		tokenSpec.Scope,
		expiresAt,
	)
	if err != nil {
		return "", "", err
	}

	return tokenValue, expiresAt, nil
}

// ensureUser 确保用户存在，如果不存在则创建
func (p *ResourceProcessor) ensureUser(userSpec types.UserSpec, actualUsername, actualEmail string) (int, error) {
	existingUser, err := p.Client.GetUser(actualUsername)
	if err != nil {
		log.Printf("  ⚠ 检查用户失败: %v\n", err)
	}

	if existingUser != nil {
		log.Printf("  ⚠ 用户 '%s' 已存在 (ID: %d)\n", actualUsername, existingUser.ID)
		return existingUser.ID, nil
	}

	log.Printf("  创建用户: %s\n", actualUsername)
	user, err := p.Client.CreateUser(actualUsername, actualEmail, userSpec.Name, userSpec.Password)
	if err != nil {
		return 0, fmt.Errorf("创建用户 %s: %w", actualUsername, err)
	}

	log.Printf("  ✓ 用户创建成功 (ID: %d)\n", user.ID)
	return user.ID, nil
}

// createGroupsWithOutput 创建多个组及其项目并返回输出结果
func (p *ResourceProcessor) createGroupsWithOutput(username string, groups []types.GroupSpec, userNameMode string) ([]types.GroupOutput, error) {
	var groupOutputs []types.GroupOutput

	for j, groupSpec := range groups {
		log.Printf("  ------------------------------------------\n")
		log.Printf("  处理组 [%d/%d]: %s\n", j+1, len(groups), groupSpec.Name)

		// 确定组的 nameMode（如果组没有指定，则继承用户的 nameMode）
		groupNameMode := groupSpec.NameMode
		if groupNameMode == "" {
			groupNameMode = userNameMode
		}

		groupID, groupPath, err := p.ensureGroup(username, groupSpec, groupNameMode)
		if err != nil {
			log.Printf("    ⚠ 创建组失败 %s: %v\n", groupSpec.Path, err)
			continue
		}

		groupOutput := types.GroupOutput{
			Name:       groupSpec.Name,
			Path:       groupPath,
			GroupID:    groupID,
			Visibility: groupSpec.Visibility,
		}

		// 创建组下的项目
		if len(groupSpec.Projects) > 0 {
			log.Printf("    创建 %d 个项目...\n", len(groupSpec.Projects))
			projectOutputs, err := p.createProjectsWithOutput(username, groupID, groupPath, groupSpec.Projects, groupNameMode)
			if err != nil {
				log.Printf("    ⚠ 创建项目失败: %v\n", err)
			}
			groupOutput.Projects = projectOutputs
		}

		groupOutputs = append(groupOutputs, groupOutput)
	}
	return groupOutputs, nil
}

// ensureGroup 确保组存在，如果不存在则创建
func (p *ResourceProcessor) ensureGroup(username string, groupSpec types.GroupSpec, nameMode string) (int, string, error) {
	// 根据 nameMode 生成实际的 group path
	var actualGroupPath string
	if nameMode == "name" {
		// name 模式：直接使用配置文件中的名称
		actualGroupPath = groupSpec.Path
		if actualGroupPath == "" {
			actualGroupPath = groupSpec.Name
		}
		log.Printf("    使用 name 模式，组 path: %s\n", actualGroupPath)
	} else {
		// prefix 模式：添加时间戳
		if groupSpec.Path == "" {
			actualGroupPath = utils.GenerateGroupPathWithTimestamp(groupSpec.Name)
		} else {
			actualGroupPath = utils.GenerateGroupPathWithTimestamp(groupSpec.Path)
		}
		log.Printf("    使用 prefix 模式，生成组 path: %s\n", actualGroupPath)
	}

	existingGroup, _ := p.Client.GetGroup(actualGroupPath)

	if existingGroup != nil {
		log.Printf("    ⚠ 组 '%s' 已存在 (ID: %d)\n", existingGroup.Path, existingGroup.ID)
		return existingGroup.ID, existingGroup.Path, nil
	}

	log.Printf("    创建组: %s (path: %s)\n", groupSpec.Name, actualGroupPath)
	group, err := p.Client.CreateGroup(
		username,
		groupSpec.Name,
		actualGroupPath,
		utils.GetVisibility(groupSpec.Visibility),
	)
	if err != nil {
		return 0, "", err
	}

	log.Printf("    ✓ 组创建成功 (ID: %d, Path: %s)\n", group.ID, group.Path)
	return group.ID, group.Path, nil
}

// createUserProjectsWithOutput 创建用户级别的项目（不属于任何组）
func (p *ResourceProcessor) createUserProjectsWithOutput(username string, projects []types.ProjectSpec, userNameMode string) ([]types.ProjectOutput, error) {
	var projectOutputs []types.ProjectOutput

	// 获取用户信息以获取用户的 namespace ID
	user, err := p.Client.GetUser(username)
	if err != nil || user == nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	for _, projSpec := range projects {
		// 确定项目的 nameMode（如果项目没有指定，则继承用户的 nameMode）
		projectNameMode := projSpec.NameMode
		if projectNameMode == "" {
			projectNameMode = userNameMode
		}

		// 根据 nameMode 生成实际的 project path
		var actualProjectPath string
		if projectNameMode == "name" {
			// name 模式：直接使用配置文件中的名称
			actualProjectPath = projSpec.Path
			if actualProjectPath == "" {
				actualProjectPath = projSpec.Name
			}
			log.Printf("    使用 name 模式，项目 path: %s\n", actualProjectPath)
		} else {
			// prefix 模式：添加时间戳
			if projSpec.Path == "" {
				actualProjectPath = utils.GenerateProjectPathWithTimestamp(projSpec.Name)
			} else {
				actualProjectPath = utils.GenerateProjectPathWithTimestamp(projSpec.Path)
			}
			log.Printf("    使用 prefix 模式，生成项目 path: %s\n", actualProjectPath)
		}

		// 用户级项目的 full path 是 username/project-path
		fullPath := fmt.Sprintf("%s/%s", username, actualProjectPath)
		existingProj, _ := p.Client.GetProject(fullPath)

		var projectID int
		var webURL string

		if existingProj != nil {
			log.Printf("    ⚠ 项目 '%s' 已存在 (ID: %d)\n", projSpec.Name, existingProj.ID)
			projectID = existingProj.ID
			webURL = existingProj.WebURL
		} else {
			log.Printf("    创建用户级项目: %s (path: %s)\n", projSpec.Name, actualProjectPath)
			// 用户级项目使用用户的 ID 作为 namespace ID
			project, err := p.Client.CreateProject(
				username,
				user.ID, // 使用用户 ID 作为 namespace ID
				projSpec.Name,
				actualProjectPath,
				projSpec.Description,
				utils.GetVisibility(projSpec.Visibility),
			)
			if err != nil {
				log.Printf("    ⚠ 创建项目失败 %s: %v\n", projSpec.Name, err)
				continue
			}
			log.Printf("    ✓ 项目创建成功 (ID: %d, Path: %s)\n", project.ID, project.PathWithNamespace)
			projectID = project.ID
			webURL = project.WebURL
		}

		projectOutputs = append(projectOutputs, types.ProjectOutput{
			Name:        projSpec.Name,
			Path:        fullPath,
			ProjectID:   projectID,
			Description: projSpec.Description,
			Visibility:  projSpec.Visibility,
			WebURL:      webURL,
		})
	}
	return projectOutputs, nil
}

// createProjectsWithOutput 创建多个项目并返回输出结果
func (p *ResourceProcessor) createProjectsWithOutput(username string, groupID int, groupPath string, projects []types.ProjectSpec, groupNameMode string) ([]types.ProjectOutput, error) {
	var projectOutputs []types.ProjectOutput

	for _, projSpec := range projects {
		// 确定项目的 nameMode（如果项目没有指定，则继承组的 nameMode）
		projectNameMode := projSpec.NameMode
		if projectNameMode == "" {
			projectNameMode = groupNameMode
		}

		// 根据 nameMode 生成实际的 project path
		var actualProjectPath string
		if projectNameMode == "name" {
			// name 模式：直接使用配置文件中的名称
			actualProjectPath = projSpec.Path
			if actualProjectPath == "" {
				actualProjectPath = projSpec.Name
			}
			log.Printf("      使用 name 模式，项目 path: %s\n", actualProjectPath)
		} else {
			// prefix 模式：添加时间戳
			if projSpec.Path == "" {
				actualProjectPath = utils.GenerateProjectPathWithTimestamp(projSpec.Name)
			} else {
				actualProjectPath = utils.GenerateProjectPathWithTimestamp(projSpec.Path)
			}
			log.Printf("      使用 prefix 模式，生成项目 path: %s\n", actualProjectPath)
		}

		fullPath := fmt.Sprintf("%s/%s", groupPath, actualProjectPath)
		existingProj, _ := p.Client.GetProject(fullPath)

		var projectID int
		var webURL string

		if existingProj != nil {
			log.Printf("      ⚠ 项目 '%s' 已存在 (ID: %d)\n", projSpec.Name, existingProj.ID)
			projectID = existingProj.ID
			webURL = existingProj.WebURL
		} else {
			log.Printf("      创建项目: %s (path: %s)\n", projSpec.Name, actualProjectPath)
			project, err := p.Client.CreateProject(
				username,
				groupID,
				projSpec.Name,
				actualProjectPath,
				projSpec.Description,
				utils.GetVisibility(projSpec.Visibility),
			)
			if err != nil {
				log.Printf("      ⚠ 创建项目失败 %s: %v\n", projSpec.Name, err)
				continue
			}
			log.Printf("      ✓ 项目创建成功 (ID: %d, Path: %s)\n", project.ID, project.PathWithNamespace)
			projectID = project.ID
			webURL = project.WebURL
		}

		projectOutputs = append(projectOutputs, types.ProjectOutput{
			Name:        projSpec.Name,
			Path:        fullPath,
			ProjectID:   projectID,
			Description: projSpec.Description,
			Visibility:  projSpec.Visibility,
			WebURL:      webURL,
		})
	}
	return projectOutputs, nil
}

// ========================================
// 用户清理流程
// ========================================

// ProcessUserCleanup 处理单个用户的清理流程
func (p *ResourceProcessor) ProcessUserCleanup(userSpec types.UserSpec) error {
	user, err := p.Client.GetUser(userSpec.Username)
	if err != nil {
		log.Printf("  ⚠ 检查用户失败: %v\n", err)
		return nil
	}

	if user == nil {
		log.Printf("  用户不存在，跳过: %s\n\n", userSpec.Username)
		return nil
	}

	log.Printf("  找到用户 '%s' (ID: %d, 邮箱: %s)\n", user.Username, user.ID, user.Email)

	// 1. 删除用户级项目（不属于任何组的项目）
	if len(userSpec.Projects) > 0 {
		log.Printf("  删除 %d 个用户级项目...\n", len(userSpec.Projects))
		p.deleteUserProjects(userSpec.Username, userSpec.Projects)
	}

	// 2. 删除配置文件中定义的组和项目
	if len(userSpec.Groups) > 0 {
		log.Printf("  删除 %d 个组及其项目...\n", len(userSpec.Groups))
		p.deleteConfiguredGroups(userSpec.Groups)

		// 验证配置的组已删除
		if !p.verifyGroupsDeletion(userSpec.Groups, 6, 5*time.Second) {
			log.Printf("  ⚠ 警告: 部分组可能仍然存在\n")
		}
	}

	// 2. 删除用户拥有的其他组
	p.deleteUserOwnedGroups(userSpec.Username)

	// 3. 等待数据同步
	log.Printf("  等待 GitLab 内部数据同步 (10秒)...\n")
	time.Sleep(10 * time.Second)

	// 4. 删除用户
	if err := p.deleteUser(user.ID, userSpec.Username); err != nil {
		log.Printf("  ⚠ 删除用户失败: %v\n\n", err)
		return err
	}

	return nil
}

// deleteUserProjects 删除用户级项目
func (p *ResourceProcessor) deleteUserProjects(username string, projects []types.ProjectSpec) {
	for i, projSpec := range projects {
		log.Printf("  ------------------------------------------\n")
		log.Printf("  处理用户级项目 [%d/%d]: %s\n", i+1, len(projects), projSpec.Name)

		// 用户级项目的 full path 是 username/project-path
		projectPath := projSpec.Path
		if projectPath == "" {
			projectPath = projSpec.Name
		}
		fullPath := fmt.Sprintf("%s/%s", username, projectPath)
		project, _ := p.Client.GetProject(fullPath)

		if project != nil {
			log.Printf("    删除项目: %s (ID: %d)\n", projSpec.Name, project.ID)
			if err := p.Client.DeleteProject(project.ID); err != nil {
				log.Printf("    ⚠ 删除项目失败: %v\n", err)
			} else {
				log.Printf("    ✓ 项目删除成功\n")
			}
		} else {
			log.Printf("    ⚠ 项目不存在，跳过: %s\n", fullPath)
		}
	}
}

// deleteConfiguredGroups 删除配置文件中定义的组及其项目
func (p *ResourceProcessor) deleteConfiguredGroups(groups []types.GroupSpec) {
	for j, groupSpec := range groups {
		log.Printf("  ------------------------------------------\n")
		log.Printf("  处理组 [%d/%d]: %s\n", j+1, len(groups), groupSpec.Name)

		// 删除组下的项目
		if len(groupSpec.Projects) > 0 {
			log.Printf("    删除 %d 个项目...\n", len(groupSpec.Projects))
			p.deleteProjects(groupSpec.Path, groupSpec.Projects)
		}

		// 删除组
		group, _ := p.Client.GetGroup(groupSpec.Path)
		if group != nil {
			log.Printf("    删除组: %s (ID: %d)\n", groupSpec.Name, group.ID)
			if err := p.Client.DeleteGroup(group.ID); err != nil {
				log.Printf("    ⚠ 删除组失败: %v\n", err)
			} else {
				log.Printf("    ✓ 组删除成功\n")
			}
		}
	}
}

// deleteProjects 删除多个项目
func (p *ResourceProcessor) deleteProjects(groupPath string, projects []types.ProjectSpec) {
	for _, projSpec := range projects {
		fullPath := fmt.Sprintf("%s/%s", groupPath, projSpec.Path)
		project, _ := p.Client.GetProject(fullPath)

		if project != nil {
			log.Printf("      删除项目: %s (ID: %d)\n", projSpec.Name, project.ID)
			if err := p.Client.DeleteProject(project.ID); err != nil {
				log.Printf("      ⚠ 删除项目失败: %v\n", err)
			} else {
				log.Printf("      ✓ 项目删除成功\n")
			}
		}
	}
}

// verifyGroupsDeletion 验证组是否已删除
func (p *ResourceProcessor) verifyGroupsDeletion(groups []types.GroupSpec, maxRetries int, retryInterval time.Duration) bool {
	log.Printf("  等待 GitLab 处理组删除...\n")

	for retry := 1; retry <= maxRetries; retry++ {
		log.Printf("  验证组删除状态 (尝试 %d/%d)...\n", retry, maxRetries)
		time.Sleep(retryInterval)

		remainingGroups := 0
		for _, groupSpec := range groups {
			verifyGroup, _ := p.Client.GetGroup(groupSpec.Path)
			if verifyGroup != nil {
				remainingGroups++
				log.Printf("    ⚠ 组 '%s' 仍然存在\n", groupSpec.Name)
			}
		}

		if remainingGroups == 0 {
			log.Printf("  ✓ 验证通过: 配置文件中的组已彻底删除\n")
			return true
		}

		log.Printf("  还有 %d 个组未完全删除，继续等待...\n", remainingGroups)
	}

	return false
}

// deleteUserOwnedGroups 删除用户拥有的所有其他组
func (p *ResourceProcessor) deleteUserOwnedGroups(username string) {
	log.Printf("  检查用户是否还拥有其他组...\n")

	userGroups, err := p.Client.ListUserGroups(username)
	if err != nil {
		log.Printf("  ⚠ 获取用户组列表失败: %v\n", err)
		return
	}

	if len(userGroups) == 0 {
		log.Printf("  ✓ 用户没有拥有其他组\n")
		return
	}

	log.Printf("  发现用户还拥有 %d 个组，开始删除...\n", len(userGroups))
	for _, group := range userGroups {
		log.Printf("    删除组: %s (ID: %d)\n", group.FullPath, group.ID)
		if err := p.Client.DeleteGroup(group.ID); err != nil {
			log.Printf("    ⚠ 删除组失败: %v\n", err)
		} else {
			log.Printf("    ✓ 组删除成功\n")
		}
	}

	// 验证所有组已删除
	p.verifyUserGroupsDeletion(username, 10, 5*time.Second)
}

// verifyUserGroupsDeletion 验证用户的所有组是否已删除
func (p *ResourceProcessor) verifyUserGroupsDeletion(username string, maxRetries int, retryInterval time.Duration) {
	log.Printf("  等待用户所有组删除完成...\n")

	for retry := 1; retry <= maxRetries; retry++ {
		time.Sleep(retryInterval)
		log.Printf("  验证所有组删除状态 (尝试 %d/%d)...\n", retry, maxRetries)

		remainingUserGroups, _ := p.Client.ListUserGroups(username)
		if len(remainingUserGroups) == 0 {
			log.Printf("  ✓ 验证通过: 用户所有组已彻底删除\n")
			return
		}

		log.Printf("  还有 %d 个组未完全删除，继续等待...\n", len(remainingUserGroups))
		for _, g := range remainingUserGroups {
			log.Printf("    - %s (ID: %d)\n", g.FullPath, g.ID)
		}
	}
}

// deleteUser 删除用户并验证
func (p *ResourceProcessor) deleteUser(userID int, username string) error {
	log.Printf("  删除用户: %s\n", username)
	if err := p.Client.DeleteUser(userID); err != nil {
		return err
	}

	log.Printf("  ✓ 用户删除成功\n")
	log.Printf("  等待 GitLab 完成删除操作 (10秒)...\n")
	time.Sleep(10 * time.Second)

	// 验证删除
	verifyUser, _ := p.Client.GetUser(username)
	if verifyUser == nil {
		log.Printf("  ✓ 验证通过: 用户已彻底删除\n\n")
	} else {
		log.Printf("  ⚠ 验证失败: 用户可能仍然存在\n\n")
	}

	return nil
}
