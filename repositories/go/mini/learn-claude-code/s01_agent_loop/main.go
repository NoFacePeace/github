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
//	# 可选：走自建代理或第三方兼容网关时设置
//	export ANTHROPIC_BASE_URL=https://your-proxy.example.com/
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
	// ANSI 颜色：用于终端输出着色（命令黄色、提示符青色）。
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorReset  = "\033[0m"

	// 工具输出截断上限，对齐 Python 版的 50000。超过此长度的输出会被截断，
	// 防止单个工具结果撑爆下一轮请求的 token 上限。
	maxToolOutput = 50000
	// bash 执行超时。超过即 kill 进程并返回超时错误，避免 agent 卡死。
	bashTimeout = 120 * time.Second
	// 打印工具输出时的预览长度。完整输出会进历史，但终端只显示前 N 字符。
	previewLen = 200
	// 单次请求最大 token。控制每轮回复长度上限，也间接限制单轮工具调用量。
	maxTokens = 8000
)

func main() {
	// 启动横幅与交互提示。
	fmt.Println("s01: Agent Loop")
	fmt.Println("输入问题，回车发送。输入 q 退出。")
	fmt.Println()

	// 把当前工作目录注入 system prompt，让模型知道自己“身处哪个项目根”，
	// 从而在生成 bash 命令时使用正确的相对路径。
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("getwd error: %s\n", err)
		return
	}
	system := fmt.Sprintf("You are a coding agent at %s. Use bash to solve tasks. Act, don't explain.", cwd)

	// 收集 SDK 选项：API key 和 base url 都从环境变量读，留空则用 SDK 默认值
	// （默认 base url 为 https://api.anthropic.com/）。ANTHROPIC_BASE_URL 用于
	// 走自建代理或第三方兼容网关。
	var sdkOpts []aiopt.RequestOption
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		sdkOpts = append(sdkOpts, aiopt.WithAPIKey(key))
	}
	if baseURL := os.Getenv("ANTHROPIC_BASE_URL"); baseURL != "" {
		sdkOpts = append(sdkOpts, aiopt.WithBaseURL(baseURL))
	}
	// 两个选项都缺省时，SDK 会从 ANTHROPIC_API_KEY 环境变量自动取 key。
	client := ai.NewClient(sdkOpts...)

	model := os.Getenv("MODEL_ID")
	// 只暴露一个 bash 工具给模型——这是整个 agent 唯一的执行能力来源。
	tools := []ai.ToolUnionParam{bashTool()}

	// history 保存整轮对话（user / assistant / tool_result），跨多轮 REPL 保持。
	history := []ai.MessageParam{}
	scanner := bufio.NewScanner(os.Stdin)
	// 把 scanner 缓冲扩到 1MB，避免长行（如粘贴大段代码）被截断。
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for {
		// 打印提示符（青色），读取一行用户输入。
		fmt.Printf("%ss01 >> %s", colorCyan, colorReset)
		if !scanner.Scan() {
			break
		}
		query := strings.TrimSpace(scanner.Text())
		// 空、q、exit 都视为退出信号。
		if query == "" || strings.EqualFold(query, "q") || strings.EqualFold(query, "exit") {
			break
		}

		// 把用户输入作为 user 消息追加进历史，然后交给 agentLoop 跑到模型停下。
		history = append(history, ai.NewUserMessage(ai.NewTextBlock(query)))

		finalText, err := agentLoop(context.Background(), client, model, system, tools, &history)
		if err != nil {
			fmt.Println(err)
			break
		}
		// 模型最后一轮若返回文本，打印给用户。
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
		// 组装每次请求的参数：模型、上限、历史消息、可用工具。
		// history 按 *history 传入并就地追加，循环间共享同一条对话。
		params := ai.MessageNewParams{
			Model:     ai.Model(model),
			MaxTokens: maxTokens,
			Messages:  *history,
			Tools:     tools,
		}
		// system 非空时作为独立的 system block 传入（不进 messages）。
		if system != "" {
			params.System = []ai.TextBlockParam{{Text: system}}
		}
		// 发起一次 Messages API 调用。
		resp, err := client.Messages.New(ctx, params)
		if err != nil {
			return "", fmt.Errorf("agent loop round %d error: [%w]", round, err)
		}

		// 把这一轮助手回复整条回灌进历史——保留 text 和 tool_use 块，
		// 因为下一轮模型需要看到自己上一轮发出了哪些工具调用。
		*history = append(*history, resp.ToParam())

		// stop_reason 不是 tool_use 表示模型已决定停下（end_turn / max_tokens 等），
		// 取出文本返回给用户，退出循环。
		if resp.StopReason != ai.StopReasonToolUse {
			return extractText(resp.Content), nil
		}

		// 模型发了 tool_use 块：逐个执行工具，把结果作为 tool_result 收集。
		// tool_result 必须装在一条 user 消息里回传（Anthropic API 约定）。
		var results []ai.ContentBlockParamUnion
		for _, block := range resp.Content {
			// 只处理 tool_use 块，跳过 text 等其它块。
			tu, ok := block.AsAny().(ai.ToolUseBlock)
			if !ok {
				continue
			}
			cmd := bashCommand(tu.Input)
			// 打印模型要执行的命令（黄色），方便用户观察 agent 在干什么。
			fmt.Printf("%s$ %s%s\n", colorYellow, cmd, colorReset)
			out := runBash(ctx, cmd)
			// 打印前 previewLen 个字符的预览，避免长输出刷屏。
			preview := out
			if len(preview) > previewLen {
				preview = preview[:previewLen]
			}
			fmt.Println(preview)
			// 用 tool_use 的 ID 关联结果，false 表示非错误结果。
			results = append(results, ai.NewToolResultBlock(tu.ID, out, false))
		}
		// 所有工具结果汇成一条 user 消息，进入下一轮请求。
		*history = append(*history, ai.NewUserMessage(results...))
	}
}

// bashTool 返回唯一的 bash 工具定义：模型只能通过调用名为 bash 的工具、
// 传入 command 字段来执行 shell 命令。
func bashTool() ai.ToolUnionParam {
	return ai.ToolUnionParam{OfTool: &ai.ToolParam{
		Name:        "bash",
		Description: ai.String("Run a shell command."),
		// InputSchema 用 JSON Schema 描述入参结构，供模型生成合法调用。
		InputSchema: ai.ToolInputSchemaParam{
			Properties: map[string]any{
				"command": map[string]any{"type": "string"},
			},
			Required: []string{"command"},
		},
	}}
}

// bashCommand 从 tool_use 的 input（json.RawMessage）里提取 command 字段。
// 反序列化失败时原样返回原始字节，作为兜底命令串。
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
// 返回 stdout+stderr 合并文本（必要时截断）；危险命令直接拒绝，超时单独报告。
func runBash(ctx context.Context, command string) string {
	// 危险命令黑名单：匹配到就直接拦截，不执行。
	dangerous := []string{"rm -rf /", "sudo", "shutdown", "reboot", "> /dev/"}
	for _, d := range dangerous {
		if strings.Contains(command, d) {
			return "Error: Dangerous command blocked"
		}
	}

	// 命令在当前工作目录下执行，与 system prompt 中告诉模型的 cwd 一致。
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	// 设置 120s 超时，避免命令挂死把整个 agent 卡住。
	cctx, cancel := context.WithTimeout(ctx, bashTimeout)
	defer cancel()
	// 用 sh -c 执行，支持管道、重定向等 shell 语法。
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
		// 非超时错误：把 stdout/stderr/err 拼起来返回，让模型看到失败原因。
		out := strings.TrimSpace(stdout.String() + stderr.String() + err.Error())
		if out == "" {
			out = "(no output)"
		}
		return truncate(out, maxToolOutput)
	}

	// 成功：合并 stdout+stderr，空输出给个占位避免模型误判。
	out := strings.TrimSpace(stdout.String() + stderr.String())
	if out == "" {
		return "(no output)"
	}
	return truncate(out, maxToolOutput)
}

// extractText 从一轮回复的 content 块里抽出所有 text 块并拼接。
// 用于在循环结束时把模型的最终文本回复返回给用户。
func extractText(content []ai.ContentBlockUnion) string {
	var sb strings.Builder
	for _, c := range content {
		if tb, ok := c.AsAny().(ai.TextBlock); ok {
			sb.WriteString(tb.Text)
		}
	}
	return sb.String()
}

// truncate 把字符串截断到 n 字节，用于限制工具输出长度，避免超长输出
// 撑爆请求 token 上限。
func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
