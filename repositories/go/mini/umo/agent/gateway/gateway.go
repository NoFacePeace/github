package gateway

type Gateway interface {
	GetAgent(args ...any) *Agent
}
