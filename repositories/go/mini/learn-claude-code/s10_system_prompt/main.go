package main

// s10: System Prompt — runtime prompt assembly with a deterministic cache.
//
// Run from this directory:
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
	"strings"
	"sync"
	"time"

	ai "github.com/anthropics/anthropic-sdk-go"
	aiopt "github.com/anthropics/anthropic-sdk-go/option"
)

const (
	maxTokens     = 8000
	maxToolOutput = 50000
	bashTimeout   = 120 * time.Second
)

var promptSections = map[string]string{
	"identity": "You are a coding agent. Act, don't explain.",
}

// promptContext is deliberately made only of real, serializable state. This
// makes its JSON representation a stable process-local cache key.
type promptContext struct {
	EnabledTools []string `json:"enabled_tools"`
	Workspace    string   `json:"workspace"`
	Memories     string   `json:"memories"`
}

type promptCache struct {
	mu     sync.Mutex
	key    string
	prompt string
}

func main() {
	workdir, err := os.Getwd()
	if err != nil {
		fmt.Println("getwd:", err)
		return
	}
	model := os.Getenv("MODEL_ID")
	if model == "" {
		fmt.Println("MODEL_ID is required")
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

	fmt.Println("s10: system prompt — runtime assembly")
	fmt.Println("Enter a question, press Enter to send. Type q to quit.")
	reader, cache, history := bufio.NewReader(os.Stdin), &promptCache{}, []ai.MessageParam{}
	for {
		fmt.Print("s10 >> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		query := strings.TrimSpace(line)
		if query == "" || strings.EqualFold(query, "q") || strings.EqualFold(query, "exit") {
			return
		}
		history = append(history, ai.NewUserMessage(ai.NewTextBlock(query)))
		if err := agentLoop(context.Background(), client, model, workdir, cache, &history); err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println(extractTextFromMessage(history[len(history)-1]))
	}
}

// assembleSystemPrompt selects sections from actual runtime context, rather
// than relying on an ever-growing hard-coded prompt string.
func assembleSystemPrompt(ctx promptContext) string {
	sections := []string{promptSections["identity"]}
	if len(ctx.EnabledTools) > 0 {
		sections = append(sections, "Available tools: "+strings.Join(ctx.EnabledTools, ", ")+".")
	}
	sections = append(sections, "Working directory: "+ctx.Workspace)
	if ctx.Memories != "" {
		sections = append(sections, "Relevant memories:\n"+ctx.Memories)
	}
	return strings.Join(sections, "\n\n")
}

// getSystemPrompt avoids reassembling an unchanged prompt. json.Marshal is
// deterministic for struct fields and is safe for nested state unlike hashes.
func (c *promptCache) getSystemPrompt(ctx promptContext) string {
	keyBytes, err := json.Marshal(ctx)
	if err != nil {
		return assembleSystemPrompt(ctx)
	}
	key := string(keyBytes)
	c.mu.Lock()
	defer c.mu.Unlock()
	if key == c.key && c.prompt != "" {
		fmt.Println("  [cache hit] system prompt unchanged")
		return c.prompt
	}
	c.key, c.prompt = key, assembleSystemPrompt(ctx)
	loaded := []string{"identity", "tools", "workspace"}
	if ctx.Memories != "" {
		loaded = append(loaded, "memory")
	}
	fmt.Println("  [assembled] sections: " + strings.Join(loaded, ", "))
	return c.prompt
}

func updateContext(workdir string) promptContext {
	memories := ""
	if b, err := os.ReadFile(filepath.Join(workdir, ".memory", "MEMORY.md")); err == nil {
		memories = strings.TrimSpace(string(b))
	}
	return promptContext{EnabledTools: []string{"bash", "read_file", "write_file"}, Workspace: workdir, Memories: memories}
}

func agentLoop(ctx context.Context, client ai.Client, model, workdir string, cache *promptCache, history *[]ai.MessageParam) error {
	state := updateContext(workdir)
	system := cache.getSystemPrompt(state)
	for turn := 0; turn < 30; turn++ {
		resp, err := client.Messages.New(ctx, ai.MessageNewParams{Model: ai.Model(model), MaxTokens: maxTokens, System: []ai.TextBlockParam{{Text: system}}, Messages: *history, Tools: tools()})
		if err != nil {
			return err
		}
		*history = append(*history, resp.ToParam())
		if resp.StopReason != ai.StopReasonToolUse {
			return nil
		}
		var results []ai.ContentBlockParamUnion
		for _, block := range resp.Content {
			tu, ok := block.AsAny().(ai.ToolUseBlock)
			if !ok {
				continue
			}
			fmt.Printf("> %s\n", tu.Name)
			out := dispatchTool(ctx, workdir, tu.Name, tu.Input)
			fmt.Println(truncate(out, 200))
			results = append(results, ai.NewToolResultBlock(tu.ID, out, false))
		}
		*history = append(*history, ai.NewUserMessage(results...))
		// A tool might have created/edited .memory/MEMORY.md, so refresh from disk.
		state = updateContext(workdir)
		system = cache.getSystemPrompt(state)
	}
	return fmt.Errorf("agent stopped after 30 tool rounds")
}

func tools() []ai.ToolUnionParam { return []ai.ToolUnionParam{bashTool(), readTool(), writeTool()} }
func tool(name, description string, properties map[string]any, required ...string) ai.ToolUnionParam {
	return ai.ToolUnionParam{OfTool: &ai.ToolParam{Name: name, Description: ai.String(description), InputSchema: ai.ToolInputSchemaParam{Properties: properties, Required: required}}}
}
func bashTool() ai.ToolUnionParam {
	return tool("bash", "Run a shell command.", map[string]any{"command": map[string]any{"type": "string"}}, "command")
}
func readTool() ai.ToolUnionParam {
	return tool("read_file", "Read file contents.", map[string]any{"path": map[string]any{"type": "string"}, "limit": map[string]any{"type": "integer"}}, "path")
}
func writeTool() ai.ToolUnionParam {
	return tool("write_file", "Write content to a file.", map[string]any{"path": map[string]any{"type": "string"}, "content": map[string]any{"type": "string"}}, "path", "content")
}

func dispatchTool(ctx context.Context, workdir, name string, input json.RawMessage) string {
	switch name {
	case "bash":
		var in struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return "Error: " + err.Error()
		}
		toolCtx, cancel := context.WithTimeout(ctx, bashTimeout)
		defer cancel()
		cmd := exec.CommandContext(toolCtx, "sh", "-c", in.Command)
		cmd.Dir = workdir
		out, err := cmd.CombinedOutput()
		if toolCtx.Err() == context.DeadlineExceeded {
			return "Error: Timeout (120s)"
		}
		if err != nil {
			return fmt.Sprintf("%s\nError: %v", truncate(string(out), maxToolOutput), err)
		}
		return orOutput(truncate(string(out), maxToolOutput))
	case "read_file":
		var in struct {
			Path  string `json:"path"`
			Limit int    `json:"limit"`
		}
		if err := json.Unmarshal(input, &in); err != nil {
			return "Error: " + err.Error()
		}
		path, err := safePath(workdir, in.Path)
		if err != nil {
			return "Error: " + err.Error()
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return "Error: " + err.Error()
		}
		lines := strings.Split(string(b), "\n")
		if in.Limit > 0 && in.Limit < len(lines) {
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
		path, err := safePath(workdir, in.Path)
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
	default:
		return "Unknown: " + name
	}
}

func safePath(workdir, path string) (string, error) {
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
func extractTextFromMessage(m ai.MessageParam) string {
	var b strings.Builder
	for _, c := range m.Content {
		if c.OfText != nil {
			b.WriteString(c.OfText.Text)
		}
	}
	return b.String()
}
func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
func orOutput(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(no output)"
	}
	return s
}
