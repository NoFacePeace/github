package knowledge

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// Document 表示知识库中的文档。
type Document struct {
	ID string
	// Fingerprint 表示文档指纹，用于文档去重。
	Fingerprint string
	FileName    string
	Title       string
	Content     string
	ProjectID   string
	ProjectName string
}

// NewDocument 校验文档字段并计算用于项目内去重的指纹。
func NewDocument(id, fileName, title, content, projectID, projectName string) (Document, error) {
	doc := Document{
		ID:          strings.TrimSpace(id),
		FileName:    strings.TrimSpace(fileName),
		Title:       strings.TrimSpace(title),
		Content:     content,
		ProjectID:   strings.TrimSpace(projectID),
		ProjectName: strings.TrimSpace(projectName),
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{"id", doc.ID},
		{"file name", doc.FileName},
		{"title", doc.Title},
		{"content", doc.Content},
		{"project id", doc.ProjectID},
		{"project name", doc.ProjectName},
	} {
		if strings.TrimSpace(field.value) == "" {
			return Document{}, fmt.Errorf("document %s is required", field.name)
		}
	}

	// 保留正文原文；使用固定字段结构序列化，避免简单拼接导致字段边界混淆。
	// 随机 ID 和项目展示名称不参与指纹计算。
	payload, err := json.Marshal(struct {
		ProjectID string `json:"projectId"`
		FileName  string `json:"fileName"`
		Title     string `json:"title"`
		Content   string `json:"content"`
	}{doc.ProjectID, doc.FileName, doc.Title, doc.Content})
	if err != nil {
		return Document{}, fmt.Errorf("encode document fingerprint: %w", err)
	}
	sum := sha256.Sum256(payload)
	doc.Fingerprint = hex.EncodeToString(sum[:])
	return doc, nil
}
