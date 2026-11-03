package knowledge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/coder/hnsw"
)

// Indexer 定义文档索引的同步构建接口。
type Indexer interface {
	// Build 根据传入文档构建索引，完成后返回。
	// 重建成功后切换索引，失败时保留旧索引。
	Build(ctx context.Context, doc Document) error
}

// Embedder 定义文本向量生成接口。
type Embedder interface {
	// Embed 同步生成一段文本的向量。
	Embed(ctx context.Context, content string) ([]float32, error)
}

// indexer 实现文档索引的同步构建接口。
type indexer struct {
	config   BuildConfig
	store    Store
	embedder Embedder
}

// NewIndexer 根据构建配置创建文档索引构建器。
func NewIndexer(config BuildConfig, store Store, embedder Embedder) Indexer {
	return &indexer{config: config, store: store, embedder: embedder}
}

// Build 构建指定文档的索引，具体逻辑待实现。
func (i *indexer) Build(ctx context.Context, doc Document) error {
	fulltextFingerprint, err := i.config.FulltextFingerprint()
	if err != nil {
		return err
	}
	vectorFingerprint, err := i.config.VectorFingerprint()
	if err != nil {
		return err
	}

	fulltextIndex, fulltextErr := i.store.GetFulltextIndex(ctx, doc.ID, fulltextFingerprint)
	if fulltextErr != nil && !errors.Is(fulltextErr, ErrIndexNotFound) {
		return fulltextErr
	}
	vectorIndex, vectorErr := i.store.GetVectorIndex(ctx, doc.ID, vectorFingerprint)
	if vectorErr != nil && !errors.Is(vectorErr, ErrIndexNotFound) {
		return vectorErr
	}

	// 1. 保留并校验查询到的索引产物，未找到或产物无效时标记为需要构建。
	// 待实现产物有效性校验，当前仅判断两种构建指纹是否匹配。

	// 2. 两种索引都存在且有效时直接返回，否则保留可复用的索引，继续处理。
	// 主动强制重建时跳过复用，强制标识的传递方式待补充。
	if fulltextErr == nil && vectorErr == nil {
		return nil
	}

	// 3. 解析章节并切分文本块，两种索引共用一次处理结果。
	chunks, err := splitChunks(doc, i.config.Split)
	if err != nil {
		return err
	}
	// 4. 全文索引需要构建时，根据文本块生成全文索引产物。
	if errors.Is(fulltextErr, ErrIndexNotFound) {
		fulltextIndex, err = i.buildFulltextIndex(ctx, chunks)
		if err != nil {
			return err
		}
	}

	// 5. 向量索引需要构建时，调用向量模型，校验结果数量、维度和数值有效性，再生成向量索引产物。
	if errors.Is(vectorErr, ErrIndexNotFound) {
		vectorIndex, err = i.buildVectorIndex(ctx, chunks)
		if err != nil {
			return err
		}
	}

	// 6. 所需产物全部构建成功后，仅保存本次新建的索引及各自的构建指纹。
	// 构建配置快照和完成时间待索引结构补充后保存。
	if errors.Is(fulltextErr, ErrIndexNotFound) {
		if err := i.store.CreateFulltextIndex(ctx, doc.ID, fulltextFingerprint, fulltextIndex); err != nil {
			return err
		}
	}
	if errors.Is(vectorErr, ErrIndexNotFound) {
		if err := i.store.CreateVectorIndex(ctx, doc.ID, vectorFingerprint, vectorIndex); err != nil {
			return err
		}
	}

	// 7. 相关缓存失效待接入；跨索引原子保存及多版本发布暂不实现。
	// 全程同步执行，失败时返回错误，文档始终可见。
	return nil
}

// BuildConfig 表示实际生效的索引构建配置快照。
type BuildConfig struct {
	Split    SplitConfig
	Fulltext FulltextConfig
	Vector   VectorConfig
}

// FulltextFingerprint 根据公共切分配置和全文索引配置计算构建指纹，具体逻辑待实现。
func (c BuildConfig) FulltextFingerprint() (string, error) {
	panic("not implemented")
}

// VectorFingerprint 根据公共切分配置和向量索引配置计算构建指纹，具体逻辑待实现。
func (c BuildConfig) VectorFingerprint() (string, error) {
	panic("not implemented")
}

// SplitConfig 表示两种索引共用的文档解析和文本切分配置。
type SplitConfig struct {
	// ChapterMaxTokens 表示章节合并预算，0 使用默认值 512。
	ChapterMaxTokens int
	// ChunkMaxTokens 表示每个文本块的词元上限，0 使用默认值 512。
	ChunkMaxTokens int
}

// FulltextConfig 表示全文索引的构建配置。
// 全文索引的构建指纹由公共切分配置和此配置共同计算。
type FulltextConfig struct{}

// VectorConfig 表示向量索引的构建配置。
// 向量索引的构建指纹由公共切分配置和此配置共同计算。
type VectorConfig struct{}

// FulltextIndex 表示用于关键词检索的全文索引产物。
type FulltextIndex struct{}

// buildFulltextIndex 根据文本块及构建器持有的全文配置构建索引，具体逻辑待实现。
func (i *indexer) buildFulltextIndex(ctx context.Context, chunks []Chunk) (FulltextIndex, error) {
	panic("not implemented")
}

// VectorIndex 表示用于语义检索的向量索引产物。
type VectorIndex struct {
	Dimensions int
	// Graph 保存 HNSW 导出的图数据，包含文本块 ID、向量及节点连接。
	Graph []byte
}

// buildVectorIndex 生成文本块向量，使用 HNSW 库默认参数构建并导出索引。
func (i *indexer) buildVectorIndex(ctx context.Context, chunks []Chunk) (VectorIndex, error) {
	graph := hnsw.NewGraph[string]()
	dimensions := 0
	seen := make(map[string]bool, len(chunks))
	for _, chunk := range chunks {
		if err := ctx.Err(); err != nil {
			return VectorIndex{}, err
		}
		if chunk.ID == "" || seen[chunk.ID] {
			return VectorIndex{}, fmt.Errorf("chunk id must be nonempty and unique: %q", chunk.ID)
		}
		seen[chunk.ID] = true
		input := chunk.Content
		if heading := strings.Join(chunk.HeadingPath, " > "); heading != "" {
			input = heading + "\n" + input
		}
		vector, err := i.embedder.Embed(ctx, input)
		if err != nil {
			return VectorIndex{}, err
		}
		if len(vector) == 0 {
			return VectorIndex{}, fmt.Errorf("empty embedding for chunk %s", chunk.ID)
		}
		if dimensions == 0 {
			dimensions = len(vector)
		}
		if len(vector) != dimensions {
			return VectorIndex{}, fmt.Errorf("embedding dimension mismatch for chunk %s: got %d, want %d", chunk.ID, len(vector), dimensions)
		}
		var norm float64
		for _, value := range vector {
			v := float64(value)
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return VectorIndex{}, fmt.Errorf("non-finite embedding for chunk %s", chunk.ID)
			}
			norm += v * v
		}
		// 默认使用余弦距离，零向量没有有效的余弦相似度。
		if norm == 0 {
			return VectorIndex{}, fmt.Errorf("zero embedding for chunk %s", chunk.ID)
		}
		graph.Add(hnsw.Node[string]{Key: chunk.ID, Value: append([]float32(nil), vector...)})
	}
	if err := ctx.Err(); err != nil {
		return VectorIndex{}, err
	}
	var encoded bytes.Buffer
	if err := graph.Export(&encoded); err != nil {
		return VectorIndex{}, fmt.Errorf("export vector index: %w", err)
	}
	return VectorIndex{Dimensions: dimensions, Graph: encoded.Bytes()}, nil
}
