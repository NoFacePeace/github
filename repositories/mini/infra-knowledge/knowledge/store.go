package knowledge

import "context"

// Store 定义知识文档的存储接口。
type Store interface {
	Save(ctx context.Context, doc Document) error
}
