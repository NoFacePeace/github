package knowledge

import "context"

// Service 定义知识库模块对外提供的业务接口。
type Service interface {
	Create(ctx context.Context, request CreateDocumentRequest) (Document, error)
}

// CreateDocumentRequest 表示创建知识文档的请求。
type CreateDocumentRequest struct {
	FileName string
	Title    string
	Content  string
}
