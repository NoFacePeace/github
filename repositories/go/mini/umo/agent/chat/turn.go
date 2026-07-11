package chat

import (
	"github.com/NoFacePeace/github/repositories/go/mini/umo/agent/client"
	"github.com/NoFacePeace/github/repositories/go/mini/umo/agent/session"
)

type Turn struct {
	Session   *session.Session
	RequestID string
	Stream    bool
	Agent     client.Agent
	Request   *client.Request
	Message   string
}

func newTurn(req *Request, args ...any) *Turn {

	return &Turn{}
}
