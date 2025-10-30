package types

// UserConfig 用户配置结构
type UserConfig struct {
	Users []UserSpec `yaml:"users"`
}

// UserSpec 用户规格定义
type UserSpec struct {
	NameMode string        `yaml:"nameMode,omitempty"` // 命名模式: "prefix" (添加时间戳) 或 "name" (不添加时间戳)，默认为 "prefix"
	Username string        `yaml:"username"`
	Email    string        `yaml:"email"`
	Name     string        `yaml:"name"`
	Password string        `yaml:"password"`
	Token    *TokenSpec    `yaml:"token"`    // Personal Access Token 配置
	Groups   []GroupSpec   `yaml:"groups"`   // 支持多个组
	Projects []ProjectSpec `yaml:"projects"` // 用户级别的项目（不属于任何组）
}

// TokenSpec Personal Access Token 规格定义
type TokenSpec struct {
	Scope     []string `yaml:"scope"`      // Token 的权限范围
	ExpiresAt string   `yaml:"expires_at"` // 过期时间 (格式: YYYY-MM-DD)
}

// GroupSpec 组规格定义
type GroupSpec struct {
	NameMode   string        `yaml:"nameMode,omitempty"` // 命名模式: "prefix" (添加时间戳) 或 "name" (不添加时间戳)，继承 UserSpec.NameMode
	Name       string        `yaml:"name"`
	Path       string        `yaml:"path"`
	Visibility string        `yaml:"visibility"`
	Projects   []ProjectSpec `yaml:"projects"` // 每个组下有多个项目
}

// ProjectSpec 项目规格定义
type ProjectSpec struct {
	NameMode    string `yaml:"nameMode,omitempty"` // 命名模式: "prefix" (添加时间戳) 或 "name" (不添加时间戳)，继承 GroupSpec.NameMode
	Name        string `yaml:"name"`
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
	Visibility  string `yaml:"visibility"`
}

// ========================================
// 输出结果类型
// ========================================

// OutputConfig 输出配置结构
type OutputConfig struct {
	Endpoint string       `yaml:"endpoint"`
	Scheme   string       `yaml:"scheme"`
	Host     string       `yaml:"host"`
	Users    []UserOutput `yaml:"users"`
}

// UserOutput 用户输出结果
type UserOutput struct {
	Username string          `yaml:"username"`
	Email    string          `yaml:"email"`
	Name     string          `yaml:"name"`
	UserID   int             `yaml:"user_id"`
	Password string          `yaml:"password,omitempty"` // 用户密码
	Token    *TokenOutput    `yaml:"token,omitempty"`
	Groups   []GroupOutput   `yaml:"groups,omitempty"`
	Projects []ProjectOutput `yaml:"projects,omitempty"` // 用户级别的项目
}

// TokenOutput Token 输出结果
type TokenOutput struct {
	Value     string   `yaml:"value"`
	Scope     []string `yaml:"scope"`
	ExpiresAt string   `yaml:"expires_at"`
}

// GroupOutput 组输出结果
type GroupOutput struct {
	Name       string          `yaml:"name"`
	Path       string          `yaml:"path"`
	GroupID    int             `yaml:"group_id"`
	Visibility string          `yaml:"visibility"`
	Projects   []ProjectOutput `yaml:"projects,omitempty"`
}

// ProjectOutput 项目输出结果
type ProjectOutput struct {
	Name        string `yaml:"name"`
	Path        string `yaml:"path"`
	ProjectID   int    `yaml:"project_id"`
	Description string `yaml:"description"`
	Visibility  string `yaml:"visibility"`
	WebURL      string `yaml:"web_url,omitempty"`
}
