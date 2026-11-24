package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Codex struct {
	sessionID     string
	text          *string
	usage         *Usage
	turnCompleted bool
	turnFailure   string
	lastError     string
	parseErr      error
}

func NewCodex() *Codex {
	return &Codex{}
}

// WriteConfig 将 Codex 配置和可选的订阅认证写入本轮指定的配置目录。
func (c *Codex) WriteConfig(spec *Spec, secrets *Secrets) error {
	dir := spec.Paths.Config
	if dir == "" {
		return errors.New("codex config directory is required")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create codex config directory: %w", err)
	}

	config := "cli_auth_credentials_store = \"file\"\n"
	if spec.Config.SystemPrompt != "" {
		quoted, err := json.Marshal(spec.Config.SystemPrompt)
		if err != nil {
			return fmt.Errorf("encode codex system prompt: %w", err)
		}
		// JSON 字符串转义可用于 TOML 基本字符串，另需转义 TOML 禁止的 DEL 字符。
		config += "developer_instructions = " + strings.ReplaceAll(string(quoted), "\x7f", `\u007f`) + "\n"
	}
	if err := writeCodexFile(dir, "config.toml", config); err != nil {
		return fmt.Errorf("write codex config: %w", err)
	}
	if secrets != nil && secrets.CodexAuthJSON != "" {
		if err := writeCodexFile(dir, "auth.json", secrets.CodexAuthJSON); err != nil {
			return fmt.Errorf("write codex auth file: %w", err)
		}
	}
	return nil
}

// BuildCommand 生成普通模式的 Codex CLI 命令，包括可选的会话续接。
func (c *Codex) BuildCommand(spec *Spec, secrets *Secrets) (EngineCommand, error) {
	if spec.SteerPromptMode != "" && spec.SteerPromptMode != SteerPromptModeLegacy {
		return EngineCommand{}, errors.New("codex command supports only legacy input mode")
	}
	if spec.Paths.Config == "" {
		return EngineCommand{}, errors.New("codex config directory is required")
	}
	configDir, err := filepath.Abs(spec.Paths.Config)
	if err != nil {
		return EngineCommand{}, fmt.Errorf("resolve codex config directory: %w", err)
	}
	workspace, err := os.Stat(spec.Paths.Workspace)
	if err != nil {
		return EngineCommand{}, fmt.Errorf("access codex workspace: %w", err)
	}
	if !workspace.IsDir() {
		return EngineCommand{}, errors.New("codex workspace must be a directory")
	}

	args := []string{"exec"}
	if spec.ResumeSessionID != "" {
		args = append(args, "resume")
	}
	args = append(args, "--json", "--skip-git-repo-check")
	if spec.Config.Model != "" {
		args = append(args, "--model", spec.Config.Model)
	}
	args = append(args, "--")
	if spec.ResumeSessionID != "" {
		args = append(args, spec.ResumeSessionID)
	}
	args = append(args, spec.Prompt)

	env := []string{"CODEX_HOME=" + configDir}
	if secrets != nil && secrets.ProviderAPIKey != "" {
		env = append(env, "OPENAI_API_KEY="+secrets.ProviderAPIKey)
	}
	return EngineCommand{
		Executable: "codex",
		Args:       args,
		Dir:        spec.Paths.Workspace,
		Env:        env,
	}, nil
}

// ParseOutput 逐行收集 Codex JSONL 中的会话、回复和用量。
func (c *Codex) ParseOutput(line string) {
	if strings.TrimSpace(line) == "" {
		return
	}
	var event struct {
		Type     string `json:"type"`
		ThreadID string `json:"thread_id"`
		Item     struct {
			Type       string  `json:"type"`
			Text       *string `json:"text"`
			OutputText *string `json:"output_text"`
		} `json:"item"`
		Usage *struct {
			InputTokens           *int64 `json:"input_tokens"`
			CachedInputTokens     *int64 `json:"cached_input_tokens"`
			OutputTokens          *int64 `json:"output_tokens"`
			ReasoningOutputTokens *int64 `json:"reasoning_output_tokens"`
			TotalTokens           *int64 `json:"total_tokens"`
		} `json:"usage"`
		Message string `json:"message"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(line), &event); err != nil || event.Type == "" {
		c.parseErr = errors.New("invalid codex output event")
		return
	}
	switch event.Type {
	case "thread.started":
		c.sessionID = event.ThreadID
	case "item.completed":
		if event.Item.Type == "agent_message" {
			if event.Item.OutputText != nil {
				c.text = event.Item.OutputText
			} else if event.Item.Text != nil {
				c.text = event.Item.Text
			}
		}
	case "turn.completed":
		c.turnCompleted = true
		if usage := event.Usage; usage != nil {
			c.usage = &Usage{
				InputTokens:           usage.InputTokens,
				CachedInputTokens:     usage.CachedInputTokens,
				OutputTokens:          usage.OutputTokens,
				ReasoningOutputTokens: usage.ReasoningOutputTokens,
				TotalTokens:           usage.TotalTokens,
			}
		}
	case "turn.failed":
		c.turnFailure = event.Error.Message
		if c.turnFailure == "" {
			c.turnFailure = "codex turn failed"
		}
	case "error":
		c.lastError = event.Message
		if c.lastError == "" {
			c.lastError = event.Error.Message
		}
	}
}

// ParseResult 填充已收到的结果，验证 Codex 是否报告了执行终态。
func (c *Codex) ParseResult(result *Result) error {
	result.ExecutorSessionID = c.sessionID
	result.Text = c.text
	result.Usage = c.usage
	if c.parseErr != nil {
		return c.parseErr
	}
	if c.turnFailure != "" {
		return errors.New(c.turnFailure)
	}
	if c.turnCompleted {
		return nil
	}
	if c.lastError != "" {
		return errors.New(c.lastError)
	}
	return errors.New("codex output ended without a terminal turn event")
}

// writeCodexFile 原子替换文件，新文件仅允许当前用户读写。
func writeCodexFile(dir, name, content string) error {
	file, err := os.CreateTemp(dir, ".codex-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.WriteString(content)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	return os.Rename(file.Name(), filepath.Join(dir, name))
}

func (c *Codex) Run(ctx context.Context, spec *Spec, emit func(Event) error) (*Result, error) {
	// TODO: 启动 Codex、解析流式输出，并在取消时终止进程。
	return nil, errors.New("codex execution is not implemented")
}
