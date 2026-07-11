package session

type Service interface {
	Get(args ...any) (*Session, error)
	Create(args ...any) (*Session, error)
	AddMessage(args ...any) error
}

type Session struct {
	ID      string
	AgentID string
}
