package chat

import (
	"context"

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

func newTurn(ctx context.Context, req *Request, args ...any) *Turn {

	return &Turn{}
}
