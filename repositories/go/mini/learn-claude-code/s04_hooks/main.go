package main

// s04: Hooks — 把扩展逻辑从循环体里搬出去，挂到 hook 上。
//
//	User types query
//	     │
//	     ▼
//	┌──────────────────┐
//	│ UserPromptSubmit │ ── triggerHooks() before LLM
//	└────────┬─────────┘
//	         ▼
//	┌────────────┐     ┌─────────────────────────────┐
//	│  messages  │────▶│  LLM (stop_reason=tool_use?)│
//	└────────────┘     │   No ──▶ Stop hooks ──▶ exit │
//	                 │   Yes ──▶ tool_use block ──┐ │
//	                 └────────────────────────────┘ │
//	                                              ▼
//	                                    ┌──────────────────┐
//	                                    │ triggerHooks()   │
//	                                    │  PreToolUse:     │
//	                                    │   permissionHook  │
//	                                    │   logHook         │
//	                                    └───────┬──────────┘
//	                                            │ (not blocked)
//	                                    ┌───────▼──────────┐
//	                                    │ dispatchTool[x]    │
//	                                    └───────┬──────────┘
//	                                            │
//	                                    ┌───────▼──────────┐
//	                                    │ triggerHooks()   │
//	                                    │  PostToolUse:    │
//	                                    │   largeOutputHook │
//	                                    └───────┬──────────┘
//	                                            │
//	                                    results ──▶ back to messages
//
// 相比 s03 的变化:
//   + HOOKS 注册表（event -> []callback）
//   + registerHook / triggerHooks
//   + contextInjectHook (UserPromptSubmit)
//   + permissionHook、logHook (PreToolUse)
//   + largeOutputHook (PostToolUse)
//   + summaryHook (Stop)
//   - 循环体里的 checkPermission 删除（逻辑搬进 permissionHook，经 PreToolUse 触发）
//
// 运行:
//
//	export ANTHROPIC_API_KEY=...
//	export MODEL_ID=...
//	# 可选：走自建代理或第三方兼容网关时设置
//	export ANTHROPIC_BASE_URL=https://your-proxy.example.com/
//	go run ./mini/learn-claude-code/s04_hooks

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
	// 青色：工具名/提示符；黄色：Gate2 警告/大输出警告；红色：Gate1 硬拒；
	// 灰色：hook 日志。
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorGray   = "\033[90m"
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
	// PostToolUse 大输出警告阈值，对齐 Python 版的 100000。
	largeOutputThreshold = 100000
)

// stdinReader 贯穿 REPL 输入与 permissionHook 的 y/N 询问。
// 用 *bufio.Reader 而非 bufio.Scanner，因为权限询问要按行读，
// 而且不能和 Scanner 各自缓冲 stdin —— 混用会互相抢字节。
var stdinReader = bufio.NewReader(os.Stdin)

func main() {
	// 启动横幅与交互提示。
	fmt.Println("s04: Hooks — extension logic on hooks, loop stays clean")
	fmt.Println("Type a question, press Enter. Type q to quit.")
	fmt.Println()

	// 把当前工作目录作为工作区根：所有 read/write/edit/glob 都限制在它下面。
	// 路径直接注入 system prompt，让模型知道自己“身处哪个项目根”。
	workdir, err := os.Getwd()
	if err != nil {
		fmt.Printf("getwd error: %s\n", err)
		return
	}
	// s04 system prompt改回 s02 的“行动而非解释”（权限由 hook 处理，不在 prompt 里声明）。
	system := fmt.Sprintf("You are a coding agent at %s. Use tools to solve tasks. Act, don't explain.", workdir)

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
	// 工具集与 s02/s03 一致：bash + read_file + write_file + edit_file + glob。
	tools := allTools()

	// 注册全部 hook（顺序即触发顺序）。permissionHook 必须在 logHook 之前，
	// 这样被拒的调用不会产生“即将执行”日志。
	registerHooks()

	// history 保存整轮对话（user / assistant / tool_result），跨多轮 REPL 保持。
	history := []ai.MessageParam{}
	for {
		// 打印提示符（青色），读取一行用户输入。
		fmt.Printf("%ss04 >> %s", colorCyan, colorReset)
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			break
		}
		query := strings.TrimSpace(line)
		// 空、q、exit 都视为退出信号。
		if query == "" || strings.EqualFold(query, "q") || strings.EqualFold(query, "exit") {
			if query == "" {
				continue
			}
			break
		}

		// UserPromptSubmit hook：用户输入进 LLM 前触发。
		triggerHooks(hookUserPromptSubmit, query)

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
// 结构与 s03 一致；唯一差别是 s03 的 checkPermission 被替换为 triggerHooks(PreToolUse)，
// 并且新增 triggerHooks(PostToolUse) 与 Stop hook。
func agentLoop(ctx context.Context, client ai.Client, model, system string, tools []ai.ToolUnionParam, history *[]ai.MessageParam) (string, error) {
	for round := 1; ; round++ {
		// 组装每次请求的参数：模型、上限、历史消息、可用工具。
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

		// stop_reason 不是 tool_use：模型决定停下。触发 Stop hook，
		// 若 hook 返回 force 字符串，把它作为 user 消息塞回历史并继续（不直接 return），
		// 让模型在 force 引导下继续工作。这是 s04 的特殊语义。
		if resp.StopReason != ai.StopReasonToolUse {
			force := triggerHooks(hookStop, *history)
			if force != "" {
				*history = append(*history, ai.NewUserMessage(ai.NewTextBlock(force)))
				continue
			}
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
			// 打印模型要调用的工具名（青色）。
			fmt.Printf("%s> %s%s\n", colorCyan, tu.Name, colorReset)

			// s04 改动：s03 的 checkPermission 换成 PreToolUse hook。
			// 任一 hook 返回非空字符串即视为“阻止”，把该字符串作为 tool_result 回灌。
			blocked := triggerHooks(hookPreToolUse, tu.Name, tu.Input)
			if blocked != "" {
				results = append(results, ai.NewToolResultBlock(tu.ID, blocked, false))
				continue
			}

			out := dispatchTool(ctx, tu.Name, tu.Input)

			// s04 新增：PostToolUse hook（大输出警告等）。返回值忽略。
			triggerHooks(hookPostToolUse, tu.Name, tu.Input, out)

			// 把完整输出（截断后）作为 tool_result 回灌。截断对齐 Python 版的 out[:50000]。
			results = append(results, ai.NewToolResultBlock(tu.ID, out, false))
		}
		// 所有工具结果汇成一条 user 消息，进入下一轮请求。
		*history = append(*history, ai.NewUserMessage(results...))
	}
}

// ═══════════════════════════════════════════════════════════
//  NEW in s04: Hook 系统（s03 的权限逻辑现在挂在 hook 上）
// ═══════════════════════════════════════════════════════════

// hookEvent 标识 hook 触发点。
type hookEvent string

const (
	hookUserPromptSubmit hookEvent = "UserPromptSubmit"
	hookPreToolUse       hookEvent = "PreToolUse"
	hookPostToolUse      hookEvent = "PostToolUse"
	hookStop             hookEvent = "Stop"
)

// hookFunc 是 hook 回调的统一签名。返回非空字符串表示“阻止/干预”
// （PreToolUse 用来阻止工具执行；Stop 用来强制继续）。
//
// UserPromptSubmit 或 PostToolUse 的 hook 返回值会被忽略。
type hookFunc func(args ...any) string

// hooks 是 event -> 回调列表的注册表。Python 用 dict of list，Go 用 map。
var hooks = map[hookEvent][]hookFunc{
	hookUserPromptSubmit: {},
	hookPreToolUse:       {},
	hookPostToolUse:      {},
	hookStop:             {},
}

// registerHook 把 callback 追加到 event 的 hook 列表末尾。顺序即触发顺序。
func registerHook(event hookEvent, cb hookFunc) {
	hooks[event] = append(hooks[event], cb)
}

// triggerHooks 按 event 依次调用已注册 hook。
//
// 教学版语义（对齐 Python）：任一 hook 返回非空字符串即短路返回该字符串；
// 否则返回空串。这里 PreToolUse 用它阻止工具执行，Stop 用它强制 continue。
//
// 注意：PostToolUse / UserPromptSubmit 的 hook 返回值按约定应被忽略，
// 但教学版统一用了 triggerHooks，所以这些 hook 内部应返回 ""。
func triggerHooks(event hookEvent, args ...any) string {
	for _, cb := range hooks[event] {
		if result := cb(args...); result != "" {
			return result
		}
	}
	return ""
}

// ── hook 入参解包辅助 ───────────────────────────────────────

// hookArgs 解包 hook 收到的变参：toolName + input（PreToolUse/PostToolUse）。
// PostToolUse 额外多一个 output 参数。
func hookArgs(args []any) (name string, input map[string]any) {
	if len(args) >= 1 {
		name, _ = args[0].(string)
	}
	if len(args) >= 2 {
		input, _ = args[1].(map[string]any)
	}
	return
}

// ── s03 权限逻辑，现在包成 hook ─────────────────────────────

// denyList 是 bash 硬拒绝列表，对齐 Python s04（已去掉 "> /dev/sda"，
// 与 s04 源码 DENY_LIST 一致）。命中直接拒，不询问。
var denyList = []string{"rm -rf /", "sudo", "shutdown", "reboot", "mkfs", "dd if="}

// destructive 是 bash 破坏性关键字，命中触发用户确认。
var destructive = []string{"rm ", "> /etc/", "chmod 777"}

// permissionHook 是 PreToolUse hook：把 s03 的三道闸门逻辑搬进来。
// 返回非空字符串 = 阻止工具执行（作为 tool_result 回灌给模型）。
func permissionHook(args ...any) string {
	name, input := hookArgs(args)
	if name == "" {
		return ""
	}

	if name == "bash" {
		cmd, _ := input["command"].(string)
		// Gate1：硬拒。
		for _, pattern := range denyList {
			if strings.Contains(cmd, pattern) {
				fmt.Printf("\n%s⛔ Blocked: '%s'%s\n", colorRed, pattern, colorReset)
				return "Permission denied by deny list"
			}
		}
		// Gate2+3：破坏性命令询问 y/N。
		for _, kw := range destructive {
			if strings.Contains(cmd, kw) {
				fmt.Printf("\n%s⚠  Potentially destructive command%s\n", colorYellow, colorReset)
				printCall(name, input)
				if !askUser() {
					return "Permission denied by user"
				}
				break
			}
		}
	}
	if name == "write_file" || name == "edit_file" {
		p, _ := input["path"].(string)
		// 写工作区外：触发询问。safePath 已会硬拒，这里按 Python 版语义做成“可询问”。
		if !isWithinWorkspace(p) {
			fmt.Printf("\n%s⚠  Writing outside workspace%s\n", colorYellow, colorReset)
			printCall(name, input)
			if !askUser() {
				return "Permission denied by user"
			}
		}
	}
	return ""
}

// logHook 是 PreToolUse hook：把每次工具调用打印一行灰色日志。
// 在 permissionHook 之后注册，所以被拒的调用不会进到这里。
func logHook(args ...any) string {
	name, input := hookArgs(args)
	// 预览入参前两个 value，对齐 Python 版 str(list(input.values())[:2])[:60]。
	preview := previewInput(input, 2, 60)
	fmt.Printf("%s[HOOK] %s(%s)%s\n", colorGray, name, preview, colorReset)
	return ""
}

// largeOutputHook 是 PostToolUse hook：输出超阈值时黄色警告。
func largeOutputHook(args ...any) string {
	name, _ := hookArgs(args)
	output := ""
	if len(args) >= 3 {
		output, _ = args[2].(string)
	}
	if len(output) > largeOutputThreshold {
		fmt.Printf("%s[HOOK] ⚠ Large output from %s: %d chars%s\n", colorGray, name, len(output), colorReset)
	}
	return ""
}

// contextInjectHook 是 UserPromptSubmit hook：用户输入进 LLM 前打印工作目录。
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
	// 统计 user 消息里 tool_result 块的数量（每个 tool_result 对应一次工具调用）。
	// MessageParam.Content 是 []ContentBlockParamUnion，用 OfToolResult 字段判别。
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

// registerHooks 注册全部 hook（顺序即触发顺序）。
// permissionHook 必须在 logHook 之前，避免被拒的调用也打“即将执行”日志。
func registerHooks() {
	registerHook(hookUserPromptSubmit, contextInjectHook)
	registerHook(hookPreToolUse, permissionHook)
	registerHook(hookPreToolUse, logHook)
	registerHook(hookPostToolUse, largeOutputHook)
	registerHook(hookStop, summaryHook)
}

// ── hook 辅助函数 ──────────────────────────────────────────

// printCall 打印将要执行的工具及其参数，用于权限询问。
func printCall(name string, input map[string]any) {
	argsJSON, _ := json.Marshal(input)
	fmt.Printf("   Tool: %s(%s)\n", name, string(argsJSON))
}

// askUser 读一行 y/N，返回是否放行。对齐 Python: choice in ("y","yes")。
func askUser() bool {
	fmt.Printf("   Allow? [y/N] ")
	line, err := stdinReader.ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.TrimSpace(strings.ToLower(line))
	return answer == "y" || answer == "yes"
}

// previewInput 把 input 的前 N 个 value 拼成短预览，用于日志。
func previewInput(input map[string]any, n, maxLen int) string {
	if len(input) == 0 {
		return ""
	}
	// map 顺序不确定，取前 N 个 value 用 JSON 装 list 形式预览。
	vals := make([]any, 0, n)
	for _, v := range input {
		vals = append(vals, v)
		if len(vals) >= n {
			break
		}
	}
	b, _ := json.Marshal(vals)
	s := string(b)
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	return s
}

// ═══════════════════════════════════════════════════════════
//  FROM s02-s03 (unchanged): 工具实现、安全路径、工具定义、分发
// ═══════════════════════════════════════════════════════════

// safePath 把相对路径解析到工作目录下，并拒绝任何逃逸出工作目录的路径
// （含 `..` 越界、绝对路径指向外部、符号链接指向外部等）。所有文件类工具都先过它。
func safePath(p string) (string, error) {
	workdir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Symlink-aware 规范化：先 EvalSymlinks 解开工作目录里的符号链接，
	// 再解析目标路径。这样做是为了挡住“工作目录内符号链接指向外部”的逃逸。
	workdirResolved, err := filepath.EvalSymlinks(workdir)
	if err != nil {
		// 目录本身无符号链接时直接用原路径。
		workdirResolved = workdir
	}
	// 关键：绝对路径不要用 filepath.Join 拼——Join 会剥掉前导分隔符再接到
	// workdir 后面，导致 `/workdir + /workdir/foo` 这样的怪路径恰好以 workdir
	// 开头而绕过前缀检查。绝对路径直接取其本身（再规范化）。
	var abs string
	if filepath.IsAbs(p) {
		abs = filepath.Clean(p)
	} else {
		abs = filepath.Join(workdirResolved, p)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		// 目标文件尚不存在（write 场景）：退到对父目录做规范化校验。
		parentResolved, perr := filepath.EvalSymlinks(filepath.Dir(abs))
		if perr != nil {
			parentResolved = filepath.Dir(abs)
		}
		// 用父目录 + 文件名重建一个用于 prefix 比较的候选路径。
		resolved = filepath.Join(parentResolved, filepath.Base(abs))
	}
	// 规范化工作目录用于 prefix 比较（补上结尾分隔符，避免 `/foo` 误判 `/foobar` 合法）。
	workdirPrefix := filepath.Clean(workdirResolved) + string(os.PathSeparator)
	if !strings.HasPrefix(resolved+string(os.PathSeparator), workdirPrefix) {
		return "", fmt.Errorf("Path escapes workspace: %s", p)
	}
	return resolved, nil
}

// isWithinWorkspace 判断路径解析后是否仍在工作区内，用于 permissionHook。
// 与 safePath 共用解析逻辑，但只返回布尔，不报错。
func isWithinWorkspace(p string) bool {
	_, err := safePath(p)
	return err == nil
}

// runBash 执行一条 shell 命令，对齐 Python 版的超时与截断。
// s04 注意：危险命令黑名单已上移到 permissionHook（PreToolUse），runBash 自身不再检查。
func runBash(ctx context.Context, command string) string {
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

// runRead 读取文件内容。limit 非零时只读前 limit 行，并附上“还有多少行”的提示。
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

// runWrite 把 content 写入文件，必要时创建父目录。返回写入字节数。
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

// runEdit 把文件里第一处 old_text 替换成 new_text。找不到 old_text 即报错。
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
	// 用 strings.Replace 限制 1 次，对齐 Python 的 str.replace(old, new, 1)。
	if !strings.Contains(text, oldText) {
		return fmt.Sprintf("Error: text not found in %s", path)
	}
	edited := strings.Replace(text, oldText, newText, 1)
	if err := os.WriteFile(abs, []byte(edited), 0o644); err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return fmt.Sprintf("Edited %s", path)
}

// runGlob 按通配符在工作目录下匹配文件，结果过滤掉任何逃逸工作目录的匹配
// （对齐 Python 版用 is_relative_to(WORKDIR) 的收口）。
func runGlob(pattern string) string {
	workdir, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	matches, err := globMatch(workdir, pattern)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	// 过滤：规范化后必须仍在工作目录前缀下（挡住 `..` 之类的匹配）。
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
		// 返回相对工作目录的路径，与 Python glob(root_dir=WORKDIR) 行为一致。
		if rel, err := filepath.Rel(workdir, m); err == nil {
			results = append(results, rel)
		}
	}
	if len(results) == 0 {
		return "(no matches)"
	}
	return strings.Join(results, "\n")
}

// globMatch 在 root 下按 pattern 匹配，支持 ** 递归。
// 未含 ** 时退回 filepath.Glob（绝对路径匹配）。
func globMatch(root, pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		// 普通通配符：把模式拼到 root 下交给 filepath.Glob。
		return filepath.Glob(filepath.Join(root, pattern))
	}
	// ** 支持：遍历 root 下所有文件，再用 matchSegs 判断是否匹配。
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

// splitSegments 把模式/路径按分隔符切成段，并丢弃空段。
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

// matchSegs 判断路径段 path 是否匹配模式段 pat（支持 ** 跨多段、* 单段通配）。
func matchSegs(path, pat []string) bool {
	// 双指针递归：** 可匹配 0 个或多个连续段。
	for len(pat) > 0 {
		if pat[0] == "**" {
			// ** 匹配 0..len(path) 段，逐个尝试。
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
		// 普通段：用 filepath.Match 做单段匹配。
		ok, err := filepath.Match(pat[0], path[0])
		if err != nil || !ok {
			return false
		}
		path = path[1:]
		pat = pat[1:]
	}
	return len(path) == 0
}

// allTools 返回全部 5 个工具的定义，与 s02/s03 一致。
func allTools() []ai.ToolUnionParam {
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

// dispatchTool 按工具名查表执行，找不到则返回 Unknown。
// 对应 Python 的: output = TOOL_HANDLERS[block.name](**block.input)。
// 入参 input 是模型给出的 JSON，按工具名解析出对应字段后调用处理函数。
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
	default:
		return fmt.Sprintf("Unknown: %s", name)
	}
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

// truncate 把字符串截断到 n 字符，用于限制工具输出长度，避免超长输出
// 撑爆请求 token 上限。
func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
