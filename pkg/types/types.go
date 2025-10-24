package types

// UserConfig 用户配置结构
type UserConfig struct {
	Users []UserSpec `yaml:"users"`
}

// UserSpec 用户规格定义
type UserSpec struct {
	Username string      `yaml:"username"`
	Email    string      `yaml:"email"`
	Name     string      `yaml:"name"`
	Password string      `yaml:"password"`
	Groups   []GroupSpec `yaml:"groups"` // 支持多个组
}

// GroupSpec 组规格定义
type GroupSpec struct {
	Name       string        `yaml:"name"`
	Path       string        `yaml:"path"`
	Visibility string        `yaml:"visibility"`
	Projects   []ProjectSpec `yaml:"projects"` // 每个组下有多个项目
}

// ProjectSpec 项目规格定义
type ProjectSpec struct {
	Name        string `yaml:"name"`
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
	Visibility  string `yaml:"visibility"`
}
