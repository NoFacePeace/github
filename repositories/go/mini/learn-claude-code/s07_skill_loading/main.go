package main

// s07: Skill Loading — 两级按需知识注入。
//
//	Layer 1（便宜，常驻）:
//	  SYSTEM prompt 包含技能名 + 一行描述（~100 tokens/skill）
//	  "Skills available: agent-builder, code-review, mcp-builder, pdf"
//
//	Layer 2（贵，按需）:
//	  Agent 调用 load_skill("code-review") → 完整 SKILL.md 内容
//	  经 tool_result 注入（~2000 tokens/skill）
//
//	skills/
//	  agent-builder/SKILL.md
//	  code-review/SKILL.md
//	  mcp-builder/SKILL.md
//	  pdf/SKILL.md
//
// 相比 s06 的变化:
//   + buildSystem() — 启动时扫描 skills/ 目录，把目录注入 SYSTEM
//   + loadSkill(name) — 返回完整 SKILL.md 内容（经 tool_result）
//   + SKILLS_DIR 配置
//   循环不变：load_skill 经 dispatchTool 自动分发。
//
// 安全要点：load_skill 只查注册表（白名单），不从路径加载——
// 无路径穿越风险（不像 read_file）。即便模型传 "../../etc/passwd"
// 也只会在注册表里查不到，返回 "Skill not found"。
//
// 运行:
//
//	export ANTHROPIC_API_KEY=...
//	export MODEL_ID=...
//	# 可选：走自建代理或第三方兼容网关时设置
//	export ANTHROPIC_BASE_URL=https://your-proxy.example.com/
//	go run ./mini/learn-claude-code/s07_skill_loading

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	ai "github.com/anthropics/anthropic-sdk-go"
	aiopt "github.com/anthropics/anthropic-sdk-go/option"
	yaml "gopkg.in/yaml.v3"
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
	subMaxTurns   = 30
	// skillsDirName 是技能目录名（相对工作目录），对齐 Python 的 SKILLS_DIR = WORKDIR / "skills"。
	skillsDirName = "skills"
	// skillFileName 是技能清单文件名，对齐 Python 的 SKILL.md。
	skillFileName = "SKILL.md"
)

var stdinReader = bufio.NewReader(os.Stdin)

var currentTodos []todoItem

type todoItem struct {
	Content string `json:"content"`
	Status  string `json:"status"`
}

// skillEntry 是注册表里的一条技能：name/description 取自 frontmatter，
// content 是 SKILL.md 原文（含 frontmatter，原样回灌给模型）。
type skillEntry struct {
	name        string
	description string
	content     string
}

// skillRegistry 是启动时扫描 skills/ 目录建立的注册表。
// loadSkill 只查这张表（白名单），不读路径——这是防穿越的关键。
var skillRegistry = map[string]skillEntry{}

func main() {
	fmt.Println("s07: Skill Loading — catalog in SYSTEM, content on demand")
	fmt.Println("Type a question, press Enter. Type q to quit.")
	fmt.Println()

	workdir, err := os.Getwd()
	if err != nil {
		fmt.Printf("getwd error: %s\n", err)
		return
	}

	// s07: 启动时扫描 skills/ 目录，建立注册表。目录不存在则注册表为空。
	scanSkills(filepath.Join(workdir, skillsDirName))

	// s07: SYSTEM 注入技能目录（便宜层：只有名字 + 一行描述）。
	system := buildSystem(workdir)

	var sdkOpts []aiopt.RequestOption
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		sdkOpts = append(sdkOpts, aiopt.WithAPIKey(key))
	}
	if baseURL := os.Getenv("ANTHROPIC_BASE_URL"); baseURL != "" {
		sdkOpts = append(sdkOpts, aiopt.WithBaseURL(baseURL))
	}
	client := ai.NewClient(sdkOpts...)

	model := os.Getenv("MODEL_ID")
	// 父 agent 工具集 = s06 的 7 个 + load_skill。
	tools := allTools()
	subTools := subAgentTools()

	registerHooks()

	// 子 agent 不加载技能（无 load_skill 工具，system 也不含目录）。
	subSystem := fmt.Sprintf(
		"You are a coding agent at %s. "+
			"Complete the task you were given, then return a concise summary. "+
			"Do not delegate further.",
		workdir,
	)

	history := []ai.MessageParam{}
	roundsSinceTodo := 0
	for {
		fmt.Printf("%ss07 >> %s", colorCyan, colorReset)
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

// agentLoop 运行 模型→工具调用→工具结果→模型 的循环。与 s06 一致。
func agentLoop(ctx context.Context, client ai.Client, model, system string, tools []ai.ToolUnionParam, history *[]ai.MessageParam, roundsSinceTodo *int, subSystem string, subTools []ai.ToolUnionParam) (string, error) {
	for round := 1; ; round++ {
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

			// task 工具派生子 agent，需要 client/model/subSystem/subTools，单独处理。
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
//  NEW in s07: 技能注册表 + SYSTEM 目录构建 + 按需加载
// ═══════════════════════════════════════════════════════════

// skillFrontmatter 是 SKILL.md 头部 YAML frontmatter 里只取这两个字段。
// description 在源文件里可能是标量也可能是块标量（| 多行），yaml 库统一解码成 string。
type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// parseFrontmatter 解析 SKILL.md 的 YAML frontmatter，返回 (meta, body)。
// frontmatter 由首行 `---` 包裹；无 frontmatter 则 meta 为空、body 为原文。
// 对齐 Python 版 _parse_frontmatter。
func parseFrontmatter(text string) (skillFrontmatter, string) {
	var meta skillFrontmatter
	// 必须以 `---` 开头才算有 frontmatter。
	if !strings.HasPrefix(text, "---") {
		return meta, text
	}
	rest := strings.TrimPrefix(text, "---")
	// 找闭合的 `---`。frontmatter 体内本身不会再出现独占一行的 `---`。
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return meta, text
	}
	header := rest[:end]
	// body 是闭合分隔之后的部分，去掉前导换行。
	body := strings.TrimLeft(rest[end+4:], "\n")
	// 解析 frontmatter；失败则 meta 留空，body 仍可用。
	if err := yaml.Unmarshal([]byte(header), &meta); err != nil {
		// 解析失败：按 Python 行为，meta 留空，返回原文 body。
		return skillFrontmatter{}, text
	}
	return meta, body
}

// scanSkills 扫描 skillsDir 下的 `<name>/SKILL.md`，填充 skillRegistry。
// 对齐 Python 版 _scan_skills：目录不存在直接返回；按子目录名排序保证目录顺序稳定。
func scanSkills(skillsDir string) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		// 目录不存在或不可读：注册表保持空。
		return
	}
	// 排序保证 system prompt 里目录顺序稳定（os.ReadDir 本身已按名排序，这里显式一下）。
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		manifest := filepath.Join(skillsDir, e.Name(), skillFileName)
		raw, err := os.ReadFile(manifest)
		if err != nil {
			continue
		}
		text := string(raw)
		meta, _ := parseFrontmatter(text)
		name := meta.Name
		if name == "" {
			name = e.Name()
		}
		desc := meta.Description
		if desc == "" {
			// 回退：取正文第一行去掉前导 # 作为描述，对齐 Python 版的回退逻辑。
			firstLine := strings.SplitN(strings.TrimLeft(text, " \t\r\n"), "\n", 2)[0]
			desc = strings.TrimLeft(firstLine, "# ")
		}
		skillRegistry[name] = skillEntry{
			name:        name,
			description: desc,
			content:     text,
		}
	}
}

// listSkills 列出所有技能（名字 + 一行描述），用于注入 SYSTEM prompt。
func listSkills() string {
	if len(skillRegistry) == 0 {
		return "(no skills found)"
	}
	// 按名字排序，保证 SYSTEM 目录顺序稳定。
	names := make([]string, 0, len(skillRegistry))
	for n := range skillRegistry {
		names = append(names, n)
	}
	sort.Strings(names)
	var sb strings.Builder
	for i, n := range names {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("- **%s**: %s", skillRegistry[n].name, skillRegistry[n].description))
	}
	return sb.String()
}

// buildSystem 构造父 agent 的 SYSTEM prompt：工作目录 + 技能目录 + 加载提示。
// 这是 Layer 1（便宜层），只放名字和描述。
func buildSystem(workdir string) string {
	catalog := listSkills()
	return fmt.Sprintf(
		"You are a coding agent at %s. "+
			"Skills available:\n%s\n"+
			"Use load_skill to get full details when needed.",
		workdir, catalog,
	)
}

// loadSkill 按名字加载技能完整内容。只查注册表，不读路径——防穿越。
// 对齐 Python 版 load_skill。
func loadSkill(name string) string {
	skill, ok := skillRegistry[name]
	if !ok {
		return fmt.Sprintf("Skill not found: %s", name)
	}
	return skill.content
}

// ═══════════════════════════════════════════════════════════
//  FROM s06: 子 agent
// ═══════════════════════════════════════════════════════════

func dispatchTask(ctx context.Context, client ai.Client, model, subSystem string, subTools []ai.ToolUnionParam, input json.RawMessage) string {
	var in struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return spawnSubagent(ctx, client, model, subSystem, subTools, in.Description)
}

// spawnSubagent 派生子 agent：全新 messages[]、最多 subMaxTurns 轮、跑 hook、
// 只返回最终摘要。对齐 Python 版 spawn_subagent。
func spawnSubagent(ctx context.Context, client ai.Client, model, subSystem string, subTools []ai.ToolUnionParam, description string) string {
	fmt.Printf("\n%s[Subagent spawned]%s\n", colorPurple, colorReset)
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
			fmt.Printf("%s[Subagent error]%s %s\n", colorPurple, colorReset, err)
			return fmt.Sprintf("Subagent error: %s", err)
		}
		subHistory = append(subHistory, resp.ToParam())

		if resp.StopReason != ai.StopReasonToolUse {
			stoppedNormally = true
			break
		}

		var results []ai.ContentBlockParamUnion
		for _, block := range resp.Content {
			tu, ok := block.AsAny().(ai.ToolUseBlock)
			if !ok {
				continue
			}
			blocked := triggerHooks(hookPreToolUse, tu.Name, tu.Input)
			if blocked != "" {
				results = append(results, ai.NewToolResultBlock(tu.ID, blocked, false))
				continue
			}
			out := dispatchTool(ctx, tu.Name, tu.Input)
			triggerHooks(hookPostToolUse, tu.Name, tu.Input, out)
			preview := out
			if len(preview) > 100 {
				preview = preview[:100]
			}
			fmt.Printf("  %s[sub] %s: %s%s\n", colorGray, tu.Name, preview, colorReset)
			results = append(results, ai.NewToolResultBlock(tu.ID, out, false))
		}
		subHistory = append(subHistory, ai.NewUserMessage(results...))
	}

	// 摘要提取 + 超限回退，对齐 Python 版。
	result := ""
	if stoppedNormally && len(subHistory) > 0 {
		result = extractTextFromMessage(subHistory[len(subHistory)-1])
	}
	if result == "" {
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
//  FROM s02-s06 (unchanged): 工具实现、安全路径、工具定义、分发
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

// allTools 返回父 agent 的全部 8 个工具：s06 的 7 个 + load_skill。
func allTools() []ai.ToolUnionParam {
	return []ai.ToolUnionParam{
		bashToolDef(),
		readFileToolDef(),
		writeFileToolDef(),
		editFileToolDef(),
		globToolDef(),
		todoWriteToolDef(),
		taskToolDef(),
		loadSkillToolDef(),
	}
}

// subAgentTools 返回子 agent 的工具集：5 个基础工具，不含 task（防递归）、
// 不含 load_skill（子 agent 不加载技能）、不含 todo_write。
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

// loadSkillToolDef 是 s07 新增工具：按名字加载技能完整内容。
// 目录已在 SYSTEM prompt 里（便宜层），这个工具是 Layer 2（贵层，按需）。
func loadSkillToolDef() ai.ToolUnionParam {
	return ai.ToolUnionParam{OfTool: &ai.ToolParam{
		Name:        "load_skill",
		Description: ai.String("Load the full content of a skill by name."),
		InputSchema: ai.ToolInputSchemaParam{
			Properties: map[string]any{
				"name": map[string]any{"type": "string"},
			},
			Required: []string{"name"},
		},
	}}
}

// dispatchTool 按工具名查表执行（task 在 agentLoop 里单独处理）。
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
	case "load_skill":
		var in struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return fmt.Sprintf("Error: %s", err)
		}
		return loadSkill(in.Name)
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
