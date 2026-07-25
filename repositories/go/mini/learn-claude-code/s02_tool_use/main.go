package main

// s02: Tool Use — 在 s01 基础上新增 4 个工具 + 分发映射。
//
// 本文件 = s01 的全部代码 + 以下新增:
//   + runRead / runWrite / runEdit / runGlob 四个工具实现
//   + 工具分发映射（替代 s01 中硬编码的 runBash 调用，改成按名字查表）
//   + safePath 路径安全校验（read/write/edit/glob 都先把相对路径解析到
//     工作目录下，并拒绝任何逃逸出工作目录的路径）
//
// 循环本身（agentLoop）结构与 s01 一致，唯一差别是工具执行那段：
//
//	s01: output = runBash(block.input["command"])
//	s02: output = TOOL_HANDLERS[block.name](**block.input)
//
// 运行:
//
//	export ANTHROPIC_API_KEY=...
//	export MODEL_ID=...
//	# 可选：走自建代理或第三方兼容网关时设置
//	export ANTHROPIC_BASE_URL=https://your-proxy.example.com/
//	go run ./mini/learn-claude-code/s02_tool_use

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
	// ANSI 颜色：用于终端输出着色（工具名/命令黄色、提示符青色）。
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorReset   = "\033[0m"

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
	fmt.Println("s02: Tool Use — 在 s01 基础上加了 4 个工具")
	fmt.Println("输入问题，回车发送。输入 q 退出。")
	fmt.Println()

	// 把当前工作目录作为工作区根：所有 read/write/edit/glob 都限制在它下面。
	// 路径直接注入 system prompt，让模型知道自己“身处哪个项目根”。
	workdir, err := os.Getwd()
	if err != nil {
		fmt.Printf("getwd error: %s\n", err)
		return
	}
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
	// s02 把工具从 1 个扩展到 5 个：bash + read_file + write_file + edit_file + glob。
	tools := allTools()

	// history 保存整轮对话（user / assistant / tool_result），跨多轮 REPL 保持。
	history := []ai.MessageParam{}
	scanner := bufio.NewScanner(os.Stdin)
	// 把 scanner 缓冲扩到 1MB，避免长行（如粘贴大段代码）被截断。
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for {
		// 打印提示符（青色），读取一行用户输入。
		fmt.Printf("%ss02 >> %s", colorCyan, colorReset)
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
// 结构与 s01 一致；唯一差别是工具执行改成分发映射 dispatchTool。
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
			// 打印模型要调用的工具名（黄色），方便用户观察 agent 在干什么。
			fmt.Printf("%s> %s%s\n", colorYellow, tu.Name, colorReset)
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
//  FROM s01 (unchanged)
// ═══════════════════════════════════════════════════════════

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

// ═══════════════════════════════════════════════════════════
//  NEW in s02: 安全部径与 4 个新工具
// ═══════════════════════════════════════════════════════════

// safePath 把相对路径解析到工作目录下，并拒绝任何逃逸出工作目录的路径
// （含 `..` 越界、绝对路径、符号链接指向外部等）。所有文件类工具都先过它。
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
	abs := filepath.Join(workdirResolved, p)
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
	// 双星号递归匹配 + 一般通配符，filepath.Glob 不支持 **，这里手动处理。
	// 为保持简单与 Python glob 一致，这里用 filepath.Glob（支持 * ? [..]），
	// 并对 ** 提供一个回落：把 ** 当作递归目录展开。
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
	// ** 支持：用双指针递归遍历目录树，逐路径段匹配。
	// 这里用最简实现：遍历 root 下所有文件，再用 matchSegs 判断是否匹配。
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

// ═══════════════════════════════════════════════════════════
//  NEW in s02: 工具定义（s01 只有一个 bash，现在扩展到 5 个）
// ═══════════════════════════════════════════════════════════

// allTools 返回全部 5 个工具的定义，对齐 Python 版的 TOOLS 列表。
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

// ═══════════════════════════════════════════════════════════
//  NEW in s02: 工具分发映射（s01 是硬编码 runBash，现在改为查表）
// ═══════════════════════════════════════════════════════════

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
