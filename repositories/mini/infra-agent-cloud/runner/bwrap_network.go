package runner

import (
	"fmt"
	"os"
	"path/filepath"
)

// prepareExecutionNetwork 为本轮准备 RootlessKit 状态目录。
func (r *BwrapRuntime) prepareExecutionNetwork() (string, error) {
	if r.RootlessKitExecutable == "" {
		return "", nil
	}
	root, err := filepath.Abs(filepath.Join(r.DataRoot, "networks"))
	if err != nil {
		return "", fmt.Errorf("resolve bwrap network directory: %w", err)
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", fmt.Errorf("create bwrap network directory: %w", err)
	}
	dir, err := os.MkdirTemp(root, "execution-")
	if err != nil {
		return "", fmt.Errorf("create execution network directory: %w", err)
	}
	return dir, nil
}
