package repository

// Repository 表示一个代码仓库及其项目归属。
type Repository struct {
	ID        string
	Name      string
	GitURL    string
	ProjectID string
}
