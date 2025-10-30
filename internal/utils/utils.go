package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// GetVisibility 获取可见性设置，默认为 private
func GetVisibility(v string) string {
	if v == "" {
		return "private"
	}
	return v
}

// GenerateTimestampSuffix 生成时间戳后缀（格式：20060102150405）
func GenerateTimestampSuffix() string {
	return time.Now().Format("20060102150405")
}

// GenerateUsernameWithTimestamp 基于前缀生成符合 GitLab 要求的 username
// 格式：prefix-timestamp
func GenerateUsernameWithTimestamp(prefix string) string {
	timestamp := GenerateTimestampSuffix()
	username := fmt.Sprintf("%s-%s", prefix, timestamp)
	// 确保符合 GitLab username 规则：只能包含字母、数字、下划线、点号、破折号
	username = sanitizeUsername(username)
	// 限制长度为 255
	if len(username) > 255 {
		username = username[:255]
	}
	return username
}

// GenerateEmailWithTimestamp 基于前缀生成唯一的 email
// 格式：prefix-timestamp@domain
func GenerateEmailWithTimestamp(emailPrefix string) string {
	// 解析邮箱前缀和域名
	parts := strings.Split(emailPrefix, "@")
	if len(parts) != 2 {
		// 如果不是有效的邮箱格式，使用默认域名
		return fmt.Sprintf("%s-%s@test.example.com", emailPrefix, GenerateTimestampSuffix())
	}

	localPart := parts[0]
	domain := parts[1]
	timestamp := GenerateTimestampSuffix()

	return fmt.Sprintf("%s-%s@%s", localPart, timestamp, domain)
}

// GenerateGroupPathWithTimestamp 基于前缀生成符合 GitLab 要求的 group path
// 格式：prefix-timestamp
func GenerateGroupPathWithTimestamp(prefix string) string {
	timestamp := GenerateTimestampSuffix()
	path := fmt.Sprintf("%s-%s", prefix, timestamp)
	// 确保符合 GitLab group path 规则：只能包含小写字母、数字、下划线、破折号
	path = sanitizeGroupPath(path)
	// 限制长度为 255
	if len(path) > 255 {
		path = path[:255]
	}
	return path
}

// GenerateProjectPathWithTimestamp 基于前缀生成符合 GitLab 要求的 project path
// 格式：prefix-timestamp（项目 path 规则与组 path 相同）
func GenerateProjectPathWithTimestamp(prefix string) string {
	return GenerateGroupPathWithTimestamp(prefix)
}

// sanitizeUsername 清理 username，确保符合 GitLab 规则
// 允许：字母、数字、下划线、点号、破折号
func sanitizeUsername(username string) string {
	// 移除不允许的字符
	reg := regexp.MustCompile(`[^a-zA-Z0-9_.-]`)
	username = reg.ReplaceAllString(username, "")

	// 确保不以破折号或点号开头/结尾
	username = strings.Trim(username, "-.")

	return username
}

// sanitizeGroupPath 清理 group path，确保符合 GitLab 规则
// 允许：小写字母、数字、下划线、破折号
func sanitizeGroupPath(path string) string {
	// 转换为小写
	path = strings.ToLower(path)

	// 移除不允许的字符
	reg := regexp.MustCompile(`[^a-z0-9_-]`)
	path = reg.ReplaceAllString(path, "")

	// 确保不以破折号开头/结尾
	path = strings.Trim(path, "-")

	return path
}
