package main

// s06: Subagent — 用全新 messages[] 派生子 agent，实现上下文隔离。
//
//	Parent Agent                           Subagent
//	+------------------+                  +------------------+
//	| messages=[...]   |                  | messages=[task]  | <-- fresh
//	|                  |   dispatch        |                  |
//	| tool: task       | ----------------> | own loop         |
//	|   prompt="..."   |                  |   bash/read/...  |
//	|                  |   summary only    |   (max 30 turns) |
//	| result = "..."   | <---------------- | return last text |
//	+------------------+                  +------------------+
//	      ^                                      |
//	      |       intermediate results DISCARDED  |
//	      +--------------------------------------+
//
// 相比 s05 的变化:
//   + task 工具 + spawnSubagent()（全新 messages[]）
//   + 安全上限：每个子 agent 最多 30 轮
//   + extractText() 辅助函数
//   子 agent 不能派生子子 agent（子工具集不含 task）。
//   主循环不变：task 经 dispatchTool 自动分发。
//
// 运行:
//
//	export ANTHROPIC_API_KEY=...
//	export MODEL_ID=...
//	# 可选：走自建代理或第三方兼容网关时设置
//	export ANTHROPIC_BASE_URL=https://your-proxy.example.com/
//	go run ./mini/learn-claude-code/s06_subagent

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
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorGray   = "\033[90m"
	colorPurple = "\033[35m" // 子 agent 标记色
	colorReset  = "\033[0m"

	maxToolOutput = 50000
	bashTimeout   = 120 * time.Second
	previewLen    = 200
	maxTokens     = 8000
	nagRounds     = 3
// subMaxTurns 是子 agent 的安全轮数上限，对齐 Python 版的 30。
	subMaxTurns = 30
)

var stdinReader = bufio.NewReader(os.Stdin)

var currentTodos []todoItem

type todoItem struct {
	Content string `json:"content"`
	Status  string `json:"status"`
}

func main() {
	fmt.Println("s06: Subagent — spawn sub-agents with fresh context, summary only")
	fmt.Println("Type a question, press Enter. Type q to quit.")
	fmt.Println()

	workdir, err := os.Getwd()
	if err != nil {
		fmt.Printf("getwd error: %s\n", err)
		return
	}
	// s06 system prompt 提示：复杂子问题用 task 工具派生子 agent。
	system := fmt.Sprintf(
		"You are a coding agent at %s. "+
			"For complex sub-problems, use the task tool to spawn a subagent.",
		workdir,
	)

	var sdkOpts []aiopt.RequestOption
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		sdkOpts = append(sdkOpts, aiopt.WithAPIKey(key))
	}
	if baseURL := os.Getenv("ANTHROPIC_BASE_URL"); baseURL != "" {
		sdkOpts = append(sdkOpts, aiopt.WithBaseURL(baseURL))
	}
	client := ai.NewClient(sdkOpts...)

	model := os.Getenv("MODEL_ID")
	// 父 agent 工具集 = s05 的 6 个 + task（追加在 allTools 末尾）。
	tools := allTools()

	// s06: 子 agent 专属工具集——只读/写/编辑/执行，不含 task（防递归）。
	subTools := subAgentTools()

	registerHooks()

	subSystem := fmt.Sprintf(
		"You are a coding agent at %s. "+
			"Complete the task you were given, then return a concise summary. "+
			"Do not delegate further.",
		workdir,
	)

	history := []ai.MessageParam{}
	roundsSinceTodo := 0
	for {
		fmt.Printf("%ss06 >> %s", colorCyan, colorReset)
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

		triggerHooks(hookUserPromptSubmit, query)
		history = append(history, ai.NewUserMessage(ai.NewTextBlock(query)))

		finalText, err := agentLoop(context.Background(), client, model, system, tools, &history, &roundsSinceTodo, subSystem, subTools)
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

// agentLoop 运行 模型→工具调用→工具结果→模型 的循环。
// 与 s05 一致；唯一新增是 tools 里多了 task，dispatchTool 会把它分发给 spawnSubagent。
func agentLoop(ctx context.Context, client ai.Client, model, system string, tools []ai.ToolUnionParam, history *[]ai.MessageParam, roundsSinceTodo *int, subSystem string, subTools []ai.ToolUnionParam) (string, error) {
	for round := 1; ; round++ {
		// s05 nag 提醒。
		if *roundsSinceTodo >= nagRounds && len(*history) > 0 {
			*history = append(*history, ai.NewUserMessage(ai.NewTextBlock("<reminder>Update your todos.</reminder>")))
			*roundsSinceTodo = 0
		}

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
		*history = append(*history, resp.ToParam())

		if resp.StopReason != ai.StopReasonToolUse {
			force := triggerHooks(hookStop, *history)
			if force != "" {
				*history = append(*history, ai.NewUserMessage(ai.NewTextBlock(force)))
				continue
			}
			return extractText(resp.Content), nil
		}

		*roundsSinceTodo++

		var results []ai.ContentBlockParamUnion
		for _, block := range resp.Content {
			tu, ok := block.AsAny().(ai.ToolUseBlock)
			if !ok {
				continue
			}
			fmt.Printf("%s> %s%s\n", colorCyan, tu.Name, colorReset)

			blocked := triggerHooks(hookPreToolUse, tu.Name, tu.Input)
			if blocked != "" {
				results = append(results, ai.NewToolResultBlock(tu.ID, blocked, false))
				continue
			}

			// task 工具派生子 agent：需要 client/model/subSystem/subTools，单独处理。
			var out string
			if tu.Name == "task" {
				out = dispatchTask(ctx, client, model, subSystem, subTools, tu.Input)
			} else {
				out = dispatchTool(ctx, tu.Name, tu.Input)
			}

			triggerHooks(hookPostToolUse, tu.Name, tu.Input, out)

			if tu.Name == "todo_write" {
				*roundsSinceTodo = 0
			}

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
//  NEW in s06: 子 agent — 全新 messages[]，只返回摘要
// ═══════════════════════════════════════════════════════════

// dispatchTask 解析 task 工具入参，派生子 agent，返回摘要文本。
// 对应 Python 的: TOOL_HANDLERS["task"] = spawn_subagent。
func dispatchTask(ctx context.Context, client ai.Client, model, subSystem string, subTools []ai.ToolUnionParam, input json.RawMessage) string {
	var in struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return spawnSubagent(ctx, client, model, subSystem, subTools, in.Description)
}

// spawnSubagent 派生一个子 agent：
//   - 全新 messages[]，只含 task description（上下文隔离）
//   - 最多 subMaxTurns 轮（安全上限）
//   - 子 agent 也跑 PreToolUse/PostToolUse hook（权限同样生效）
//   - 子工具集不含 task（防递归）
//   - 返回最后一条 assistant 文本作为摘要；中间历史全部丢弃
//
// 对齐 Python 版 spawn_subagent，包括超限回退逻辑。
func spawnSubagent(ctx context.Context, client ai.Client, model, subSystem string, subTools []ai.ToolUnionParam, description string) string {
	fmt.Printf("\n%s[Subagent spawned]%s\n", colorPurple, colorReset)
	// 全新上下文：只有一条 user 消息（任务描述）。
	subHistory := []ai.MessageParam{ai.NewUserMessage(ai.NewTextBlock(description))}

	stoppedNormally := false
	for turn := 0; turn < subMaxTurns; turn++ {
		params := ai.MessageNewParams{
			Model:     ai.Model(model),
			MaxTokens: maxTokens,
			Messages:  subHistory,
			Tools:     subTools,
		}
		if subSystem != "" {
			params.System = []ai.TextBlockParam{{Text: subSystem}}
		}
		resp, err := client.Messages.New(ctx, params)
		if err != nil {
			// API 错误：把错误作为摘要返回给父 agent。
			fmt.Printf("%s[Subagent error]%s %s\n", colorPurple, colorReset, err)
			return fmt.Sprintf("Subagent error: %s", err)
		}
		subHistory = append(subHistory, resp.ToParam())

		// 模型停下（非 tool_use）：正常结束，跳出。
		if resp.StopReason != ai.StopReasonToolUse {
			stoppedNormally = true
			break
		}

		// 执行子 agent 的工具调用。
		var results []ai.ContentBlockParamUnion
		for _, block := range resp.Content {
			tu, ok := block.AsAny().(ai.ToolUseBlock)
			if !ok {
				continue
			}
			// 子 agent 也跑 hook（权限同样生效）。
			blocked := triggerHooks(hookPreToolUse, tu.Name, tu.Input)
			if blocked != "" {
				results = append(results, ai.NewToolResultBlock(tu.ID, blocked, false))
				continue
			}
			out := dispatchTool(ctx, tu.Name, tu.Input)
			triggerHooks(hookPostToolUse, tu.Name, tu.Input, out)
			// 子 agent 工具调用预览（缩进打印，区别于父 agent）。
			preview := out
			if len(preview) > 100 {
				preview = preview[:100]
			}
			fmt.Printf("  %s[sub] %s: %s%s\n", colorGray, tu.Name, preview, colorReset)
			results = append(results, ai.NewToolResultBlock(tu.ID, out, false))
		}
		subHistory = append(subHistory, ai.NewUserMessage(results...))
	}

	// 摘要提取：优先取最后一条 assistant 文本；
	// 超限兜底——往前找最近一条 assistant 文本；都没有就给占位串。
	result := ""
	if stoppedNormally && len(subHistory) > 0 {
		// 最后一条是结束轮的 assistant 消息，取其文本。
		result = extractTextFromMessage(subHistory[len(subHistory)-1])
	}
	if result == "" {
		// 回退：从后往前找 assistant 消息。
		for i := len(subHistory) - 1; i >= 0; i-- {
			if subHistory[i].Role == "assistant" {
				if t := extractTextFromMessage(subHistory[i]); t != "" {
					result = t
					break
				}
			}
		}
	}
	if result == "" {
		result = fmt.Sprintf("Subagent stopped after %d turns without final answer.", subMaxTurns)
	}
	fmt.Printf("%s[Subagent done]%s\n", colorPurple, colorReset)
	return result
}

// extractTextFromMessage 从一条 MessageParam 的 content 块里抽 text。
func extractTextFromMessage(m ai.MessageParam) string {
	var sb strings.Builder
	for _, b := range m.Content {
		if b.OfText != nil {
			sb.WriteString(b.OfText.Text)
		}
	}
	return sb.String()
}

// ═══════════════════════════════════════════════════════════
//  FROM s05: todo_write 工具
// ═══════════════════════════════════════════════════════════

func runTodoWrite(todos []todoItem) string {
	if err := validateTodos(todos); err != "" {
		return err
	}
	currentTodos = todos
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%s## Current Tasks%s\n", colorYellow, colorReset))
	for _, t := range currentTodos {
		sb.WriteString(fmt.Sprintf("  [%s] %s\n", statusIcon(t.Status), t.Content))
	}
	fmt.Print(sb.String())
	return fmt.Sprintf("Updated %d tasks", len(currentTodos))
}

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

func triggerHooks(event hookEvent, args ...any) string {
	for _, cb := range hooks[event] {
		if result := cb(args...); result != "" {
			return result
		}
	}
	return ""
}

func hookArgs(args []any) (name string, input map[string]any) {
	if len(args) >= 1 {
		name, _ = args[0].(string)
	}
	if len(args) >= 2 {
		input, _ = args[1].(map[string]any)
	}
	return
}

var denyList = []string{"rm -rf /", "sudo", "shutdown", "reboot", "mkfs", "dd if="}

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

func logHook(args ...any) string {
	name, _ := hookArgs(args)
	fmt.Printf("%s[HOOK] %s%s\n", colorGray, name, colorReset)
	return ""
}

func contextInjectHook(args ...any) string {
	workdir, _ := os.Getwd()
	fmt.Printf("%s[HOOK] UserPromptSubmit: working in %s%s\n", colorGray, workdir, colorReset)
	return ""
}

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

func registerHooks() {
	registerHook(hookUserPromptSubmit, contextInjectHook)
	registerHook(hookPreToolUse, permissionHook)
	registerHook(hookPreToolUse, logHook)
	registerHook(hookStop, summaryHook)
}

// ═══════════════════════════════════════════════════════════
//  FROM s02-s05 (unchanged): 工具实现、安全路径、工具定义、分发
// ═══════════════════════════════════════════════════════════

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

// allTools 返回父 agent 的全部 7 个工具：s05 的 6 个 + task。
func allTools() []ai.ToolUnionParam {
	return []ai.ToolUnionParam{
		bashToolDef(),
		readFileToolDef(),
		writeFileToolDef(),
		editFileToolDef(),
		globToolDef(),
		todoWriteToolDef(),
		taskToolDef(),
	}
}

// subAgentTools 返回子 agent 的工具集：5 个基础工具，不含 todo_write 也不含 task。
// 不含 task 是为了防止递归派生；不含 todo_write 是因为子 agent 只需完成任务、返回摘要。
func subAgentTools() []ai.ToolUnionParam {
	return []ai.ToolUnionParam{
		bashToolDef(),
		readFileToolDef(),
		writeFileToolDef(),
		editFileToolDef(),
		globToolDef(),
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

// taskToolDef 是 s06 新增的父 agent 工具：派生子 agent 处理复杂子任务。
// 只返回最终结论，子 agent 的中间历史全部丢弃。
func taskToolDef() ai.ToolUnionParam {
	return ai.ToolUnionParam{OfTool: &ai.ToolParam{
		Name:        "task",
		Description: ai.String("Launch a subagent to handle a complex subtask. Returns only the final conclusion."),
		InputSchema: ai.ToolInputSchemaParam{
			Properties: map[string]any{
				"description": map[string]any{"type": "string"},
			},
			Required: []string{"description"},
		},
	}}
}

// dispatchTool 按工具名查表执行（不含 task，task 在 agentLoop 里单独处理）。
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
