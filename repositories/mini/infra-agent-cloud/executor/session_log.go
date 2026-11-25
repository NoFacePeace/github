package executor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SessionLog 保存引擎输出的本地会话日志。
type SessionLog struct {
	file *os.File
	mu   sync.Mutex
}

// openSessionLog 以追加模式打开本轮的日志文件。
func openSessionLog(spec *Spec) (*SessionLog, error) {
	name := spec.SessionID + ".jsonl"
	if spec.SessionID == "" || !filepath.IsLocal(name) || filepath.Base(name) != name {
		return nil, errors.New("invalid session log filename")
	}
	if err := os.MkdirAll(spec.Paths.Session, 0o700); err != nil {
		return nil, fmt.Errorf("create session log directory: %w", err)
	}
	file, err := os.OpenFile(filepath.Join(spec.Paths.Session, name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open session log: %w", err)
	}
	return &SessionLog{file: file}, nil
}

// append 追加一行，过滤结果为 nil 或空白时跳过。
func (l *SessionLog) append(line *string) error {
	if line == nil || strings.TrimSpace(*line) == "" {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	text := strings.TrimRight(*line, "\r\n")
	if _, err := l.file.WriteString(normalizeLogLine(text) + "\n"); err != nil {
		return fmt.Errorf("append session log: %w", err)
	}
	return nil
}

// close 等待正在进行的写入并关闭日志文件。
func (l *SessionLog) close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

// normalizeLogLine 给缺少字符串时间戳的 JSON 对象补上当前时间。
func normalizeLogLine(line string) string {
	var record map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &record); err != nil || record == nil {
		return line
	}
	if timestamp := record["timestamp"]; len(timestamp) == 0 || timestamp[0] != '"' {
		record["timestamp"], _ = json.Marshal(time.Now().UTC().Format(time.RFC3339Nano))
	}
	data, err := json.Marshal(record)
	if err != nil {
		return line
	}
	return string(data)
}
