package main

import (
	"os"

	"gitlab-cli-sdk/internal/cli"
	"gitlab-cli-sdk/internal/config"
)

// Version 应用程序版本号，通过编译时 -ldflags 注入
var Version = "dev"

func main() {
	cfg := &config.CLIConfig{}
	rootCmd := cli.BuildRootCommand(cfg, Version)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
