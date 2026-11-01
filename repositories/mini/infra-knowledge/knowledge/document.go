package knowledge

// Document 表示知识库中的文档。
type Document struct {
	ID       string
	FileName string
	Title    string
	Content  string
}

// NewDocument 创建文档，具体逻辑待实现。
func NewDocument(id, fileName, title, content string) (Document, error) {
	var doc Document
	return doc, nil
}
