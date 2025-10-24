.PHONY: help build install clean test fmt lint run build-all deps

# 变量定义
BINARY_NAME=gitlab-cli
VERSION?=0.2.0
BUILD_DIR=bin
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# 平台定义
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# 颜色定义
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## 显示帮助信息
	@echo "GitLab CLI SDK - Makefile 命令"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "环境变量:"
	@echo "  VERSION          gitlab-cli-sdk 版本 (默认: $(VERSION))"

deps: ## 下载 Go 依赖
	@echo "下载 Go 依赖..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(GREEN)✓ 依赖下载完成$(NC)"

build: deps ## 构建当前平台的二进制文件
	@echo "构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gitlab-cli
	@echo "$(GREEN)✓ 构建完成: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

build-all: deps ## 构建所有平台的二进制文件
	@echo "构建所有平台的二进制文件..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		echo "构建 $${platform%/*}/$${platform#*/}..."; \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/} ./cmd/gitlab-cli || exit 1; \
		echo "$(GREEN)✓ 构建完成: $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}$(NC)"; \
	done
	@echo ""
	@echo "$(GREEN)✓ 所有平台构建完成$(NC)"

install: build ## 安装到系统路径
	@echo "安装 $(BINARY_NAME) 到 /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✓ 安装完成$(NC)"

clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR) release
	@echo "$(GREEN)✓ 清理完成$(NC)"

test: ## 运行测试
	$(GO) test -v ./...

fmt: ## 格式化代码
	$(GO) fmt ./...

lint: ## 运行代码检查
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过..."; \
	fi

run: ## 运行程序（示例）
	$(GO) run ./cmd/gitlab-cli --help

release: clean build-all ## 创建发布包
	@echo "创建发布包..."
	@mkdir -p release
	@cd $(BUILD_DIR) && for file in $(BINARY_NAME)-*; do \
		if [ -f "$$file" ]; then \
			tar czf ../release/$$file-$(VERSION).tar.gz $$file; \
			echo "✓ 创建: release/$$file-$(VERSION).tar.gz"; \
		fi \
	done
	@echo "$(GREEN)✓ 发布包创建完成$(NC)"

.DEFAULT_GOAL := help
