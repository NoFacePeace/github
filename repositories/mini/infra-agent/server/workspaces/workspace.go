// Package workspaces 提供工作区相关的领域模型和业务逻辑。
package workspaces

// Workspace 表示一个工作区。
type Workspace struct {
	ID        string
	Directory string
}
