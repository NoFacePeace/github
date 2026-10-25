package model

import "github.com/NoFacePeace/github/repositories/go/demo/agent/tool"

type Request struct {
	Messages []Message
	Tools    []tool.Definition
}
