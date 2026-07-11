package gateway

type AgentStatus string

const (
	AgentStatusConnected AgentStatus = "connected"
)

type Agent struct {
	ID     string
	Status AgentStatus
}
