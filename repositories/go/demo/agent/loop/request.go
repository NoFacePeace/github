package loop

import (
	"github.com/NoFacePeace/github/repositories/go/demo/agent/model"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/tool"
)

type Request struct {
	Model    model.Model
	Messages []model.Message
	Tools    []tool.Tool
	MaxSteps int // 最大模型调用轮数，零值使用默认配置
}
