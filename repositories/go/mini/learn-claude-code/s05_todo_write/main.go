package main

// s05: TodoWrite — 在 s04（hooks）基础上加一个计划工具。
//
//	+---------+      +-------+      +------------------+
//	|  User   | ---> |  LLM  | ---> | TOOL_HANDLERS    |
//	| prompt  |      |       |      |  bash            |
//	+---------+      +---+---+      |  read_file       |
//	                    ^         |  write_file      |
//	                    | result  |  edit_file       |
//	                    +---------+  glob            |
//	                                  todo_write ← NEW
//	                               +------------------+
//	                                    |
//	                     in-memory currentTodos
//	                                    |
//	                    if roundsSinceTodo >= 3:
//	                      inject <reminder>
//
// 相比 s04 的变化:
//   + todo_write 工具 + runTodoWrite 实现
//   + Nag 提醒（连续 3 轮未更新 todo 就注入一条 reminder）
//   + SYSTEM prompt 加上“plan before execute”计划引导
//   + agent_loop 内 roundsSinceTodo 计数器
//   循环体本身不变：新工具经 dispatchTool 自动分发。
//
// 注意：s05 的 hook 系统是 s04 的简化版——permission_hook 只保留 deny-list 硬拒，
// 去掉了破坏性命令询问和写工作区外询问；只剩 contextInject / log / summary 三个 hook。
// 这里按 s05 源码忠实复刻这个简化版。
//
// 运行:
//
//	export ANTHROPIC_API_KEY=...
//	export MODEL_ID=...
//	# 可选：走自建代理或第三方兼容网关时设置
//	export ANTHROPIC_BASE_URL=https://your-proxy.example.com/
//	go run ./mini/learn-claude-code/s05_todo_write

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	ai "github.com/anthropics/anthropic-sdk-go"
	aiopt "github.com/anthropics/anthropic-sdk-go/option"
)

const (
	// ANSI 颜色：用于终端输出着色。
	// 青色：工具名/提示符/进行中图标；黄色：任务清单标题/警告；红色：硬拒；
	// 绿色：已完成图标；灰色：hook 日志。
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorGray   = "\033[90m"
	colorReset  = "\033[0m"

	// 工具输出截断上限，对齐 Python 版的 50000。
	maxToolOutput = 50000
	// bash 执行超时。
	bashTimeout = 120 * time.Second
	// 打印工具输出时的预览长度。
	previewLen = 200
	// 单次请求最大 token。
	maxTokens = 8000
	// 连续多少轮没更新 todo 就注入一次 nag 提醒。
	nagRounds = 3
)

// stdinReader 贯穿 REPL 输入（保留与 s03/s04 一致，便于未来加回询问式 hook）。
var stdinReader = bufio.NewReader(os.Stdin)

// currentTodos 是 todo_write 维护的内存计划清单，跨 REPL 轮保持。
// 对应 Python 的全局 CURRENT_TODOS。
var currentTodos []todoItem

// todoItem 是单条计划项：内容 + 状态（pending/in_progress/completed）。
type todoItem struct {
	Content string `json:"content"`
	Status  string `json:"status"`
}

func main() {
	// 启动横幅与交互提示。
	fmt.Println("s05: TodoWrite — plan before execute, nag if you forget")
	fmt.Println("Type a question, press Enter. Type q to quit.")
	fmt.Println()

	// 把当前工作目录作为工作区根：所有 read/write/edit/glob 都限制在它下面。
	workdir, err := os.Getwd()
	if err != nil {
		fmt.Printf("getwd error: %s\n", err)
		return
	}
	// s05 system prompt 改为要求计划先行：多步任务先用 todo_write 规划，边做边更状态。
	system := fmt.Sprintf(
		"You are a coding agent at %s. "+
			"Before starting any multi-step task, use todo_write to plan your steps. "+
			"Update status as you go.",
		workdir,
	)

	// 收集 SDK 选项：API key 和 base url 都从环境变量读，留空则用 SDK 默认值。
	var sdkOpts []aiopt.RequestOption
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		sdkOpts = append(sdkOpts, aiopt.WithAPIKey(key))
	}
	if baseURL := os.Getenv("ANTHROPIC_BASE_URL"); baseURL != "" {
		sdkOpts = append(sdkOpts, aiopt.WithBaseURL(baseURL))
	}
	client := ai.NewClient(sdkOpts...)

	model := os.Getenv("MODEL_ID")
	// s05 工具集 = s04 的 5 个 + todo_write。
	tools := allTools()

	// 注册 hook（s05 简化版：permission 只剩 deny-list，PostToolUse largeOutput 已移除）。
	registerHooks()

	// history 保存整轮对话，跨多轮 REPL 保持。
	history := []ai.MessageParam{}
	// roundsSinceTodo 是 nag 提醒的计数器：每轮 tool_use +1，调 todo_write 归零，
	// 达到 nagRounds 就注入一条 reminder。放在 main 里跨 agentLoop 多次调用保持。
	roundsSinceTodo := 0
	for {
		// 打印提示符（青色），读取一行用户输入。
		fmt.Printf("%ss05 >> %s", colorCyan, colorReset)
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			break
		}
		query := strings.TrimSpace(line)
		if query == "" || strings.EqualFold(query, "q") || strings.EqualFold(query, "exit") {
			if query == "" {
				continue
			}
			break
		}

		// UserPromptSubmit hook。
		triggerHooks(hookUserPromptSubmit, query)

		// 把用户输入追加进历史，交给 agentLoop 跑到模型停下。
		history = append(history, ai.NewUserMessage(ai.NewTextBlock(query)))

		finalText, err := agentLoop(context.Background(), client, model, system, tools, &history, &roundsSinceTodo)
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
// 结构与 s04 一致；新增 nag 提醒逻辑：每轮进 LLM 前检查 roundsSinceTodo。
func agentLoop(ctx context.Context, client ai.Client, model, system string, tools []ai.ToolUnionParam, history *[]ai.MessageParam, roundsSinceTodo *int) (string, error) {
	for round := 1; ; round++ {
		// s05 nag：若连续 nagRounds 轮未更新 todo，注入一条 reminder 并归零计数器。
		// 注意必须在 history 非空时注入（避免首次空对话就提醒）。
		if *roundsSinceTodo >= nagRounds && len(*history) > 0 {
			*history = append(*history, ai.NewUserMessage(ai.NewTextBlock("<reminder>Update your todos.</reminder>")))
			*roundsSinceTodo = 0
		}

		// 组装每次请求的参数。
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

		// 模型决定停下：触发 Stop hook，返回 force 则塞回历史继续，否则结束。
		if resp.StopReason != ai.StopReasonToolUse {
			force := triggerHooks(hookStop, *history)
			if force != "" {
				*history = append(*history, ai.NewUserMessage(ai.NewTextBlock(force)))
				continue
			}
			return extractText(resp.Content), nil
		}

		// 本轮是 tool_use：计数器 +1，等执行到 todo_write 时再归零。
		*roundsSinceTodo++

		var results []ai.ContentBlockParamUnion
		for _, block := range resp.Content {
			tu, ok := block.AsAny().(ai.ToolUseBlock)
			if !ok {
				continue
			}
			fmt.Printf("%s> %s%s\n", colorCyan, tu.Name, colorReset)

			// PreToolUse hook（s05: deny-list 硬拒 + 日志）。
			blocked := triggerHooks(hookPreToolUse, tu.Name, tu.Input)
			if blocked != "" {
				results = append(results, ai.NewToolResultBlock(tu.ID, blocked, false))
				continue
			}

			out := dispatchTool(ctx, tu.Name, tu.Input)

			// PostToolUse hook（s05 已移除 largeOutput，仅留扩展点空触发）。
			triggerHooks(hookPostToolUse, tu.Name, tu.Input, out)

			// s05: 调了 todo_write 就把 nag 计数器归零。
			if tu.Name == "todo_write" {
				*roundsSinceTodo = 0
			}

			// 预览打印 + 完整结果回灌。
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

// ═══════════════════════════════════════════════════════════
//  NEW in s05: todo_write 工具 — 仅做计划，不执行任何操作
// ═══════════════════════════════════════════════════════════

// runTodoWrite 接收 todos 数组，校验后写入内存 currentTodos，并打印彩色清单。
// 返回给模型的是简短确认串；详细的清单只打印给人看。
func runTodoWrite(todos []todoItem) string {
	if err := validateTodos(todos); err != "" {
		return err
	}
	currentTodos = todos

	// 打印彩色任务清单（给人看，不回灌给模型）。
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%s## Current Tasks%s\n", colorYellow, colorReset))
	for _, t := range currentTodos {
		icon := statusIcon(t.Status)
		sb.WriteString(fmt.Sprintf("  [%s] %s\n", icon, t.Content))
	}
	fmt.Print(sb.String())
	return fmt.Sprintf("Updated %d tasks", len(currentTodos))
}

// validateTodos 校验 todos 数组，返回空串表示通过，非空串为错误信息。
// 对齐 Python 版 _normalize_todos 的校验：每项必须是对象、有 content+status、status 合法。
func validateTodos(todos []todoItem) string {
	for i, t := range todos {
		if t.Content == "" && t.Status == "" {
			return fmt.Sprintf("Error: todos[%d] must be an object", i)
		}
		if t.Content == "" {
			return fmt.Sprintf("Error: todos[%d] missing 'content'", i)
		}
		switch t.Status {
		case "pending", "in_progress", "completed":
		default:
			return fmt.Sprintf("Error: todos[%d] has invalid status '%s'", i, t.Status)
		}
	}
	return ""
}

// statusIcon 把状态映射成彩色图标，对齐 Python 版的 icon 字典。
func statusIcon(status string) string {
	switch status {
	case "pending":
		return " "
	case "in_progress":
		return fmt.Sprintf("%s▸%s", colorCyan, colorReset)
	case "completed":
		return fmt.Sprintf("%s✓%s", colorGreen, colorReset)
	default:
		return " "
	}
}

// ═══════════════════════════════════════════════════════════
//  FROM s04 (simplified): Hook 系统
// ═══════════════════════════════════════════════════════════

type hookEvent string

const (
	hookUserPromptSubmit hookEvent = "UserPromptSubmit"
	hookPreToolUse       hookEvent = "PreToolUse"
	hookPostToolUse      hookEvent = "PostToolUse"
	hookStop             hookEvent = "Stop"
)

// hookFunc 是 hook 回调的统一签名。返回非空字符串表示“阻止/干预”。
type hookFunc func(args ...any) string

var hooks = map[hookEvent][]hookFunc{
	hookUserPromptSubmit: {},
	hookPreToolUse:       {},
	hookPostToolUse:      {},
	hookStop:             {},
}

func registerHook(event hookEvent, cb hookFunc) {
	hooks[event] = append(hooks[event], cb)
}

// triggerHooks 按 event 依次调用已注册 hook。
// 任一 hook 返回非空字符串即短路返回该字符串；否则返回空串。
func triggerHooks(event hookEvent, args ...any) string {
	for _, cb := range hooks[event] {
		if result := cb(args...); result != "" {
			return result
		}
	}
	return ""
}

// hookArgs 解包 hook 收到的变参：toolName + input。
func hookArgs(args []any) (name string, input map[string]any) {
	if len(args) >= 1 {
		name, _ = args[0].(string)
	}
	if len(args) >= 2 {
		input, _ = args[1].(map[string]any)
	}
	return
}

// denyList 是 bash 硬拒绝列表（s05 简化版 permission_hook 只用这个）。
var denyList = []string{"rm -rf /", "sudo", "shutdown", "reboot", "mkfs", "dd if="}

// permissionHook 是 PreToolUse hook：bash 命令命中 denyList 直接拒。
// s05 简化版：去掉了破坏性命令询问和写工作区外询问。
func permissionHook(args ...any) string {
	name, input := hookArgs(args)
	if name == "bash" {
		cmd, _ := input["command"].(string)
		for _, p := range denyList {
			if strings.Contains(cmd, p) {
				fmt.Printf("\n%s⛔ Blocked: '%s'%s\n", colorRed, p, colorReset)
				return "Permission denied"
			}
		}
	}
	return ""
}

// logHook 是 PreToolUse hook：把每次工具调用打一行灰色日志。
func logHook(args ...any) string {
	name, _ := hookArgs(args)
	fmt.Printf("%s[HOOK] %s%s\n", colorGray, name, colorReset)
	return ""
}

// contextInjectHook 是 UserPromptSubmit hook：打印当前工作目录。
func contextInjectHook(args ...any) string {
	workdir, _ := os.Getwd()
	fmt.Printf("%s[HOOK] UserPromptSubmit: working in %s%s\n", colorGray, workdir, colorReset)
	return ""
}

// summaryHook 是 Stop hook：统计本次会话用了几次工具调用。
func summaryHook(args ...any) string {
	if len(args) < 1 {
		return ""
	}
	history, ok := args[0].([]ai.MessageParam)
	if !ok {
		return ""
	}
	toolCount := 0
	for _, m := range history {
		for _, b := range m.Content {
			if b.OfToolResult != nil {
				toolCount++
			}
		}
	}
	fmt.Printf("%s[HOOK] Stop: session used %d tool calls%s\n", colorGray, toolCount, colorReset)
	return ""
}

// registerHooks 注册 s05 简化版 hook 集。
// permissionHook 在 logHook 之前，避免被拒的调用也打日志。
func registerHooks() {
	registerHook(hookUserPromptSubmit, contextInjectHook)
	registerHook(hookPreToolUse, permissionHook)
	registerHook(hookPreToolUse, logHook)
	registerHook(hookStop, summaryHook)
}

// ═══════════════════════════════════════════════════════════
//  FROM s02-s04 (unchanged): 工具实现、安全路径、工具定义、分发
// ═══════════════════════════════════════════════════════════

// safePath 把相对路径解析到工作目录下，并拒绝任何逃逸出工作目录的路径。
func safePath(p string) (string, error) {
	workdir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	workdirResolved, err := filepath.EvalSymlinks(workdir)
	if err != nil {
		workdirResolved = workdir
	}
	var abs string
	if filepath.IsAbs(p) {
		abs = filepath.Clean(p)
	} else {
		abs = filepath.Join(workdirResolved, p)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		parentResolved, perr := filepath.EvalSymlinks(filepath.Dir(abs))
		if perr != nil {
			parentResolved = filepath.Dir(abs)
		}
		resolved = filepath.Join(parentResolved, filepath.Base(abs))
	}
	workdirPrefix := filepath.Clean(workdirResolved) + string(os.PathSeparator)
	if !strings.HasPrefix(resolved+string(os.PathSeparator), workdirPrefix) {
		return "", fmt.Errorf("Path escapes workspace: %s", p)
	}
	return resolved, nil
}

func runBash(ctx context.Context, command string) string {
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

func runRead(path string, limit int) string {
	abs, err := safePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if limit > 0 && limit < len(lines) {
		lines = append(lines[:limit], fmt.Sprintf("... (%d more lines)", len(lines)-limit))
	}
	return strings.Join(lines, "\n")
}

func runWrite(path, content string) string {
	abs, err := safePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return fmt.Sprintf("Wrote %d bytes to %s", len(content), path)
}

func runEdit(path, oldText, newText string) string {
	abs, err := safePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	text := string(data)
	if !strings.Contains(text, oldText) {
		return fmt.Sprintf("Error: text not found in %s", path)
	}
	edited := strings.Replace(text, oldText, newText, 1)
	if err := os.WriteFile(abs, []byte(edited), 0o644); err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return fmt.Sprintf("Edited %s", path)
}

func runGlob(pattern string) string {
	workdir, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	matches, err := globMatch(workdir, pattern)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	workdirResolved, err := filepath.EvalSymlinks(workdir)
	if err != nil {
		workdirResolved = workdir
	}
	workdirPrefix := filepath.Clean(workdirResolved) + string(os.PathSeparator)
	results := make([]string, 0, len(matches))
	for _, m := range matches {
		resolved, err := filepath.EvalSymlinks(m)
		if err != nil {
			resolved = m
		}
		if !strings.HasPrefix(resolved+string(os.PathSeparator), workdirPrefix) {
			continue
		}
		if rel, err := filepath.Rel(workdir, m); err == nil {
			results = append(results, rel)
		}
	}
	if len(results) == 0 {
		return "(no matches)"
	}
	return strings.Join(results, "\n")
}

func globMatch(root, pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		return filepath.Glob(filepath.Join(root, pattern))
	}
	var results []string
	segments := splitSegments(pattern)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		if matchSegs(splitSegments(rel), segments) {
			results = append(results, path)
		}
		return nil
	})
	return results, err
}

func splitSegments(s string) []string {
	parts := strings.Split(filepath.ToSlash(s), "/")
	out := parts[:0]
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func matchSegs(path, pat []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			if len(pat) == 1 {
				return true
			}
			for i := 0; i <= len(path); i++ {
				if matchSegs(path[i:], pat[1:]) {
					return true
				}
			}
			return false
		}
		if len(path) == 0 {
			return false
		}
		ok, err := filepath.Match(pat[0], path[0])
		if err != nil || !ok {
			return false
		}
		path = path[1:]
		pat = pat[1:]
	}
	return len(path) == 0
}

// allTools 返回全部 6 个工具定义（s04 的 5 个 + todo_write）。
func allTools() []ai.ToolUnionParam {
	return []ai.ToolUnionParam{
		bashToolDef(),
		readFileToolDef(),
		writeFileToolDef(),
		editFileToolDef(),
		globToolDef(),
		todoWriteToolDef(),
	}
}

func bashToolDef() ai.ToolUnionParam {
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

func readFileToolDef() ai.ToolUnionParam {
	return ai.ToolUnionParam{OfTool: &ai.ToolParam{
		Name:        "read_file",
		Description: ai.String("Read file contents."),
		InputSchema: ai.ToolInputSchemaParam{
			Properties: map[string]any{
				"path":  map[string]any{"type": "string"},
				"limit": map[string]any{"type": "integer"},
			},
			Required: []string{"path"},
		},
	}}
}

func writeFileToolDef() ai.ToolUnionParam {
	return ai.ToolUnionParam{OfTool: &ai.ToolParam{
		Name:        "write_file",
		Description: ai.String("Write content to a file."),
		InputSchema: ai.ToolInputSchemaParam{
			Properties: map[string]any{
				"path":    map[string]any{"type": "string"},
				"content": map[string]any{"type": "string"},
			},
			Required: []string{"path", "content"},
		},
	}}
}

func editFileToolDef() ai.ToolUnionParam {
	return ai.ToolUnionParam{OfTool: &ai.ToolParam{
		Name:        "edit_file",
		Description: ai.String("Replace exact text in a file once."),
		InputSchema: ai.ToolInputSchemaParam{
			Properties: map[string]any{
				"path":     map[string]any{"type": "string"},
				"old_text": map[string]any{"type": "string"},
				"new_text": map[string]any{"type": "string"},
			},
			Required: []string{"path", "old_text", "new_text"},
		},
	}}
}

func globToolDef() ai.ToolUnionParam {
	return ai.ToolUnionParam{OfTool: &ai.ToolParam{
		Name:        "glob",
		Description: ai.String("Find files matching a glob pattern."),
		InputSchema: ai.ToolInputSchemaParam{
			Properties: map[string]any{
				"pattern": map[string]any{"type": "string"},
			},
			Required: []string{"pattern"},
		},
	}}
}

// todoWriteToolDef 是 s05 新增工具：创建/管理当前会话的任务清单。
// 入参 todos 是数组，每项含 content + status(pending/in_progress/completed)。
func todoWriteToolDef() ai.ToolUnionParam {
	return ai.ToolUnionParam{OfTool: &ai.ToolParam{
		Name:        "todo_write",
		Description: ai.String("Create and manage a task list for your current coding session."),
		InputSchema: ai.ToolInputSchemaParam{
			Properties: map[string]any{
				"todos": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"content": map[string]any{"type": "string"},
							"status": map[string]any{
								"type": "string",
								"enum": []string{"pending", "in_progress", "completed"},
							},
						},
						"required": []string{"content", "status"},
					},
				},
			},
			Required: []string{"todos"},
		},
	}}
}

// dispatchTool 按工具名查表执行，找不到则返回 Unknown。
func dispatchTool(ctx context.Context, name string, input json.RawMessage) string {
	switch name {
	case "bash":
		var in struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
		return runBash(ctx, in.Command)
	case "read_file":
		var in struct {
			Path  string `json:"path"`
			Limit int    `json:"limit"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
		return runRead(in.Path, in.Limit)
	case "write_file":
		var in struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
		return runWrite(in.Path, in.Content)
	case "edit_file":
		var in struct {
			Path    string `json:"path"`
			OldText string `json:"old_text"`
			NewText string `json:"new_text"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
		return runEdit(in.Path, in.OldText, in.NewText)
	case "glob":
		var in struct {
			Pattern string `json:"pattern"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
		return runGlob(in.Pattern)
	case "todo_write":
		var in struct {
			Todos []todoItem `json:"todos"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
		return runTodoWrite(in.Todos)
	default:
		return fmt.Sprintf("Unknown: %s", name)
	}
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
