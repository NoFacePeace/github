package attachments

import "context"

// Service 定义附件模块对外提供的业务接口。
type Service interface {
	// BatchGet 根据 ID 批量获取附件元数据，结果顺序与 ids 一致。
	// 任一附件不存在时返回错误，不返回部分结果。
	BatchGet(ctx context.Context, ids []string) ([]Attachment, error)
}
