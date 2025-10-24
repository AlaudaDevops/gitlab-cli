package utils

// GetVisibility 获取可见性设置，默认为 private
func GetVisibility(v string) string {
	if v == "" {
		return "private"
	}
	return v
}
