// Package workspaces 提供工作区相关的领域模型和业务逻辑。
package workspaces

// Status 表示工作区的生命周期状态。
type Status string

const (
	// StatusActive 表示工作区已启用。
	StatusActive Status = "active"
	// StatusArchived 表示工作区已归档。
	StatusArchived Status = "archived"
)

// Workspace 表示一个工作区。
type Workspace struct {
	ID        string
	Directory string
	Status    Status
}
