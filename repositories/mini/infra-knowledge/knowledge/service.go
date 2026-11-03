package knowledge

import (
	"context"

	"github.com/google/uuid"
)

// Service 定义知识库模块对外提供的业务接口。
type Service interface {
	Create(ctx context.Context, request CreateDocumentRequest) (Document, error)
	Reindex(ctx context.Context, documentID string) error
}

// service 实现知识库模块的业务接口。
type service struct {
	store   Store
	indexer Indexer
}

// NewService 创建知识库服务。
func NewService(store Store, indexer Indexer) Service {
	return &service{store: store, indexer: indexer}
}

// newID 生成 UUID v7 文档标识。
func newID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// Create 创建知识文档，并为新建文档同步构建索引。
// 索引失败时返回错误，已保存的文档仍然可见，可通过 Reindex 重试。
func (s *service) Create(ctx context.Context, request CreateDocumentRequest) (Document, error) {
	id, err := newID()
	if err != nil {
		return Document{}, err
	}
	doc, err := NewDocument(id, request.FileName, request.Title, request.Content, request.ProjectID, request.ProjectName)
	if err != nil {
		return doc, err
	}

	saved, created, err := s.store.GetOrCreate(ctx, doc)
	if err != nil {
		return Document{}, err
	}
	if !created {
		return saved, nil
	}

	if err := s.indexer.Build(ctx, saved); err != nil {
		return Document{}, err
	}
	return saved, nil
}

// Reindex 同步重新构建指定文档的索引。
func (s *service) Reindex(ctx context.Context, documentID string) error {
	doc, err := s.store.Get(ctx, documentID)
	if err != nil {
		return err
	}
	return s.indexer.Build(ctx, doc)
}

// CreateDocumentRequest 表示创建知识文档的请求。
type CreateDocumentRequest struct {
	FileName    string
	Title       string
	Content     string
	ProjectID   string
	ProjectName string
}
