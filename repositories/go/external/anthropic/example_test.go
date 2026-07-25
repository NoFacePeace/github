//go:build example

// 这些示例需要真实 API 调用，默认不参与普通构建与测试。
// 运行方式：
//
//	ANTHROPIC_API_KEY=sk-ant-... go test -tags=example ./external/anthropic/... -run Example -v
package anthropic_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/NoFacePeace/github/repositories/go/external/anthropic"
)

func skipIfNoKey() bool {
	return os.Getenv("ANTHROPIC_API_KEY") == ""
}

func ExampleClient_complete() {
	if skipIfNoKey() {
		fmt.Println("(skipped: no ANTHROPIC_API_KEY)")
		return
	}
	c, err := anthropic.New()
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := c.Complete(context.Background(), anthropic.CompleteParams{
		System:   "Be concise.",
		Messages: []anthropic.Message{anthropic.NewUserMessage("用一句话解释什么是四元数。")},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.Text)
}

func ExampleClient_agentLoop() {
	if skipIfNoKey() {
		fmt.Println("(skipped: no ANTHROPIC_API_KEY)")
		return
	}
	c, err := anthropic.New()
	if err != nil {
		fmt.Println(err)
		return
	}
	tools := []anthropic.Tool{
		{
			Name:        "get_weather",
			Description: "获取指定城市的天气",
			InputSchema: map[string]any{
				"city": map[string]any{
					"type":        "string",
					"description": "城市名称",
				},
			},
			Required: []string{"city"},
			Call: func(ctx context.Context, input json.RawMessage) (string, error) {
				var in struct {
					City string `json:"city"`
				}
				if err := json.Unmarshal(input, &in); err != nil {
					return "", err
				}
				return fmt.Sprintf(`{"city":%q,"temp":26,"unit":"celsius"}`, in.City), nil
			},
		},
	}
	result, err := c.AgentLoop(context.Background(), anthropic.AgentLoopParams{
		System:   "Be concise. Use the get_weather tool when asked about weather.",
		Messages: []anthropic.Message{anthropic.NewUserMessage("上海今天天气怎么样？")},
		Tools:    tools,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result.Text)
}
