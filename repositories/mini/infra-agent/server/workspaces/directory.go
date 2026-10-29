package workspaces

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDirectory 根据配置确保工作目录存在，工作空间 ID 为空时使用 default。
func EnsureDirectory(cfg Config, workspaceID string) (string, error) {
	if !filepath.IsAbs(cfg.DataDirectory) {
		return "", fmt.Errorf("data directory must be absolute: %q", cfg.DataDirectory)
	}
	if workspaceID == "" {
		workspaceID = "default"
	}
	if !filepath.IsLocal(workspaceID) || filepath.Clean(workspaceID) != workspaceID || workspaceID == "." {
		return "", fmt.Errorf("invalid workspace id %q", workspaceID)
	}
	directory := filepath.Join(cfg.DataDirectory, "workspaces", workspaceID)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", fmt.Errorf("create workspace directory %q: %w", directory, err)
	}
	return directory, nil
}

// ValidateDirectory 校验路径是绝对路径且指向已存在的目录，允许符号链接。
func ValidateDirectory(directory string) error {
	if !filepath.IsAbs(directory) {
		return fmt.Errorf("workspace directory must be absolute: %q", directory)
	}
	info, err := os.Stat(directory)
	if err != nil {
		return fmt.Errorf("stat workspace directory %q: %w", directory, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("workspace directory is not a directory: %q", directory)
	}
	return nil
}
