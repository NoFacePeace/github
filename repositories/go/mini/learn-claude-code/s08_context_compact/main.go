package main

// s08: Context Compact — 四层压缩流水线 + 应急 reactive compact。
//
//	顺序对齐 CC 源码：L3 budget → L1 snip → L2 micro → [token 超阈值?] → L4 summary
//
//	L1 snip_compact       — 消息数 > 50 时裁剪中间（不切 tool_use/tool_result 对）
//	L2 micro_compact      — 旧 tool_result 替换为占位符
//	L3 tool_result_budget — 大结果落盘（> 30000 字节），保留 2000 字节预览
//	L4 compact_history    — LLM 全量摘要（1 次 API 调用）
//
//	应急 reactive_compact — API 仍报 prompt_too_long 时触发，保留最近 5 条
//
//	┌─────────────────────────────────────────────────────────────┐
//	│  messages[]                                                  │
//	│     ↓                                                        │
//	│  L3 budget ─→ L1 snip ─→ L2 micro ─→ [token > threshold?]    │
//	│                                            ├─ No → LLM       │
//	│                                            └─ Yes → L4 summary│
//	│     ↓                                                        │
//	│  LLM call                                                    │
//	│  [prompt_too_long?]                                          │
//	│     └─ Yes → reactive                                        │
//	└─────────────────────────────────────────────────────────────┘
//
// 核心原则：便宜的先做，贵的最后做。
//
// 相比 s07 的变化:
//   + tool_result_budget — 大结果落盘
//   + snipCompact / microCompact — 内存层压缩
//   + compactHistory — LLM 全量摘要
//   + reactiveCompact — 应急压缩
//   + compact 工具 — 模型主动触发 L4
//   + 模型重试 1 次（reactive）
//
// 运行:
//
//	export ANTHROPIC_API_KEY=...
//	export MODEL_ID=...
//	# 可选：走自建代理或第三方兼容网关时设置
//	export ANTHROPIC_BASE_URL=https://your-proxy.example.com/
//	go run ./mini/learn-claude-code/s08_context_compact

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
	subMaxTurns   = 30
	skillsDirName = "skills"
	skillFileName = "SKILL.md"

	// s08 压缩相关常量（对齐 Python 版）。
	contextLimit       = 50000  // 估算的消息字节阈值；超阈值触发 L4
	keepRecent         = 3      // L2 micro_compact 保留最近 K 个 tool_result
	persistThreshold   = 30000  // L3 单条结果超过此字节数才落盘
	resultBudgetMax    = 200000 // L3 一次请求内 tool_result 累计字节上限
	maxReactiveRetries = 1      // reactive compact 重试上限

	// 落盘目录与转写目录（相对工作目录）。
	toolResultsDirName = ".task_outputs/tool-results"
	transcriptDirName  = ".transcripts"
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
	fmt.Println("s08: Context Compact — four-layer compaction pipeline")
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
	// 父 agent 工具集 = s07 的 8 个 + compact。
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
		fmt.Printf("%ss08 >> %s", colorCyan, colorReset)
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

		finalText, err := agentLoop(context.Background(), client, model, system, tools, &history, &roundsSinceTodo, subSystem, subTools, workdir)
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

// ═══════════════════════════════════════════════════════════
//  agentLoop — s08 核心：每次 LLM 调用前跑压缩流水线
// ═══════════════════════════════════════════════════════════

// agentLoop 在 s07 基础上，把三次内存压缩 + 必要的 L4 摘要提前到 API 调用前。
// 顺序对齐 CC 源码：budget → snip → micro，再判定是否需要 L4；API 报 prompt_too_long
// 触发 reactive compact。重试用 reactiveRetries 计数，受 maxReactiveRetries 限制。
func agentLoop(ctx context.Context, client ai.Client, model, system string, tools []ai.ToolUnionParam, history *[]ai.MessageParam, roundsSinceTodo *int, subSystem string, subTools []ai.ToolUnionParam, workdir string) (string, error) {
	reactiveRetries := 0
	for round := 1; ; round++ {
		if *roundsSinceTodo >= nagRounds && len(*history) > 0 {
			*history = append(*history, ai.NewUserMessage(ai.NewTextBlock("<reminder>Update your todos.</reminder>")))
			*roundsSinceTodo = 0
		}

		// s08: 三段内存压缩（0 次 API 调用，便宜优先）。
		// 顺序对齐 CC 源码：budget → snip → micro。
		*history = toolResultBudget(*history, workdir)
		*history = snipCompact(*history)
		*history = microCompact(*history)

		// s08: 仍超阈值 → LLM 全量摘要（1 次 API 调用）。
		if estimateSize(*history) > contextLimit {
			fmt.Println("[auto compact]")
			*history = compactHistory(ctx, client, model, *history, workdir)
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
			// reactive: API 报 prompt_too_long，触发应急压缩并重试。
			if reactiveRetries < maxReactiveRetries && isPromptTooLong(err) {
				fmt.Println("[reactive compact]")
				*history = reactiveCompact(ctx, client, model, *history, workdir)
				reactiveRetries++
				continue
			}
			return "", fmt.Errorf("agent loop round %d error: [%w]", round, err)
		}
		reactiveRetries = 0 // 成功一次后重置

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

		// 工具结果处理：compact 工具触发 L4 摘要并替换 history。
		results, compactSummary, _ := runToolResults(ctx, client, model, workdir, resp.Content, &subSystem, subTools, roundsSinceTodo)
		if compactSummary {
			// compact 工具：对外层而言本轮至少有一个 compact tool_use。直接对外层传入的
			// *history 重做 L4 摘要（包含本轮 assistant 消息），然后整体替换。
			*history = compactHistory(ctx, client, model, *history, workdir)
			*history = append(*history, ai.NewUserMessage(results...))
			continue
		}
		*history = append(*history, ai.NewUserMessage(results...))
	}
}

// runToolResults 跑一遍 assistant 响应里的 tool_use：先 hooks，再 dispatch。
// compact 工具返回 compactSummary=true，外层会重做 L4 摘要并替换 history。
func runToolResults(ctx context.Context, client ai.Client, model, workdir string, content []ai.ContentBlockUnion, subSystem *string, subTools []ai.ToolUnionParam, roundsSinceTodo *int) ([]ai.ContentBlockParamUnion, bool, int) {
	var results []ai.ContentBlockParamUnion
	compactCount := 0
	for _, block := range content {
		tu, ok := block.AsAny().(ai.ToolUseBlock)
		if !ok {
			continue
		}
		fmt.Printf("%s> %s%s\n", colorCyan, tu.Name, colorReset)

		// s08: compact 工具触发 L4 摘要（不是 no-op）。
		if tu.Name == "compact" {
			compactCount++
			results = append(results, ai.NewToolResultBlock(tu.ID, "[Compacted. Conversation history has been summarized.]", false))
			continue
		}

		blocked := triggerHooks(hookPreToolUse, tu.Name, tu.Input)
		if blocked != "" {
			results = append(results, ai.NewToolResultBlock(tu.ID, blocked, false))
			continue
		}

		// task 工具派生子 agent，需要 client/model/subSystem/subTools，单独处理。
		var out string
		if tu.Name == "task" {
			out = dispatchTaskTool(ctx, client, model, *subSystem, subTools, tu.Input)
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
	return results, compactCount > 0, compactCount
}

// dispatchTaskTool 包装 s06 的 task 工具派发。
func dispatchTaskTool(ctx context.Context, client ai.Client, model, subSystem string, subTools []ai.ToolUnionParam, input json.RawMessage) string {
	var in struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return spawnSubagent(ctx, client, model, subSystem, subTools, in.Description)
}

// isPromptTooLong 检查 SDK 错误是否对应 prompt_too_long / too many tokens。
// 对齐 Python：`prompt_too_long` in str(e).lower() or `too many tokens` in str(e).lower()
func isPromptTooLong(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "prompt_too_long") || strings.Contains(s, "too many tokens")
}

// ═══════════════════════════════════════════════════════════
//  NEW in s08: 四层压缩 + 应急压缩
// ═══════════════════════════════════════════════════════════

// estimateSize 估算 messages 的字节大小（对齐 Python 版 len(str(msgs))）。
func estimateSize(msgs []ai.MessageParam) int {
	// 用 JSON 序列化估算大小——比直接拼字符串更稳。
	data, err := json.Marshal(msgs)
	if err != nil {
		// 退化为字符串拼接。
		return len(fmt.Sprintf("%v", msgs))
	}
	return len(data)
}

// _messageHasToolUse 判断 msg 是否包含任何 tool_use 块。
func messageHasToolUse(msg ai.MessageParam) bool {
	if msg.Role != "assistant" {
		return false
	}
	for _, b := range msg.Content {
		if b.OfToolUse != nil {
			return true
		}
	}
	return false
}

// _isToolResultMessage 判断 msg 是否包含任何 tool_result 块。
func isToolResultMessage(msg ai.MessageParam) bool {
	if msg.Role != "user" {
		return false
	}
	for _, b := range msg.Content {
		if b.OfToolResult != nil {
			return true
		}
	}
	return false
}

// L1: snipCompact — 消息数超阈值时裁剪中间，避免切断 tool_use/tool_result 对。
func snipCompact(messages []ai.MessageParam) []ai.MessageParam {
	maxMessages := 50
	if len(messages) <= maxMessages {
		return messages
	}
	keepHead, keepTail := 3, maxMessages-3
	headEnd, tailStart := keepHead, len(messages)-keepTail

	// 头：不要在 tool_use 之后立刻切走它的 tool_result。
	if headEnd > 0 && headEnd <= len(messages) && messageHasToolUse(messages[headEnd-1]) {
		for headEnd < len(messages) && isToolResultMessage(messages[headEnd]) {
			headEnd++
		}
	}

	// 尾：不要把 tool_result 与前一条 tool_use 切开。
	if tailStart > 0 && tailStart < len(messages) &&
		isToolResultMessage(messages[tailStart]) &&
		messageHasToolUse(messages[tailStart-1]) {
		tailStart--
	}

	if headEnd >= tailStart {
		return messages
	}
	snipped := tailStart - headEnd
	out := make([]ai.MessageParam, 0, headEnd+1+(len(messages)-tailStart))
	out = append(out, messages[:headEnd]...)
	out = append(out, ai.NewUserMessage(ai.NewTextBlock(fmt.Sprintf("[snipped %d messages]", snipped))))
	out = append(out, messages[tailStart:]...)
	return out
}

// collectToolResults 收集 messages 中所有 (msgIdx, blockIdx) 的 tool_result 块。
func collectToolResults(messages []ai.MessageParam) []toolResultRef {
	var out []toolResultRef
	for mi, msg := range messages {
		if msg.Role != "user" {
			continue
		}
		for bi, block := range msg.Content {
			if block.OfToolResult != nil {
				out = append(out, toolResultRef{MsgIdx: mi, BlockIdx: bi, Block: block})
			}
		}
	}
	return out
}

// toolResultRef 描述一条 tool_result 块在 messages 里的位置。
type toolResultRef struct {
	MsgIdx   int
	BlockIdx int
	Block    ai.ContentBlockParamUnion
}

// toolResultText 取出 tool_result 块的第一段文本（OfText）。非文本内容返回 ""。
func toolResultText(tr *ai.ToolResultBlockParam) string {
	if tr == nil || len(tr.Content) == 0 {
		return ""
	}
	for _, c := range tr.Content {
		if c.OfText != nil {
			return c.OfText.Text
		}
	}
	return ""
}

// setToolResultText 替换 tool_result 块的文本内容（保持单一 OfText 段）。
func setToolResultText(tr *ai.ToolResultBlockParam, text string) {
	tr.Content = []ai.ToolResultBlockParamContentUnion{
		{OfText: &ai.TextBlockParam{Text: text}},
	}
}

// L2: microCompact — 旧 tool_result 替换为占位符（保留最近 KEEP_RECENT 个）。
func microCompact(messages []ai.MessageParam) []ai.MessageParam {
	results := collectToolResults(messages)
	if len(results) <= keepRecent {
		return messages
	}
	// 复制一份 messages 以避免改到外层 slice。
	out := make([]ai.MessageParam, len(messages))
	copy(out, messages)
	// 重新在副本上收集（索引在副本里依然有效）。
	results = collectToolResults(out)
	const placeholder = "[Earlier tool result compacted. Re-run if needed.]"
	for _, r := range results[:len(results)-keepRecent] {
		tr := r.Block.OfToolResult
		if tr == nil {
			continue
		}
		text := toolResultText(tr)
		if len(text) > 120 {
			setToolResultText(tr, placeholder)
		}
	}
	return out
}

// L3: persistLargeOutput — 把大结果落盘，返回带预览的占位字符串。
func persistLargeOutput(toolUseID, output string, workdir string) string {
	if len(output) <= persistThreshold {
		return output
	}
	dir := filepath.Join(workdir, toolResultsDirName)
	_ = os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, toolUseID+".txt")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.WriteFile(path, []byte(output), 0o644)
	}
	preview := output
	if len(preview) > 2000 {
		preview = preview[:2000]
	}
	return fmt.Sprintf("<persisted-output>\nFull output: %s\nPreview:\n%s\n</persisted-output>", path, preview)
}

// L3: toolResultBudget — 当最后一条 user 消息的 tool_result 字节总和超上限时，
// 把最大的若干个落盘到磁盘。Go 版的 messages 不可直接索引修改 Content 数组，
// 因此需要重建最后一条消息。
func toolResultBudget(messages []ai.MessageParam, workdir string) []ai.MessageParam {
	if len(messages) == 0 {
		return messages
	}
	last := messages[len(messages)-1]
	if last.Role != "user" {
		return messages
	}
	// 收集所有 tool_result 块。
	type pair struct {
		idx int
		blk ai.ContentBlockParamUnion
	}
	var blocks []pair
	for i, b := range last.Content {
		if b.OfToolResult != nil {
			blocks = append(blocks, pair{i, b})
		}
	}
	if len(blocks) == 0 {
		return messages
	}
	total := 0
	for _, p := range blocks {
		total += len(toolResultText(p.blk.OfToolResult))
	}
	if total <= resultBudgetMax {
		return messages
	}
	// 按字节大小降序排列，先把最大的落盘。
	sort.Slice(blocks, func(i, j int) bool {
		return len(toolResultText(blocks[i].blk.OfToolResult)) > len(toolResultText(blocks[j].blk.OfToolResult))
	})
	for _, p := range blocks {
		if total <= resultBudgetMax {
			break
		}
		text := toolResultText(p.blk.OfToolResult)
		if len(text) <= persistThreshold {
			continue
		}
		tid := p.blk.OfToolResult.ToolUseID
		if tid == "" {
			tid = "unknown"
		}
		newContent := persistLargeOutput(tid, text, workdir)
		setToolResultText(p.blk.OfToolResult, newContent)
		// 重新计算 total。
		total = 0
		for _, q := range blocks {
			total += len(toolResultText(q.blk.OfToolResult))
		}
	}
	// 重建最后一条消息。
	newContent := make([]ai.ContentBlockParamUnion, len(last.Content))
	copy(newContent, last.Content)
	// 把改动后的 blocks 写回。
	for _, p := range blocks {
		newContent[p.idx] = p.blk
	}
	out := make([]ai.MessageParam, len(messages))
	copy(out, messages)
	out[len(out)-1] = ai.MessageParam{Role: "user", Content: newContent}
	return out
}

// L4: writeTranscript — 把 messages 序列化为 JSONL 落盘。
func writeTranscript(messages []ai.MessageParam, workdir string) string {
	dir := filepath.Join(workdir, transcriptDirName)
	_ = os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, fmt.Sprintf("transcript_%d.jsonl", time.Now().Unix()))
	f, err := os.Create(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	for _, msg := range messages {
		// ai.MessageParam 不直接 JSON-tag；用 SDK 提供的字段 + 手动拼。
		// 这里我们只关心 role/content 序列化，足够做离线摘要。
		row := map[string]any{
			"role":    msg.Role,
			"content": msg.Content,
		}
		data, err := json.Marshal(row)
		if err != nil {
			continue
		}
		f.Write(data)
		f.Write([]byte("\n"))
	}
	return path
}

// L4: summarizeHistory — 调用 LLM 生成全量摘要。
func summarizeHistory(ctx context.Context, client ai.Client, model string, messages []ai.MessageParam) string {
	conv, err := json.Marshal(messages)
	if err != nil {
		conv = []byte(fmt.Sprintf("%v", messages))
	}
	if len(conv) > 80000 {
		conv = conv[:80000]
	}
	prompt := "Summarize this coding-agent conversation so work can continue.\n" +
		"Preserve: 1. current goal, 2. key findings/decisions, 3. files read/changed, " +
		"4. remaining work, 5. user constraints.\nBe compact but concrete.\n\n" + string(conv)
	resp, err := client.Messages.New(ctx, ai.MessageNewParams{
		Model:     ai.Model(model),
		MaxTokens: 2000,
		Messages:  []ai.MessageParam{ai.NewUserMessage(ai.NewTextBlock(prompt))},
	})
	if err != nil {
		return fmt.Sprintf("(summary failed: %s)", err)
	}
	var sb strings.Builder
	for _, b := range resp.Content {
		if tb, ok := b.AsAny().(ai.TextBlock); ok {
			sb.WriteString(tb.Text)
		}
	}
	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "(empty summary)"
	}
	return out
}

// L4: compactHistory — 写转写 + 调摘要，返回单条 [Compacted] 摘要消息。
func compactHistory(ctx context.Context, client ai.Client, model string, messages []ai.MessageParam, workdir string) []ai.MessageParam {
	transcriptPath := writeTranscript(messages, workdir)
	if transcriptPath != "" {
		fmt.Printf("[transcript saved: %s]\n", transcriptPath)
	}
	summary := summarizeHistory(ctx, client, model, messages)
	return []ai.MessageParam{
		ai.NewUserMessage(ai.NewTextBlock("[Compacted]\n\n" + summary)),
	}
}

// reactiveCompact — 应急压缩：保留最近 5 条（含末尾 tool_result 配对），其余用摘要替换。
func reactiveCompact(ctx context.Context, client ai.Client, model string, messages []ai.MessageParam, workdir string) []ai.MessageParam {
	_ = writeTranscript(messages, workdir)
	tailStart := len(messages) - 5
	if tailStart < 0 {
		tailStart = 0
	}
	// 尾：不要把 tool_result 与前一条 tool_use 切开。
	if tailStart > 0 && tailStart < len(messages) &&
		isToolResultMessage(messages[tailStart]) &&
		messageHasToolUse(messages[tailStart-1]) {
		tailStart--
	}
	summary := summarizeHistory(ctx, client, model, messages[:tailStart])
	out := make([]ai.MessageParam, 0, 1+(len(messages)-tailStart))
	out = append(out, ai.NewUserMessage(ai.NewTextBlock("[Reactive compact]\n\n"+summary)))
	out = append(out, messages[tailStart:]...)
	return out
}

// ═══════════════════════════════════════════════════════════
//  FROM s07: 技能注册表 + SYSTEM 目录构建 + 按需加载
// ═══════════════════════════════════════════════════════════

// skillFrontmatter 是 SKILL.md 头部 YAML frontmatter 里只取这两个字段。
type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// parseFrontmatter 解析 SKILL.md 的 YAML frontmatter，返回 (meta, body)。
func parseFrontmatter(text string) (skillFrontmatter, string) {
	var meta skillFrontmatter
	if !strings.HasPrefix(text, "---") {
		return meta, text
	}
	rest := strings.TrimPrefix(text, "---")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return meta, text
	}
	header := rest[:end]
	body := strings.TrimLeft(rest[end+4:], "\n")
	if err := yaml.Unmarshal([]byte(header), &meta); err != nil {
		return skillFrontmatter{}, text
	}
	return meta, body
}

// scanSkills 扫描 skillsDir 下的 `<name>/SKILL.md`，填充 skillRegistry。
func scanSkills(skillsDir string) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return
	}
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

const nagRounds = 3

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
		return fmt.Sprintf("Error: text not found in %s", path)
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

// allTools 返回父 agent 的全部 9 个工具：s07 的 8 个 + compact。
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
		compactToolDef(),
	}
}

// subAgentTools 返回子 agent 的工具集：5 个基础工具，不含 task（防递归）、
// 不含 load_skill（子 agent 不加载技能）、不含 todo_write、不含 compact。
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

// compactToolDef 是 s08 新增工具：模型主动触发 L4 摘要。
func compactToolDef() ai.ToolUnionParam {
	return ai.ToolUnionParam{OfTool: &ai.ToolParam{
		Name:        "compact",
		Description: ai.String("Summarize earlier conversation to free context space."),
		InputSchema: ai.ToolInputSchemaParam{
			Properties: map[string]any{
				"focus": map[string]any{"type": "string"},
			},
		},
	}}
}

// dispatchTool 按工具名查表执行（task / compact 在 agentLoop 里单独处理）。
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
