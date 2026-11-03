package knowledge

import (
	"context"
	"errors"
)

// ErrIndexNotFound 表示没有匹配文档 ID 和构建指纹的索引。
var ErrIndexNotFound = errors.New("index not found")

// Store 定义知识文档的存储接口。
type Store interface {
	// Get 根据 ID 读取文档，不存在时返回错误。
	Get(ctx context.Context, documentID string) (Document, error)

	// GetFulltextIndex 根据文档 ID 和构建指纹查询全文索引，不存在时返回 ErrIndexNotFound。
	GetFulltextIndex(ctx context.Context, documentID, fingerprint string) (FulltextIndex, error)

	// GetVectorIndex 根据文档 ID 和构建指纹查询向量索引，不存在时返回 ErrIndexNotFound。
	GetVectorIndex(ctx context.Context, documentID, fingerprint string) (VectorIndex, error)

	// CreateFulltextIndex 保存指定文档和构建指纹对应的全文索引产物。
	CreateFulltextIndex(ctx context.Context, documentID, fingerprint string, index FulltextIndex) error

	// CreateVectorIndex 保存指定文档和构建指纹对应的向量索引产物。
	CreateVectorIndex(ctx context.Context, documentID, fingerprint string, index VectorIndex) error

	// GetOrCreate 按指纹返回已有文档，或创建新文档。
	// created 表示是否新建；实现需保证并发请求不会创建重复文档。
	GetOrCreate(ctx context.Context, doc Document) (saved Document, created bool, err error)
}
