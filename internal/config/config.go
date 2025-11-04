package config

import (
	"fmt"
	"os"

	"gitlab-cli-sdk/pkg/types"
	"gopkg.in/yaml.v3"
)

// CLIConfig 封装CLI配置参数
type CLIConfig struct {
	ConfigFile   string
	GitLabHost   string
	GitLabToken  string
	OutputFile   string // 输出文件路径
	TemplateFile string // 模板文件路径
}

// LoadGitLabCredentials 从环境变量或命令行参数加载 GitLab 凭证
func LoadGitLabCredentials(cfg *CLIConfig) error {
	if cfg.GitLabHost == "" {
		cfg.GitLabHost = os.Getenv("GITLAB_URL")
		if cfg.GitLabHost == "" {
			return fmt.Errorf("GitLab host is required (use --host or GITLAB_URL env)")
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

// SaveOutput 保存输出结果到 YAML 文件
func SaveOutput(outputFile string, output *types.OutputConfig) error {
	data, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("marshal output: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}

	return nil
}
