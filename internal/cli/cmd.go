package cli

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gitlab-cli-sdk/internal/config"
	"gitlab-cli-sdk/internal/processor"
	"gitlab-cli-sdk/internal/template"
	"gitlab-cli-sdk/pkg/client"
	"gitlab-cli-sdk/pkg/types"

	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// BuildRootCommand 构建根命令
func BuildRootCommand(cfg *config.CLIConfig, version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "gitlab-cli",
		Short:   "GitLab 用户和项目自动化管理工具（使用 GitLab Go SDK）",
		Version: version,
		Long: `GitLab CLI 是基于官方 GitLab Go SDK 的用户和项目自动化管理工具。
它通过 YAML 配置文件批量创建和管理 GitLab 用户、组和项目。

特性：
  - 使用官方 GitLab Go SDK (gitlab.com/gitlab-org/api/client-go)
  - 无需外部依赖，纯 Go 实现
  - 类型安全的 API 调用
  - 更好的性能和错误处理

前置要求：
  - GitLab 管理员权限的 Personal Access Token (api + sudo scopes)`,
	}

	// 添加子命令
	rootCmd.AddCommand(buildUserCommand(cfg))

	return rootCmd
}

// buildUserCommand 构建用户管理命令
func buildUserCommand(cfg *config.CLIConfig) *cobra.Command {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "用户管理命令",
	}

	userCmd.AddCommand(buildUserCreateCommand(cfg))
	userCmd.AddCommand(buildUserCleanupCommand(cfg))
	userCmd.AddCommand(buildUserDeleteCommand(cfg))
	userCmd.AddCommand(buildUserListCommand(cfg))
	userCmd.AddCommand(buildUserDeleteByPrefixCommand(cfg))

	return userCmd
}

// buildUserCreateCommand 构建用户创建命令
func buildUserCreateCommand(cfg *config.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "根据配置文件创建用户、组和项目",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserCreate(cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.ConfigFile, "config", "f", "../test-users.yaml", "配置文件路径")
	cmd.Flags().StringVar(&cfg.GitLabHost, "host", "", "GitLab 主机地址")
	cmd.Flags().StringVar(&cfg.GitLabToken, "token", "", "GitLab Personal Access Token")
	cmd.Flags().StringVarP(&cfg.OutputFile, "output", "o", "", "输出结果到 YAML 文件")
	cmd.Flags().StringVarP(&cfg.TemplateFile, "template", "t", "", "使用模板文件格式化输出")

	return cmd
}

// buildUserCleanupCommand 构建用户清理命令
func buildUserCleanupCommand(cfg *config.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "清理配置文件中定义的用户",
		Long: `清理配置文件中定义的用户及其所有资源。
默认只删除创建日期超过2天的用户，可通过 --days-old 参数调整。
设置 --days-old=0 将删除所有匹配的用户（不考虑创建时间）。

示例:
  gitlab-cli user cleanup -f config.yaml                    # 只删除2天前创建的用户
  gitlab-cli user cleanup -f config.yaml --days-old 7       # 只删除7天前创建的用户
  gitlab-cli user cleanup -f config.yaml --days-old 0       # 删除所有用户（不检查创建时间）`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserCleanup(cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.ConfigFile, "config", "f", "../test-users.yaml", "配置文件路径")
	cmd.Flags().StringVar(&cfg.GitLabHost, "host", "", "GitLab 主机地址")
	cmd.Flags().StringVar(&cfg.GitLabToken, "token", "", "GitLab Personal Access Token")
	cmd.Flags().IntVar(&cfg.DaysOld, "days-old", 2, "只删除创建日期超过指定天数的用户（0表示删除所有用户）")

	return cmd
}

// buildUserDeleteCommand 构建用户删除命令
func buildUserDeleteCommand(cfg *config.CLIConfig) *cobra.Command {
	var usernames string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "删除指定用户名的用户及其项目和组",
		Long: `根据用户名删除用户及其所有资源（项目和组）。
支持删除多个用户，用户名之间用逗号分隔。

示例:
  gitlab-cli user delete --username user1
  gitlab-cli user delete --username user1,user2,user3`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserDelete(cfg, usernames)
		},
	}

	cmd.Flags().StringVar(&usernames, "username", "", "要删除的用户名（多个用户用逗号分隔）")
	cmd.Flags().StringVar(&cfg.GitLabHost, "host", "", "GitLab 主机地址")
	cmd.Flags().StringVar(&cfg.GitLabToken, "token", "", "GitLab Personal Access Token")
	_ = cmd.MarkFlagRequired("username")

	return cmd
}

// runUserCreate 执行用户创建命令
func runUserCreate(cfg *config.CLIConfig) error {
	gitlabClient, err := initializeClient(cfg)
	if err != nil {
		return err
	}

	userConfig, err := config.Load(cfg.ConfigFile)
	if err != nil {
		return err
	}

	log.Printf("\n找到 %d 个用户配置\n\n", len(userConfig.Users))

	proc := &processor.ResourceProcessor{Client: gitlabClient}

	// 收集所有用户的输出结果
	var userOutputs []types.UserOutput

	for i, userSpec := range userConfig.Users {
		log.Printf("==========================================\n")
		log.Printf("处理用户 [%d/%d]: %s\n", i+1, len(userConfig.Users), userSpec.Username)
		log.Printf("==========================================\n")

		userOutput, err := proc.ProcessUserCreation(userSpec)
		if err != nil {
			return err
		}

		// 将输出结果添加到列表
		if userOutput != nil {
			userOutputs = append(userOutputs, *userOutput)
		}

		log.Printf("\n✓ 用户 '%s' 处理完成\n\n", userSpec.Username)
	}

	log.Println("========================================")
	log.Println("✓ 批量创建完成")
	log.Println("========================================")

	// 如果指定了输出文件，保存结果
	if cfg.OutputFile != "" {
		// 从 GitLabHost 解析 endpoint、scheme、host 和 port
		endpoint, scheme, host, port := parseGitLabHostURL(cfg.GitLabHost)

		output := &types.OutputConfig{
			Endpoint: endpoint,
			Scheme:   scheme,
			Host:     host,
			Port:     port,
			Users:    userOutputs,
		}

		// 如果指定了模板文件，使用模板渲染
		if cfg.TemplateFile != "" {
			log.Printf("\n使用模板渲染输出: %s\n", cfg.TemplateFile)
			log.Printf("保存结果到文件: %s\n", cfg.OutputFile)
			if err := template.SaveTemplateOutput(cfg.TemplateFile, cfg.OutputFile, output); err != nil {
				return err
			}
			log.Printf("✓ 使用模板渲染完成，结果已保存到: %s\n", cfg.OutputFile)
		} else {
			// 使用默认 YAML 格式
			log.Printf("\n保存结果到文件: %s\n", cfg.OutputFile)
			if err := config.SaveOutput(cfg.OutputFile, output); err != nil {
				return err
			}
			log.Printf("✓ 结果已保存到: %s\n", cfg.OutputFile)
		}
	}

	return nil
}

// runUserCleanup 执行用户清理命令
func runUserCleanup(cfg *config.CLIConfig) error {
	gitlabClient, err := initializeClient(cfg)
	if err != nil {
		return err
	}

	userConfig, err := config.Load(cfg.ConfigFile)
	if err != nil {
		return err
	}

	log.Printf("\n找到 %d 个用户配置\n", len(userConfig.Users))
	if cfg.DaysOld > 0 {
		log.Printf("只删除创建日期超过 %d 天的用户\n\n", cfg.DaysOld)
	} else {
		log.Printf("将删除所有匹配的用户（不检查创建时间）\n\n")
	}

	proc := &processor.ResourceProcessor{Client: gitlabClient}

	processedCount := 0
	skippedCount := 0

	for i, userSpec := range userConfig.Users {
		log.Printf("==========================================\n")
		log.Printf("处理 [%d/%d]: %s\n", i+1, len(userConfig.Users), userSpec.Username)
		log.Printf("==========================================\n")

		deleted, err := proc.ProcessUserCleanup(userSpec, cfg.DaysOld)
		if err != nil {
			log.Printf("  ⚠ 处理用户 %s 时出错: %v\n", userSpec.Username, err)
			continue
		}

		if deleted {
			processedCount++
		} else {
			skippedCount++
		}
	}

	log.Println("========================================")
	log.Printf("✓ 批量清理完成 (已删除: %d, 已跳过: %d)\n", processedCount, skippedCount)
	log.Println("========================================")
	return nil
}

// runUserDelete 执行用户删除命令
func runUserDelete(cfg *config.CLIConfig, usernames string) error {
	gitlabClient, err := initializeClient(cfg)
	if err != nil {
		return err
	}

	// 解析用户名列表（以逗号分隔）
	usernameList := strings.Split(usernames, ",")
	// 去除空格
	for i, username := range usernameList {
		usernameList[i] = strings.TrimSpace(username)
	}

	log.Printf("\n准备删除 %d 个用户\n\n", len(usernameList))

	proc := &processor.ResourceProcessor{Client: gitlabClient}

	for i, username := range usernameList {
		if username == "" {
			continue
		}

		log.Printf("==========================================\n")
		log.Printf("处理 [%d/%d]: %s\n", i+1, len(usernameList), username)
		log.Printf("==========================================\n")

		if err := proc.ProcessUserDelete(username); err != nil {
			log.Printf("  ⚠ 删除用户 %s 时出错: %v\n", username, err)
			continue
		}
	}

	log.Println("========================================")
	log.Println("✓ 批量删除完成")
	log.Println("========================================")
	return nil
}

// buildUserListCommand 构建用户列表命令
func buildUserListCommand(cfg *config.CLIConfig) *cobra.Command {
	var searchPrefix string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有用户或搜索特定前缀的用户",
		Long: `列出 GitLab 上的用户。
可以使用 --prefix 参数搜索特定前缀的用户。

示例:
  gitlab-cli user list
  gitlab-cli user list --prefix tektoncd`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserList(cfg, searchPrefix)
		},
	}

	cmd.Flags().StringVar(&searchPrefix, "prefix", "", "搜索用户名前缀")
	cmd.Flags().StringVar(&cfg.GitLabHost, "host", "", "GitLab 主机地址")
	cmd.Flags().StringVar(&cfg.GitLabToken, "token", "", "GitLab Personal Access Token")

	return cmd
}

// buildUserDeleteByPrefixCommand 构建按前缀批量删除用户命令
func buildUserDeleteByPrefixCommand(cfg *config.CLIConfig) *cobra.Command {
	var prefix string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "delete-by-prefix",
		Short: "删除所有匹配前缀的用户及其资源",
		Long: `根据用户名前缀批量删除用户及其所有资源（项目和组）。
支持 --dry-run 模式预览将要删除的用户。
默认只删除创建日期超过2天的用户，可通过 --days-old 参数调整。

示例:
  gitlab-cli user delete-by-prefix --prefix tektoncd --dry-run            # 预览2天前创建的用户
  gitlab-cli user delete-by-prefix --prefix tektoncd                      # 删除2天前创建的用户
  gitlab-cli user delete-by-prefix --prefix tektoncd --days-old 7         # 删除7天前创建的用户
  gitlab-cli user delete-by-prefix --prefix tektoncd --days-old 0         # 删除所有匹配前缀的用户`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserDeleteByPrefix(cfg, prefix, dryRun)
		},
	}

	cmd.Flags().StringVar(&prefix, "prefix", "", "要删除的用户名前缀")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "只显示将要删除的用户，不实际删除")
	cmd.Flags().IntVar(&cfg.DaysOld, "days-old", 2, "只删除创建日期超过指定天数的用户（0表示删除所有用户）")
	cmd.Flags().StringVar(&cfg.GitLabHost, "host", "", "GitLab 主机地址")
	cmd.Flags().StringVar(&cfg.GitLabToken, "token", "", "GitLab Personal Access Token")
	_ = cmd.MarkFlagRequired("prefix")

	return cmd
}

// runUserList 执行用户列表命令
func runUserList(cfg *config.CLIConfig, searchPrefix string) error {
	gitlabClient, err := initializeClient(cfg)
	if err != nil {
		return err
	}

	log.Printf("\n正在获取用户列表...\n")
	if searchPrefix != "" {
		log.Printf("搜索前缀: %s\n", searchPrefix)
	}

	users, err := gitlabClient.ListAllUsers(searchPrefix)
	if err != nil {
		return err
	}

	log.Printf("\n找到 %d 个用户:\n", len(users))
	log.Println("========================================")
	for i, user := range users {
		log.Printf("[%d] ID: %d | 用户名: %s | 姓名: %s | 邮箱: %s | 管理员: %v\n",
			i+1, user.ID, user.Username, user.Name, user.Email, user.IsAdmin)
	}
	log.Println("========================================")

	return nil
}

// runUserDeleteByPrefix 执行按前缀批量删除用户命令
func runUserDeleteByPrefix(cfg *config.CLIConfig, prefix string, dryRun bool) error {
	gitlabClient, err := initializeClient(cfg)
	if err != nil {
		return err
	}

	log.Printf("\n正在搜索用户名以 '%s' 开头的用户...\n", prefix)

	users, err := gitlabClient.ListAllUsers(prefix)
	if err != nil {
		return err
	}

	// 过滤出真正以指定前缀开头的用户（GitLab 的搜索可能返回包含该字符串的所有用户）
	var matchedUsers []*gitlab.User
	for _, user := range users {
		if strings.HasPrefix(user.Username, prefix) {
			matchedUsers = append(matchedUsers, user)
		}
	}

	if len(matchedUsers) == 0 {
		log.Printf("\n未找到以 '%s' 开头的用户\n", prefix)
		return nil
	}

	// 根据创建日期过滤用户
	var usersToDelete []*gitlab.User
	if cfg.DaysOld > 0 {
		log.Printf("\n按创建日期过滤（只处理 %d 天前创建的用户）...\n", cfg.DaysOld)
		for _, user := range matchedUsers {
			if user.CreatedAt == nil {
				log.Printf("  [跳过] %s - 无法获取创建时间\n", user.Username)
				continue
			}

			daysSinceCreation := int(time.Since(*user.CreatedAt).Hours() / 24)
			if daysSinceCreation >= cfg.DaysOld {
				usersToDelete = append(usersToDelete, user)
			} else {
				log.Printf("  [跳过] %s - 创建于 %d 天前（未超过 %d 天）\n",
					user.Username, daysSinceCreation, cfg.DaysOld)
			}
		}
	} else {
		usersToDelete = matchedUsers
	}

	if len(usersToDelete) == 0 {
		log.Printf("\n没有符合条件的用户需要删除\n")
		return nil
	}

	log.Printf("\n找到 %d 个符合条件的用户:\n", len(usersToDelete))
	log.Println("========================================")
	for i, user := range usersToDelete {
		createdInfo := ""
		if user.CreatedAt != nil {
			daysSinceCreation := int(time.Since(*user.CreatedAt).Hours() / 24)
			createdInfo = fmt.Sprintf(" | 创建于: %s (%d 天前)",
				user.CreatedAt.Format("2006-01-02"), daysSinceCreation)
		}
		log.Printf("[%d] ID: %d | 用户名: %s | 姓名: %s | 邮箱: %s%s\n",
			i+1, user.ID, user.Username, user.Name, user.Email, createdInfo)
	}
	log.Println("========================================")

	if dryRun {
		log.Printf("\n[DRY-RUN] 以上 %d 个用户将被删除（当前为预览模式）\n", len(usersToDelete))
		if cfg.DaysOld > 0 {
			log.Printf("[DRY-RUN] 已过滤掉 %d 个不满足时间条件的用户\n", len(matchedUsers)-len(usersToDelete))
		}
		return nil
	}

	log.Printf("\n准备删除 %d 个用户及其所有资源...\n", len(usersToDelete))
	if cfg.DaysOld > 0 {
		log.Printf("（已跳过 %d 个创建时间未超过 %d 天的用户）\n", len(matchedUsers)-len(usersToDelete), cfg.DaysOld)
	}

	proc := &processor.ResourceProcessor{Client: gitlabClient}

	for i, user := range usersToDelete {
		log.Printf("==========================================\n")
		log.Printf("处理 [%d/%d]: %s (ID: %d)\n", i+1, len(usersToDelete), user.Username, user.ID)
		log.Printf("==========================================\n")

		if err := proc.ProcessUserDelete(user.Username); err != nil {
			log.Printf("  ⚠ 删除用户 %s 时出错: %v\n", user.Username, err)
			continue
		}
	}

	log.Println("========================================")
	log.Printf("✓ 批量删除完成（共处理 %d 个用户）\n", len(usersToDelete))
	log.Println("========================================")
	return nil
}

// initializeClient 初始化并验证 GitLab 客户端
func initializeClient(cfg *config.CLIConfig) (*client.GitLabClient, error) {
	if err := config.LoadGitLabCredentials(cfg); err != nil {
		return nil, err
	}

	gitlabClient, err := client.NewGitLabClient(cfg.GitLabHost, cfg.GitLabToken)
	if err != nil {
		return nil, err
	}

	log.Println("检查 GitLab 连接和权限...")
	if err := gitlabClient.CheckAuth(); err != nil {
		return nil, err
	}

	return gitlabClient, nil
}

// parseGitLabHostURL 从 GitLab Host URL 解析出 endpoint、scheme 和 host
func parseGitLabHostURL(gitlabHost string) (endpoint, scheme, host string, port int) {
	// 默认值
	scheme = "https"
	port = 443
	endpoint = gitlabHost

	// 去除尾部斜杠
	endpoint = strings.TrimSuffix(endpoint, "/")

	// 检查是否包含 scheme
	if strings.HasPrefix(endpoint, "http://") {
		scheme = "http"
		port = 80
		host = strings.TrimPrefix(endpoint, "http://")
	} else if strings.HasPrefix(endpoint, "https://") {
		scheme = "https"
		port = 443
		host = strings.TrimPrefix(endpoint, "https://")
	} else {
		// 如果没有 scheme，则 host 就是原始的 endpoint
		host = endpoint
	}

	// 检查 host 中是否包含端口号
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		if len(parts) == 2 {
			host = parts[0]
			// 尝试解析端口号
			if parsedPort, err := strconv.Atoi(parts[1]); err == nil {
				port = parsedPort
			}
		}
	}

	// 重新添加 scheme 构造完整的 endpoint
	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		// 默认端口，不需要在 endpoint 中显示
		endpoint = scheme + "://" + host
	} else {
		// 非默认端口，需要在 endpoint 中显示
		endpoint = scheme + "://" + host + ":" + strconv.Itoa(port)
	}

	return endpoint, scheme, host, port
}
