package config

import (
	"fmt"
	"os"

	"gitlab-cli-sdk/pkg/types"
	"gopkg.in/yaml.v3"
)

// CLIConfig 封装CLI配置参数
type CLIConfig struct {
	ConfigFile  string
	GitLabHost  string
	GitLabToken string
}

// LoadGitLabCredentials 从环境变量或命令行参数加载 GitLab 凭证
func LoadGitLabCredentials(cfg *CLIConfig) error {
	if cfg.GitLabHost == "" {
		cfg.GitLabHost = os.Getenv("GITLAB_HOST")
		if cfg.GitLabHost == "" {
			return fmt.Errorf("GitLab host is required (use --host or GITLAB_HOST env)")
		}
	}
	if cfg.GitLabToken == "" {
		cfg.GitLabToken = os.Getenv("GITLAB_TOKEN")
		if cfg.GitLabToken == "" {
			return fmt.Errorf("GitLab token is required (use --token or GITLAB_TOKEN env)")
		}
	}
	return nil
}

// Load 加载配置文件
func Load(configFile string) (*types.UserConfig, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg types.UserConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}
