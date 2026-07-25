package main

// s01: Agent Loop
//
// 整个 AI 编码 agent 的核心只有一个模式：
//
//	while stop_reason == "tool_use":
//	    response = LLM(messages, tools)
//	    执行工具
//	    把结果回灌
//
//	+----------+      +-------+      +---------+
//	|   User   | ---> |  LLM  | ---> |  Tool   |
//	|  prompt  |      |       |      | execute |
//	+----------+      +---+---+      +----+----+
//	                      ^               |
//	                      |   tool_result |
//	                      +---------------+
//	                      (loop continues)
//
// 这就是核心循环：把工具结果回灌给模型，直到模型决定停下。
// 生产级 agent 在此之上叠加策略、hook 和生命周期控制。
//
// 运行：
//
//	export ANTHROPIC_API_KEY=...
//	export MODEL_ID=...
//	go run ./mini/learn-claude-code/s01_agent_loop

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	ai "github.com/anthropics/anthropic-sdk-go"
	aiopt "github.com/anthropics/anthropic-sdk-go/option"
)

const (
	// ANSI 颜色。
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorReset  = "\033[0m"

	// 工具输出截断上限，对齐 Python 版的 50000。
	maxToolOutput = 50000
	// bash 执行超时。
	bashTimeout = 120 * time.Second
	// 打印工具输出时的预览长度。
	previewLen = 200
	// 单次请求最大 token。
	maxTokens = 8000
)

func main() {
	fmt.Println("s01: Agent Loop")
	fmt.Println("输入问题，回车发送。输入 q 退出。")
	fmt.Println()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("getwd error: %s\n", err)
		return
	}
	system := fmt.Sprintf("You are a coding agent at %s. Use bash to solve tasks. Act, don't explain.", cwd)

	var sdkOpts []aiopt.RequestOption
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		sdkOpts = append(sdkOpts, aiopt.WithAPIKey(key))
	}
	if baseURL := os.Getenv("ANTHROPIC_BASE_URL"); baseURL != "" {
		sdkOpts = append(sdkOpts, aiopt.WithBaseURL(baseURL))
	}
	client := ai.NewClient(sdkOpts...)

	model := os.Getenv("MODEL_ID")
	tools := []ai.ToolUnionParam{bashTool()}

	history := []ai.MessageParam{}
	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for {
		fmt.Printf("%ss01 >> %s", colorCyan, colorReset)
		if !scanner.Scan() {
			break
		}
		query := strings.TrimSpace(scanner.Text())
		if query == "" || strings.EqualFold(query, "q") || strings.EqualFold(query, "exit") {
			break
		}

		history = append(history, ai.NewUserMessage(ai.NewTextBlock(query)))

		finalText, err := agentLoop(context.Background(), client, model, system, tools, &history)
		if err != nil {
			fmt.Println(err)
			break
		}
		if finalText != "" {
			fmt.Println(finalText)
		}
		fmt.Println()
	}
}

// agentLoop 运行 模型→工具调用→工具结果→模型 的循环，直到助手回复不含
// tool_use 块。返回最终一轮回复里的合并文本。
// 直接用官方 SDK，忠实还原 Python 版：把整条 response.content 回灌成 assistant 消息。
func agentLoop(ctx context.Context, client ai.Client, model, system string, tools []ai.ToolUnionParam, history *[]ai.MessageParam) (string, error) {
	for round := 1; ; round++ {
		params := ai.MessageNewParams{
			Model:     ai.Model(model),
			MaxTokens: maxTokens,
			Messages:  *history,
			Tools:     tools,
		}
		if system != "" {
			params.System = []ai.TextBlockParam{{Text: system}}
		}
		resp, err := client.Messages.New(ctx, params)
		if err != nil {
			return "", fmt.Errorf("agent loop round %d error: [%w]", round, err)
		}

		// 回灌助手回复（保留 text + tool_use 块）。
		*history = append(*history, resp.ToParam())

		// 如果模型没调用工具，循环结束。
		if resp.StopReason != ai.StopReasonToolUse {
			return extractText(resp.Content), nil
		}

		// 执行每个工具调用，收集结果。
		var results []ai.ContentBlockParamUnion
		for _, block := range resp.Content {
			tu, ok := block.AsAny().(ai.ToolUseBlock)
			if !ok {
				continue
			}
			cmd := bashCommand(tu.Input)
			fmt.Printf("%s$ %s%s\n", colorYellow, cmd, colorReset)
			out := runBash(ctx, cmd)
			preview := out
			if len(preview) > previewLen {
				preview = preview[:previewLen]
			}
			fmt.Println(preview)
			results = append(results, ai.NewToolResultBlock(tu.ID, out, false))
		}
		*history = append(*history, ai.NewUserMessage(results...))
	}
}

// bashTool 返回唯一的 bash 工具定义。
func bashTool() ai.ToolUnionParam {
	return ai.ToolUnionParam{OfTool: &ai.ToolParam{
		Name:        "bash",
		Description: ai.String("Run a shell command."),
		InputSchema: ai.ToolInputSchemaParam{
			Properties: map[string]any{
				"command": map[string]any{"type": "string"},
			},
			Required: []string{"command"},
		},
	}}
}

// bashCommand 从 tool_use 输入里提取 command 字段。
func bashCommand(input json.RawMessage) string {
	var in struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return string(input)
	}
	return in.Command
}

// runBash 执行一条 shell 命令，对齐 Python 版的安全策略与超时。
func runBash(ctx context.Context, command string) string {
	dangerous := []string{"rm -rf /", "sudo", "shutdown", "reboot", "> /dev/"}
	for _, d := range dangerous {
		if strings.Contains(command, d) {
			return "Error: Dangerous command blocked"
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	cctx, cancel := context.WithTimeout(ctx, bashTimeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "sh", "-c", command)
	cmd.Dir = cwd
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// 超时时 cmd.Run 返回的 err 携带 DeadlineExceeded。
		if cctx.Err() == context.DeadlineExceeded {
			return "Error: Timeout (120s)"
		}
		out := strings.TrimSpace(stdout.String() + stderr.String() + err.Error())
		if out == "" {
			out = "(no output)"
		}
		return truncate(out, maxToolOutput)
	}

	out := strings.TrimSpace(stdout.String() + stderr.String())
	if out == "" {
		return "(no output)"
	}
	return truncate(out, maxToolOutput)
}

func extractText(content []ai.ContentBlockUnion) string {
	var sb strings.Builder
	for _, c := range content {
		if tb, ok := c.AsAny().(ai.TextBlock); ok {
			sb.WriteString(tb.Text)
		}
	}
	return sb.String()
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
