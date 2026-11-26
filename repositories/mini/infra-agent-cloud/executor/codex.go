package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Codex struct{}

func NewCodex() *Codex {
	return &Codex{}
}

func (c *Codex) Name() string {
	return "codex"
}

func (c *Codex) StartSteer(_ context.Context, _ *Spec, _ io.WriteCloser, _ io.ReadCloser) (SteerSession, error) {
	// TODO: 初始化 Codex App Server 协议并发送首条提示词。
	return SteerSession{}, errors.New("codex steer is not implemented")
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
		Executable: c.Name(),
		Args:       args,
		Dir:        spec.Paths.Workspace,
		Env:        env,
	}, nil
}

// ParseSessionID 从 Codex 的会话启动事件提取续接标识。
func (c *Codex) ParseSessionID(line string) string {
	var event struct {
		Type     string `json:"type"`
		ThreadID string `json:"thread_id"`
	}
	if err := json.Unmarshal([]byte(line), &event); err != nil || event.Type != "thread.started" {
		return ""
	}
	return event.ThreadID
}

func (c *Codex) ParseMCPStatus(line string) *EngineMCPStatus {
	return nil
}

func (c *Codex) ParseProgress(line string) *EngineProgress {
	return nil
}

func (c *Codex) ParseChannelOutput(line string) []EngineChannelOutput {
	return nil
}

// FilterLogLine 跳过 Codex 的流式文本增量。
func (c *Codex) FilterLogLine(line string) *string {
	if strings.Contains(line, `"response.output_text.delta"`) {
		return nil
	}
	return &line
}

// FilterErrorLine 跳过 Codex 等待标准输入的提示。
func (c *Codex) FilterErrorLine(line string) *string {
	if strings.TrimSpace(line) == "Reading additional input from stdin..." {
		return nil
	}
	return &line
}

type codexOutputEvent struct {
	Type string `json:"type"`
	Item struct {
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

// ParseResult 从整轮 Codex JSONL 解析回复、用量，并检查引擎终态。
func (c *Codex) ParseResult(log string, _ *Spec) (EngineResult, error) {
	var result EngineResult
	var turnCompleted, invalidOutput bool
	var turnFailure, lastError string
	for line := range strings.Lines(log) {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var event codexOutputEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil || event.Type == "" {
			invalidOutput = true
			continue
		}
		switch event.Type {
		case "item.completed":
			if event.Item.Type == "agent_message" {
				if event.Item.OutputText != nil {
					result.Text = event.Item.OutputText
				} else if event.Item.Text != nil {
					result.Text = event.Item.Text
				}
			}
		case "turn.completed":
			turnCompleted = true
			if usage := event.Usage; usage != nil {
				result.Usage = &Usage{
					InputTokens:           usage.InputTokens,
					CachedInputTokens:     usage.CachedInputTokens,
					OutputTokens:          usage.OutputTokens,
					ReasoningOutputTokens: usage.ReasoningOutputTokens,
					TotalTokens:           usage.TotalTokens,
				}
			}
		case "turn.failed":
			turnFailure = event.Error.Message
			if turnFailure == "" {
				turnFailure = "codex turn failed"
			}
		case "error":
			lastError = event.Message
			if lastError == "" {
				lastError = event.Error.Message
			}
		}
	}
	if invalidOutput {
		return result, errors.New("invalid codex output event")
	}
	if turnFailure != "" {
		return result, errors.New(turnFailure)
	}
	if turnCompleted {
		return result, nil
	}
	if lastError != "" {
		return result, errors.New(lastError)
	}
	return result, errors.New("codex output ended without a terminal turn event")
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
