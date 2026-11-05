package server

type Agent struct {
	ID                 string
	PublishedVersionID string
}

type AgentVersionStatus string

const (
	AgentVersionStatusDraft AgentVersionStatus = "draft"
	AgentVersionStatusSaved AgentVersionStatus = "saved"
)

type AgentVersion struct {
	ID      string
	AgentID string
	Status  AgentVersionStatus
	// Version 为 0 表示草稿，正整数表示已保存的版本号。
	Version int
	Engine  string
	Model   string
	// SystemPrompt 定义 Agent 的固定角色和规则，空字符串表示未配置。
	SystemPrompt string
}
