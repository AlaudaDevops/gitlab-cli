package client

import (
	"fmt"
	"log"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// GitLabClient GitLab SDK 客户端封装
type GitLabClient struct {
	client *gitlab.Client
}

// NewGitLabClient 创建新的 GitLab 客户端
func NewGitLabClient(host, token string) (*GitLabClient, error) {
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(host+"/api/v4"))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &GitLabClient{
		client: client,
	}, nil
}

// CheckAuth 检查认证和管理员权限
func (c *GitLabClient) CheckAuth() error {
	user, _, err := c.client.Users.CurrentUser()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if !user.IsAdmin {
		return fmt.Errorf("current user is not admin")
	}

	log.Printf("✓ 认证成功（用户: %s, 管理员权限）\n", user.Username)
	return nil
}

// GetUser 获取用户
func (c *GitLabClient) GetUser(username string) (*gitlab.User, error) {
	users, _, err := c.client.Users.ListUsers(&gitlab.ListUsersOptions{
		Username: gitlab.Ptr(username),
	})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}

	return users[0], nil
}

// GetUserNamespaceID 获取用户的 namespace ID
func (c *GitLabClient) GetUserNamespaceID(username string) (int, error) {
	// 列出用户的所有 namespace
	namespaces, _, err := c.client.Namespaces.ListNamespaces(&gitlab.ListNamespacesOptions{
		Search: gitlab.Ptr(username),
	}, gitlab.WithSudo(username))
	if err != nil {
		return 0, err
	}

	// 查找用户的个人 namespace（kind 为 "user"）
	for _, ns := range namespaces {
		if ns.Kind == "user" && ns.Path == username {
			return ns.ID, nil
		}
	}

	return 0, fmt.Errorf("未找到用户 %s 的 namespace", username)
}

// ListUserProjects 列出用户的个人项目（不属于任何组的项目）
func (c *GitLabClient) ListUserProjects(username string) ([]*gitlab.Project, error) {
	// 获取用户信息
	user, err := c.GetUser(username)
	if err != nil || user == nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 列出用户拥有的所有项目，过滤出个人命名空间下的项目
	projects, _, err := c.client.Projects.ListUserProjects(user.ID, &gitlab.ListProjectsOptions{
		Owned: gitlab.Ptr(true),
	})
	if err != nil {
		return nil, err
	}

	// 过滤出用户个人命名空间下的项目（namespace.kind == "user"）
	var userProjects []*gitlab.Project
	for _, project := range projects {
		if project.Namespace.Kind == "user" {
			userProjects = append(userProjects, project)
		}
	}

	return userProjects, nil
}

// CreateUser 创建用户
func (c *GitLabClient) CreateUser(username, email, name, password string) (*gitlab.User, error) {
	user, _, err := c.client.Users.CreateUser(&gitlab.CreateUserOptions{
		Email:            gitlab.Ptr(email),
		Username:         gitlab.Ptr(username),
		Name:             gitlab.Ptr(name),
		Password:         gitlab.Ptr(password),
		SkipConfirmation: gitlab.Ptr(true),
	})
	if err != nil {
		return nil, err
	}

	// 确保用户激活和批准
	err = c.client.Users.UnblockUser(user.ID)
	if err != nil {
		log.Printf("  ⚠ 解除封锁用户失败: %v\n", err)
	}

	err = c.client.Users.ApproveUser(user.ID)
	if err != nil {
		log.Printf("  ⚠ 批准用户失败: %v\n", err)
	}

	return user, nil
}

// GetGroup 获取组
func (c *GitLabClient) GetGroup(groupPath string) (*gitlab.Group, error) {
	group, resp, err := c.client.Groups.GetGroup(groupPath, &gitlab.GetGroupOptions{})
	if err != nil {
		// 404 表示组不存在
		if resp != nil && resp.StatusCode == 404 {
			return nil, nil
		}
		return nil, err
	}

	return group, nil
}

// CreateGroup 创建组
func (c *GitLabClient) CreateGroup(username, groupName, groupPath, visibility string) (*gitlab.Group, error) {
	vis := gitlab.VisibilityValue(visibility)

	group, _, err := c.client.Groups.CreateGroup(&gitlab.CreateGroupOptions{
		Name:                 gitlab.Ptr(groupName),
		Path:                 gitlab.Ptr(groupPath),
		Visibility:           &vis,
		RequestAccessEnabled: gitlab.Ptr(false),
	}, gitlab.WithSudo(username))
	if err != nil {
		return nil, err
	}

	return group, nil
}

// GetProject 获取项目
func (c *GitLabClient) GetProject(fullPath string) (*gitlab.Project, error) {
	project, resp, err := c.client.Projects.GetProject(fullPath, &gitlab.GetProjectOptions{})
	if err != nil {
		// 404 表示项目不存在
		if resp != nil && resp.StatusCode == 404 {
			return nil, nil
		}
		return nil, err
	}

	return project, nil
}

// CreateProject 创建项目
func (c *GitLabClient) CreateProject(username string, namespaceID int, projectName, projectPath, description, visibility string) (*gitlab.Project, error) {
	vis := gitlab.VisibilityValue(visibility)

	project, _, err := c.client.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:                 gitlab.Ptr(projectName),
		Path:                 gitlab.Ptr(projectPath),
		NamespaceID:          gitlab.Ptr(namespaceID),
		Description:          gitlab.Ptr(description),
		Visibility:           &vis,
		InitializeWithReadme: gitlab.Ptr(true),
		IssuesEnabled:        gitlab.Ptr(true),
		MergeRequestsEnabled: gitlab.Ptr(true),
		WikiEnabled:          gitlab.Ptr(true),
	}, gitlab.WithSudo(username))
	if err != nil {
		return nil, err
	}

	return project, nil
}

// DeleteProject 删除项目
func (c *GitLabClient) DeleteProject(projectID int) error {
	_, err := c.client.Projects.DeleteProject(projectID, nil)
	return err
}

// DeleteGroup 删除组
func (c *GitLabClient) DeleteGroup(groupID int) error {
	_, err := c.client.Groups.DeleteGroup(groupID, nil)
	return err
}

// ListUserGroups 列出用户拥有的所有组
func (c *GitLabClient) ListUserGroups(username string) ([]*gitlab.Group, error) {
	// 获取用户拥有的组
	opts := &gitlab.ListGroupsOptions{
		Owned:       gitlab.Ptr(true),
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}

	groups, _, err := c.client.Groups.ListGroups(opts, gitlab.WithSudo(username))
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// DeleteUser 删除用户
func (c *GitLabClient) DeleteUser(userID int) error {
	// 注意：GitLab 的用户删除可能是"软删除"，用户会被标记为删除但仍然存在
	// 完全删除用户可能需要一段时间，或者用户会保留在系统中但处于非活跃状态
	_, err := c.client.Users.DeleteUser(userID)
	return err
}

// CreatePersonalAccessToken 为用户创建 Personal Access Token
func (c *GitLabClient) CreatePersonalAccessToken(userID int, name string, scopes []string, expiresAt string) (string, error) {
	// 将字符串日期转换为 ISOTime 类型
	isoTime, err := gitlab.ParseISOTime(expiresAt)
	if err != nil {
		return "", fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}

	opt := &gitlab.CreatePersonalAccessTokenOptions{
		Name:      gitlab.Ptr(name),
		Scopes:    &scopes,
		ExpiresAt: &isoTime,
	}

	token, _, err := c.client.Users.CreatePersonalAccessToken(userID, opt)
	if err != nil {
		return "", fmt.Errorf("failed to create personal access token: %w", err)
	}

	// 返回 token 的 value
	return token.Token, nil
}

// ListAllUsers 列出所有用户（支持搜索过滤）
func (c *GitLabClient) ListAllUsers(searchPrefix string) ([]*gitlab.User, error) {
	var allUsers []*gitlab.User
	page := 1
	perPage := 100

	for {
		opts := &gitlab.ListUsersOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}

		// 如果指定了搜索前缀，添加搜索条件
		if searchPrefix != "" {
			opts.Search = gitlab.Ptr(searchPrefix)
		}

		users, resp, err := c.client.Users.ListUsers(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %w", err)
		}

		allUsers = append(allUsers, users...)

		// 如果没有更多页面，退出循环
		if resp.NextPage == 0 {
			break
		}

		page = resp.NextPage
	}

	return allUsers, nil
}
