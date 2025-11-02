package cli

import (
	"log"
	"strconv"
	"strings"

	"gitlab-cli-sdk/internal/config"
	"gitlab-cli-sdk/internal/processor"
	"gitlab-cli-sdk/internal/template"
	"gitlab-cli-sdk/pkg/client"
	"gitlab-cli-sdk/pkg/types"

	"github.com/spf13/cobra"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserCleanup(cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.ConfigFile, "config", "f", "../test-users.yaml", "配置文件路径")
	cmd.Flags().StringVar(&cfg.GitLabHost, "host", "", "GitLab 主机地址")
	cmd.Flags().StringVar(&cfg.GitLabToken, "token", "", "GitLab Personal Access Token")

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

	log.Printf("\n找到 %d 个用户配置\n\n", len(userConfig.Users))

	proc := &processor.ResourceProcessor{Client: gitlabClient}

	for i, userSpec := range userConfig.Users {
		log.Printf("==========================================\n")
		log.Printf("处理 [%d/%d]: %s\n", i+1, len(userConfig.Users), userSpec.Username)
		log.Printf("==========================================\n")

		if err := proc.ProcessUserCleanup(userSpec); err != nil {
			log.Printf("  ⚠ 处理用户 %s 时出错: %v\n", userSpec.Username, err)
			continue
		}
	}

	log.Println("========================================")
	log.Println("✓ 批量清理完成")
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
