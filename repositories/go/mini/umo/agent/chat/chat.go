package chat

import (
	"context"
	"fmt"

	"github.com/NoFacePeace/github/repositories/go/mini/umo/agent/gateway"
	"github.com/NoFacePeace/github/repositories/go/mini/umo/agent/session"
)

const (
	DefaultTitle = "chat"
)

type Chat interface {
	Prepare(args ...any) (*Turn, error)
	Execute(args ...any) (*Result, error)
}

type chat struct {
	session   session.Service
	gateway   gateway.Gateway
	assembler Assembler
}

func (c *chat) Prepare(ctx context.Context, req *Request, args ...any) (*Turn, error) {
	// 获取 session
	session, err := c.resolveSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("resolve session: [%w]", err)
	}
	agent, err := c.getConnectedAgent(session.AgentID)
	if err != nil {
		return nil, fmt.Errorf("get connected agent: [%w]", err)
	}

	if err := c.addUserMessage(ctx, session.ID, req.Message); err != nil {
		return nil, fmt.Errorf("add user message: [%w]", err)
	}
	prompt := c.assemblePrompt(ctx, req)
	mcps, err := c.listMCPServers(ctx)
	role := c.getAgentRole(ctx, agent.ID)
	canPlan := c.canPlan(role)
	tools := c.listTools(ctx)
	turn := newTurn(ctx, req, session, prompt, mcps, canPlan, tools)
	if role == "" {
		c.prepareSkillControllerTurn()
	}
	return turn, nil
}

func (c *chat) resolveSession(ctx context.Context, req *Request) (*session.Session, error) {
	id := req.ID
	if id != "" {
		session, err := c.session.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get session %s: [%w]", id, err)
		}
		return session, nil
	}
	if req.Title == "" {
		req.Title = DefaultTitle
	}

	session, err := c.session.Create()
	if err != nil {
		return nil, fmt.Errorf("create session %s: [%w]", id, err)
	}

	return session, nil
}

func (c *chat) getConnectedAgent(id string) (*gateway.Agent, error) {
	agent := c.gateway.GetAgent(id)
	if agent.Status != gateway.AgentStatusConnected {
		return nil, fmt.Errorf("agent ")
	}
	return agent, nil
}

func (c *chat) addUserMessage(args ...any) error {
	return c.session.AddMessage()
}

func (c *chat) assemblePrompt(args ...any) *AssembledPrompt {
	prompt := c.assembler.AssembleChatPrompt()
	return prompt
}

func (c *chat) listMCPServers(args ...any) (string, error) {
	return "", nil
}

func (c *chat) getAgentRole(args ...any) string {
	return ""
}

func (c *chat) listTools(args ...any) any {
	return nil
}

func (c *chat) canPlan(args ...any) bool {
	return false
}

func (c *chat) prepareSkillControllerTurn(args ...any) {

}

type Request struct {
	ID      string
	AgentID string
	ModelID string
	Title   string
	Message string
}

type Result struct {
	ID     string
	Output string
}
