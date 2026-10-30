package tasks

import (
	"fmt"

	"github.com/google/uuid"
)

// newID 根据资源前缀生成 ID，格式为前缀加下划线和完整 UUIDv7。
func newID(prefix string) (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate uuid: %w", err)
	}
	return prefix + "_" + id.String(), nil
}
