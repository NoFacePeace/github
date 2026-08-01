package main

// s09: Memory System — persistent, cross-session knowledge for the coding agent.
//
// Memories are stored under .memory/. MEMORY.md is a small index injected into
// the system prompt, while only memories relevant to the current request are
// loaded into the request. At the end of each turn the model extracts durable
// preferences and project facts. Once enough files accumulate, it consolidates
// them ("dreaming") to remove duplicates and stale facts.
//
// Run (from this directory):
//
//	ANTHROPIC_API_KEY=... MODEL_ID=... go run .

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	ai "github.com/anthropics/anthropic-sdk-go"
	aiopt "github.com/anthropics/anthropic-sdk-go/option"
)

const (
	memoryDirName        = ".memory"
	memoryIndexName      = "MEMORY.md"
	consolidateThreshold = 10
	maxSelectedMemories  = 5
	maxTokens            = 8000
)

type memory struct {
	Filename    string
	Name        string
	Type        string
	Description string
	Body        string
}

type extractedMemory struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Body        string `json:"body"`
}

func main() {
	workdir, err := os.Getwd()
	if err != nil {
		fmt.Println("getwd:", err)
		return
	}
	store := newMemoryStore(workdir)
	if err := store.ensure(); err != nil {
		fmt.Println("create memory directory:", err)
		return
	}

	var opts []aiopt.RequestOption
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		opts = append(opts, aiopt.WithAPIKey(key))
	}
	if baseURL := os.Getenv("ANTHROPIC_BASE_URL"); baseURL != "" {
		opts = append(opts, aiopt.WithBaseURL(baseURL))
	}
	client := ai.NewClient(opts...)
	model := os.Getenv("MODEL_ID")
	if model == "" {
		fmt.Println("MODEL_ID is required")
		return
	}

	fmt.Println("s09: Memory System — persistent coding-agent memories")
	fmt.Println("Type a question, press Enter. Type q to quit.")
	reader := bufio.NewReader(os.Stdin)
	history := []ai.MessageParam{}
	for {
		fmt.Print("s09 >> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		query := strings.TrimSpace(line)
		if query == "" {
			continue
		}
		if strings.EqualFold(query, "q") || strings.EqualFold(query, "exit") {
			return
		}
		history = append(history, ai.NewUserMessage(ai.NewTextBlock(query)))
		answer, err := runTurn(context.Background(), client, model, store, &history)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println(answer)
	}
}

type memoryStore struct{ dir string }

func newMemoryStore(workdir string) memoryStore {
	return memoryStore{dir: filepath.Join(workdir, memoryDirName)}
}
func (s memoryStore) indexPath() string { return filepath.Join(s.dir, memoryIndexName) }
func (s memoryStore) ensure() error     { return os.MkdirAll(s.dir, 0o755) }

func (s memoryStore) write(m extractedMemory) error {
	if m.Description == "" || m.Body == "" {
		return nil
	}
	if m.Type != "user" && m.Type != "feedback" && m.Type != "project" && m.Type != "reference" {
		m.Type = "user"
	}
	name := strings.TrimSpace(m.Name)
	if name == "" {
		name = "memory"
	}
	filename := slug(name) + ".md"
	text := fmt.Sprintf("---\nname: %s\ndescription: %s\ntype: %s\n---\n\n%s\n", name, m.Description, m.Type, m.Body)
	if err := os.WriteFile(filepath.Join(s.dir, filename), []byte(text), 0o644); err != nil {
		return err
	}
	return s.rebuildIndex()
}

func (s memoryStore) rebuildIndex() error {
	files, err := s.list()
	if err != nil {
		return err
	}
	lines := make([]string, 0, len(files))
	for _, m := range files {
		lines = append(lines, fmt.Sprintf("- [%s](%s) — %s", m.Name, m.Filename, m.Description))
	}
	text := ""
	if len(lines) > 0 {
		text = strings.Join(lines, "\n") + "\n"
	}
	return os.WriteFile(s.indexPath(), []byte(text), 0o644)
}

func (s memoryStore) index() string {
	b, err := os.ReadFile(s.indexPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func (s memoryStore) list() ([]memory, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var out []memory
	for _, e := range entries {
		if e.IsDir() || e.Name() == memoryIndexName || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		meta, body := parseFrontmatter(string(b))
		out = append(out, memory{Filename: e.Name(), Name: or(meta["name"], strings.TrimSuffix(e.Name(), ".md")), Type: or(meta["type"], "user"), Description: meta["description"], Body: body})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Filename < out[j].Filename })
	return out, nil
}

func (s memoryStore) replaceAll(memories []extractedMemory) error {
	files, err := s.list()
	if err != nil {
		return err
	}
	for _, m := range files {
		if err := os.Remove(filepath.Join(s.dir, m.Filename)); err != nil {
			return err
		}
	}
	for _, m := range memories {
		if err := s.write(m); err != nil {
			return err
		}
	}
	return s.rebuildIndex()
}

func parseFrontmatter(text string) (map[string]string, string) {
	meta := map[string]string{}
	if !strings.HasPrefix(text, "---\n") {
		return meta, strings.TrimSpace(text)
	}
	parts := strings.SplitN(text, "---", 3)
	if len(parts) != 3 {
		return meta, strings.TrimSpace(text)
	}
	for _, line := range strings.Split(strings.TrimSpace(parts[1]), "\n") {
		if key, value, ok := strings.Cut(line, ":"); ok {
			meta[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), "\"'")
		}
	}
	return meta, strings.TrimSpace(parts[2])
}

func slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer(" ", "-", "/", "-").Replace(s)
	return strings.Trim(strings.Map(func(r rune) rune {
		if r == '-' || r == '_' || ('a' <= r && r <= 'z') || ('0' <= r && r <= '9') {
			return r
		}
		return -1
	}, s), "-")
}
func or(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func runTurn(ctx context.Context, client ai.Client, model string, store memoryStore, history *[]ai.MessageParam) (string, error) {
	// The index is always cheap; selected full files are attached only to this request.
	relevant := selectRelevant(ctx, client, model, store, *history)
	system := buildSystem(store)
	memoryTurn := len(*history) - 1
	answer := ""
	for turn := 0; turn < 30; turn++ {
		request := append([]ai.MessageParam(nil), (*history)...)
		if len(relevant) > 0 && memoryTurn >= 0 && memoryTurn < len(request) {
			request[memoryTurn] = ai.NewUserMessage(ai.NewTextBlock(renderRelevant(relevant) + "\n\n" + textOf(request[memoryTurn])))
		}
		resp, err := client.Messages.New(ctx, ai.MessageNewParams{Model: ai.Model(model), MaxTokens: maxTokens, System: []ai.TextBlockParam{{Text: system}}, Messages: request, Tools: basicTools()})
		if err != nil {
			return "", err
		}
		*history = append(*history, resp.ToParam())
		if resp.StopReason != ai.StopReasonToolUse {
			answer = extractText(resp.Content)
			break
		}
		var results []ai.ContentBlockParamUnion
		for _, block := range resp.Content {
			tu, ok := block.AsAny().(ai.ToolUseBlock)
			if !ok {
				continue
			}
			fmt.Printf("> %s\n", tu.Name)
			results = append(results, ai.NewToolResultBlock(tu.ID, dispatchTool(ctx, tu.Name, tu.Input), false))
		}
		*history = append(*history, ai.NewUserMessage(results...))
	}
	if answer == "" {
		answer = "Agent stopped after 30 tool rounds without a final answer."
	}

	// Extraction intentionally sees the original history, not the request augmented
	// with relevant-memory content: injected context must never become a new memory.
	if count, err := extractMemories(ctx, client, model, store, *history); err == nil && count > 0 {
		fmt.Printf("[Memory: extracted %d new memories]\n", count)
	}
	if err := consolidate(ctx, client, model, store); err != nil {
		return answer, err
	}
	return answer, nil
}

// The same small coding-tool surface used by the Python lesson. Memory logic is
// independent of tools; tool calls are still retained in history for extraction.
func basicTools() []ai.ToolUnionParam {
	return []ai.ToolUnionParam{bashTool(), readTool(), writeTool(), editTool(), globTool()}
}
func tool(name, description string, properties map[string]any, required ...string) ai.ToolUnionParam {
	return ai.ToolUnionParam{OfTool: &ai.ToolParam{Name: name, Description: ai.String(description), InputSchema: ai.ToolInputSchemaParam{Properties: properties, Required: required}}}
}
func bashTool() ai.ToolUnionParam {
	return tool("bash", "Run a shell command in the workspace.", map[string]any{"command": map[string]any{"type": "string"}}, "command")
}
func readTool() ai.ToolUnionParam {
	return tool("read_file", "Read a workspace file.", map[string]any{"path": map[string]any{"type": "string"}, "limit": map[string]any{"type": "integer"}}, "path")
}
func writeTool() ai.ToolUnionParam {
	return tool("write_file", "Write a workspace file.", map[string]any{"path": map[string]any{"type": "string"}, "content": map[string]any{"type": "string"}}, "path", "content")
}
func editTool() ai.ToolUnionParam {
	return tool("edit_file", "Replace exact text once in a workspace file.", map[string]any{"path": map[string]any{"type": "string"}, "old_text": map[string]any{"type": "string"}, "new_text": map[string]any{"type": "string"}}, "path", "old_text", "new_text")
}
func globTool() ai.ToolUnionParam {
	return tool("glob", "Find workspace files matching a glob.", map[string]any{"pattern": map[string]any{"type": "string"}}, "pattern")
}

func dispatchTool(ctx context.Context, name string, input json.RawMessage) string {
	switch name {
	case "bash":
		var in struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return "Error: " + err.Error()
		}
		cmd := exec.CommandContext(ctx, "sh", "-c", in.Command)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("%s\nError: %v", truncate(string(out), 50000), err)
		}
		return orOutput(truncate(string(out), 50000))
	case "read_file":
		var in struct {
			Path  string `json:"path"`
			Limit int    `json:"limit"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return "Error: " + err.Error()
		}
		path, err := safePath(in.Path)
		if err != nil {
			return "Error: " + err.Error()
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return "Error: " + err.Error()
		}
		lines := strings.Split(string(b), "\n")
		if in.Limit > 0 && len(lines) > in.Limit {
			lines = append(lines[:in.Limit], fmt.Sprintf("... (%d more lines)", len(lines)-in.Limit))
		}
		return strings.Join(lines, "\n")
	case "write_file":
		var in struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return "Error: " + err.Error()
		}
		path, err := safePath(in.Path)
		if err != nil {
			return "Error: " + err.Error()
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "Error: " + err.Error()
		}
		if err := os.WriteFile(path, []byte(in.Content), 0o644); err != nil {
			return "Error: " + err.Error()
		}
		return fmt.Sprintf("Wrote %d bytes to %s", len(in.Content), in.Path)
	case "edit_file":
		var in struct {
			Path string `json:"path"`
			Old  string `json:"old_text"`
			New  string `json:"new_text"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return "Error: " + err.Error()
		}
		path, err := safePath(in.Path)
		if err != nil {
			return "Error: " + err.Error()
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return "Error: " + err.Error()
		}
		if !strings.Contains(string(b), in.Old) {
			return "Error: text not found"
		}
		if err := os.WriteFile(path, []byte(strings.Replace(string(b), in.Old, in.New, 1)), 0o644); err != nil {
			return "Error: " + err.Error()
		}
		return "Edited " + in.Path
	case "glob":
		var in struct {
			Pattern string `json:"pattern"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return "Error: " + err.Error()
		}
		if filepath.IsAbs(in.Pattern) {
			return "Error: absolute patterns are not allowed"
		}
		matches, err := filepath.Glob(in.Pattern)
		if err != nil {
			return "Error: " + err.Error()
		}
		safeMatches := make([]string, 0, len(matches))
		for _, match := range matches {
			if _, err := safePath(match); err == nil {
				safeMatches = append(safeMatches, match)
			}
		}
		return orOutput(strings.Join(safeMatches, "\n"))
	default:
		return "Unknown tool: " + name
	}
}
func safePath(path string) (string, error) {
	workdir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	candidate, err := filepath.Abs(filepath.Join(workdir, path))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(workdir, candidate)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes workspace: %s", path)
	}
	return candidate, nil
}
func orOutput(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(no output)"
	}
	return s
}

func buildSystem(store memoryStore) string {
	section := ""
	if index := store.index(); index != "" {
		section = "\n\nMemories available:\n" + index
	}
	return "You are a coding agent. " + section + "\nRelevant memories are injected below. Respect user preferences from memory. When the user says 'remember' or expresses a clear preference, retain it as a memory."
}

func selectRelevant(ctx context.Context, client ai.Client, model string, store memoryStore, history []ai.MessageParam) []memory {
	files, err := store.list()
	if err != nil || len(files) == 0 {
		return nil
	}
	recent := recentUserText(history)
	if recent == "" {
		return nil
	}
	catalog := make([]string, len(files))
	for i, f := range files {
		catalog[i] = fmt.Sprintf("%d: %s — %s", i, f.Name, f.Description)
	}
	prompt := "Given the recent conversation and memory catalog, return ONLY a JSON array of clearly relevant indices (for example [0, 3]); return [] when none apply.\n\nRecent conversation:\n" + recent + "\n\nMemory catalog:\n" + strings.Join(catalog, "\n")
	resp, err := client.Messages.New(ctx, ai.MessageNewParams{Model: ai.Model(model), MaxTokens: 200, Messages: []ai.MessageParam{ai.NewUserMessage(ai.NewTextBlock(prompt))}})
	if err == nil {
		var indices []int
		if decodeJSONArray(extractText(resp.Content), &indices) == nil {
			out := []memory{}
			for _, i := range indices {
				if i >= 0 && i < len(files) && len(out) < maxSelectedMemories {
					out = append(out, files[i])
				}
			}
			return out
		}
	}
	// Safe fallback when selection request fails: keyword matching against index metadata.
	words := strings.Fields(strings.ToLower(recent))
	out := []memory{}
	for _, f := range files {
		haystack := strings.ToLower(f.Name + " " + f.Description)
		for _, w := range words {
			if len(w) > 3 && strings.Contains(haystack, w) {
				out = append(out, f)
				break
			}
		}
		if len(out) == maxSelectedMemories {
			break
		}
	}
	return out
}

func recentUserText(history []ai.MessageParam) string {
	var texts []string
	for i := len(history) - 1; i >= 0 && len(texts) < 3; i-- {
		if history[i].Role == "user" {
			if text := textOf(history[i]); text != "" {
				texts = append([]string{text}, texts...)
			}
		}
	}
	return truncate(strings.Join(texts, "\n"), 2000)
}

func renderRelevant(memories []memory) string {
	parts := []string{"<relevant_memories>"}
	for _, m := range memories {
		parts = append(parts, fmt.Sprintf("---\nname: %s\ndescription: %s\ntype: %s\n---\n\n%s", m.Name, m.Description, m.Type, m.Body))
	}
	return strings.Join(append(parts, "</relevant_memories>"), "\n\n")
}

func extractMemories(ctx context.Context, client ai.Client, model string, store memoryStore, history []ai.MessageParam) (int, error) {
	dialogue := dialogueText(history)
	if dialogue == "" {
		return 0, nil
	}
	files, err := store.list()
	if err != nil {
		return 0, err
	}
	existing := "(none)"
	if len(files) > 0 {
		lines := make([]string, len(files))
		for i, f := range files {
			lines[i] = fmt.Sprintf("- %s: %s", f.Name, f.Description)
		}
		existing = strings.Join(lines, "\n")
	}
	prompt := "Extract new durable user preferences, constraints, or project facts from this dialogue. Return ONLY a JSON array of {name,type,description,body}. type is one of user, feedback, project, reference. Do not repeat existing memories. Return [] if nothing is new.\n\nExisting memories:\n" + existing + "\n\nDialogue:\n" + truncate(dialogue, 4000)
	resp, err := client.Messages.New(ctx, ai.MessageNewParams{Model: ai.Model(model), MaxTokens: 800, Messages: []ai.MessageParam{ai.NewUserMessage(ai.NewTextBlock(prompt))}})
	if err != nil {
		return 0, err
	}
	var memories []extractedMemory
	if err := decodeJSONArray(extractText(resp.Content), &memories); err != nil {
		return 0, nil
	}
	count := 0
	for _, m := range memories {
		if m.Description != "" && m.Body != "" {
			if err := store.write(m); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

func consolidate(ctx context.Context, client ai.Client, model string, store memoryStore) error {
	files, err := store.list()
	if err != nil || len(files) < consolidateThreshold {
		return err
	}
	parts := make([]string, len(files))
	for i, f := range files {
		parts[i] = fmt.Sprintf("## %s\nname: %s\ndescription: %s\ntype: %s\n%s", f.Filename, f.Name, f.Description, f.Type, f.Body)
	}
	prompt := "Consolidate these memories: merge duplicates, remove stale or contradicted facts, preserve user preferences, and keep at most 30. Return ONLY a JSON array of {name,type,description,body}.\n\n" + truncate(strings.Join(parts, "\n\n"), 16000)
	resp, err := client.Messages.New(ctx, ai.MessageNewParams{Model: ai.Model(model), MaxTokens: 3000, Messages: []ai.MessageParam{ai.NewUserMessage(ai.NewTextBlock(prompt))}})
	if err != nil {
		return nil
	} // Memory maintenance must never break the user turn.
	var consolidated []extractedMemory
	if err := decodeJSONArray(extractText(resp.Content), &consolidated); err != nil {
		return nil
	}
	if err := store.replaceAll(consolidated); err == nil {
		fmt.Printf("[Memory: consolidated %d → %d memories]\n", len(files), len(consolidated))
	}
	return nil
}

func dialogueText(history []ai.MessageParam) string {
	start := len(history) - 10
	if start < 0 {
		start = 0
	}
	parts := []string{}
	for _, m := range history[start:] {
		if text := textOf(m); text != "" {
			parts = append(parts, string(m.Role)+": "+text)
		}
	}
	return strings.Join(parts, "\n")
}
func textOf(m ai.MessageParam) string {
	var b strings.Builder
	for _, c := range m.Content {
		if c.OfText != nil {
			b.WriteString(c.OfText.Text)
		}
	}
	return b.String()
}
func extractText(content []ai.ContentBlockUnion) string {
	var b strings.Builder
	for _, c := range content {
		if t, ok := c.AsAny().(ai.TextBlock); ok {
			b.WriteString(t.Text)
		}
	}
	return strings.TrimSpace(b.String())
}
func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

var jsonArray = regexp.MustCompile(`(?s)\[.*\]`)

func decodeJSONArray(text string, target any) error {
	match := jsonArray.FindString(text)
	if match == "" {
		return fmt.Errorf("no JSON array")
	}
	return json.Unmarshal([]byte(match), target)
}
