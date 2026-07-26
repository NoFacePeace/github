package main

// s03: Permission System — 在 s02（多工具 + 分发）基础上，于工具执行前插入
// 三道闸门的权限流水线。
//
//	+-------+    +--------+    +--------+    +--------+    +------+
//	| Tool  | -> | Gate 1 | -> | Gate 2 | -> | Gate 3 | -> | Exec |
//	| call  |    | deny?  |    | match? |    | allow? |    |      |
//	+-------+    +--------+    +--------+    +--------+    +------+
//	     |            |             |             |
//	     v            v             v             v
//	  (normal)     (blocked)    (ask user)   (user says no?)
//
//	Gate 1: 硬拒绝列表（rm -rf /、sudo、...）—— 直接拒，不询问
//	Gate 2: 规则匹配（写工作区外？破坏性 bash？）—— 触发询问
//	Gate 3: 用户确认（暂停等待 y/N）—— 拒则跳过执行
//
// 相比 s02，agent loop 只在工具执行前多一行:
//
//	if !checkPermission(block) { continue }
//
// 同时把原本在 runBash 里的危险命令黑名单上移到 Gate 1，runBash 自身不再检查。
// system prompt 改为 "All destructive operations require user approval."
//
// 运行:
//
//	export ANTHROPIC_API_KEY=...
//	export MODEL_ID=...
//	# 可选：走自建代理或第三方兼容网关时设置
//	export ANTHROPIC_BASE_URL=https://your-proxy.example.com/
//	go run ./mini/learn-claude-code/s03_permission

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
	// 青色：工具名/提示符；黄色：Gate2 警告；红色：Gate1 硬拒。
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
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

// stdinReader 贯穿 REPL 输入与 Gate3 权限询问。
// 用 *bufio.Reader 而非 bufio.Scanner，因为权限询问要按行读，
// 而且不能和 Scanner 各自缓冲 stdin —— 混用会互相抢字节。
var stdinReader = bufio.NewReader(os.Stdin)

func main() {
	// 启动横幅与交互提示。
	fmt.Println("s03: Permission")
	fmt.Println("输入问题，回车发送。输入 q 退出。")
	fmt.Println()

	// 把当前工作目录作为工作区根：所有 read/write/edit/glob 都限制在它下面。
	// 路径直接注入 system prompt，让模型知道自己“身处哪个项目根”。
	workdir, err := os.Getwd()
	if err != nil {
		fmt.Printf("getwd error: %s\n", err)
		return
	}
	// s03 改提示：所有破坏性操作都需用户确认（对接 Gate3）。
	system := fmt.Sprintf("You are a coding agent at %s. All destructive operations require user approval.", workdir)

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
	// 工具集与 s02 一致：bash + read_file + write_file + edit_file + glob。
	tools := allTools()

	// history 保存整轮对话（user / assistant / tool_result），跨多轮 REPL 保持。
	history := []ai.MessageParam{}
	for {
		// 打印提示符（青色），读取一行用户输入。
		fmt.Printf("%ss03 >> %s", colorCyan, colorReset)
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			break
		}
		query := strings.TrimSpace(line)
		// 空、q、exit 都视为退出信号。
		if query == "" || strings.EqualFold(query, "q") || strings.EqualFold(query, "exit") {
			if query == "" {
				continue // 空行不退出，重新提示
			}
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
// 结构与 s02 一致；唯一差别是工具执行前先过 checkPermission 三道闸门。
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
			// 打印模型要调用的工具名（青色），方便用户观察 agent 在干什么。
			fmt.Printf("%s> %s%s\n", colorCyan, tu.Name, colorReset)

			// s03 改动：执行前先过三道闸门。拒掉的工具不执行，直接回灌
			// "Permission denied."，让模型知道这次调用没被放行。
			if !checkPermission(tu.Name, tu.Input) {
				results = append(results, ai.NewToolResultBlock(tu.ID, "Permission denied.", false))
				continue
			}

			out := dispatchTool(ctx, tu.Name, tu.Input)
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

// ═══════════════════════════════════════════════════════════
//  NEW in s03: 三道闸门权限流水线
// ═══════════════════════════════════════════════════════════

// denyList 是 Gate1 的硬拒绝列表：bash 命令匹配到任意一条就直接拒，不询问。
// 这些是无论如何都不该跑的命令。注意：s02 里这份列表在 runBash 内部，
// s03 把它上移到权限层，runBash 自身不再检查。
var denyList = []string{"rm -rf /", "sudo", "shutdown", "reboot", "mkfs", "dd if=", "> /dev/sda"}

// permissionRule 是 Gate2 的一条规则：作用于某些工具，check 命中时返回提示信息。
type permissionRule struct {
	tools   []string
	check   func(args map[string]any) bool
	message string
}

// permissionRules 是 Gate2 的规则集。
// 规则1：write_file / edit_file 写工作区外 → 触发询问。
// 规则2：bash 含破坏性关键字（rm、> /etc/、chmod 777）→ 触发询问。
// 注意 safePath 已会拒绝工作区外的写，但那是“硬错”；Gate2 把它升级为
// “询问后可放行”的语义，对应 Python 版用 is_relative_to 做的 check。
var permissionRules = []permissionRule{
	{
		tools: []string{"write_file", "edit_file"},
		check: func(args map[string]any) bool {
			p, _ := args["path"].(string)
			// 与 Python 版一致：仅判定是否逃逸工作区，不真正改文件。
			return !isWithinWorkspace(p)
		},
		message: "Writing outside workspace",
	},
	{
		tools: []string{"bash"},
		check: func(args map[string]any) bool {
			cmd, _ := args["command"].(string)
			for _, kw := range []string{"rm ", "> /etc/", "chmod 777"} {
				if strings.Contains(cmd, kw) {
					return true
				}
			}
			return false
		},
		message: "Potentially destructive command",
	},
}

// checkDenyList 是 Gate1：bash 命令匹配 denyList 任一条 → 返回拒绝原因。
func checkDenyList(command string) string {
	for _, pattern := range denyList {
		if strings.Contains(command, pattern) {
			return fmt.Sprintf("Blocked: '%s' is on the deny list", pattern)
		}
	}
	return ""
}

// checkRules 是 Gate2：命中任一规则 → 返回该规则的提示信息。
func checkRules(toolName string, args map[string]any) string {
	for _, rule := range permissionRules {
		for _, t := range rule.tools {
			if t == toolName && rule.check(args) {
				return rule.message
			}
		}
	}
	return ""
}

// askUser 是 Gate3：打印警告，读一行 y/N，返回是否放行。
func askUser(toolName string, args map[string]any, reason string) bool {
	fmt.Printf("\n%s⚠  %s%s\n", colorYellow, reason, colorReset)
	// 把参数 JSON 化打印，方便用户看清将要执行什么。
	argsJSON, _ := json.Marshal(args)
	fmt.Printf("   Tool: %s(%s)\n", toolName, string(argsJSON))
	fmt.Printf("   Allow? [y/N] ")
	line, err := stdinReader.ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.TrimSpace(strings.ToLower(line))
	return answer == "y" || answer == "yes"
}

// checkPermission 跑完整三道闸门，返回是否允许执行该工具调用。
//
//	Gate1（硬拒，bash 专属）：命中 denyList → 红色打印 + 拒绝
//	Gate2（规则匹配）：命中规则 → 进 Gate3 询问
//	Gate3（用户确认）：y 放行 / 其它拒绝
//	未命中任何门 → 默认放行
func checkPermission(toolName string, input json.RawMessage) bool {
	// 先把 input 解成 map，Gate2/Gate3 都要用。
	var args map[string]any
	_ = json.Unmarshal(input, &args)

	// Gate1：硬拒。仅对 bash 检查命令字串。
	if toolName == "bash" {
		if cmd, _ := args["command"].(string); cmd != "" {
			if reason := checkDenyList(cmd); reason != "" {
				fmt.Printf("\n%s⛔ %s%s\n", colorRed, reason, colorReset)
				return false
			}
		}
	}

	// Gate2：规则匹配。命中则进 Gate3 询问。
	if reason := checkRules(toolName, args); reason != "" {
		if !askUser(toolName, args, reason) {
			return false
		}
	}
	return true
}

// ═══════════════════════════════════════════════════════════
//  FROM s02 (unchanged): 工具实现、安全路径、工具定义、分发
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

// isWithinWorkspace 判断路径解析后是否仍在工作区内，用于 Gate2 规则匹配。
// 与 safePath 共用解析逻辑，但只返回布尔，不报错。
func isWithinWorkspace(p string) bool {
	_, err := safePath(p)
	return err == nil
}

// runBash 执行一条 shell 命令，对齐 Python 版的超时与截断。
// s03 注意：危险命令黑名单已上移到 Gate1（checkPermission），runBash 自身不再检查。
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

// allTools 返回全部 5 个工具的定义，与 s02 一致。
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
