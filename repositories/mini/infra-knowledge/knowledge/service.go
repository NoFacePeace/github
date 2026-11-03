package knowledge

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// Service 定义知识库模块对外提供的业务接口。
type Service interface {
	Create(ctx context.Context, request CreateDocumentRequest) (Document, error)
	Reindex(ctx context.Context, documentID string) error
	Retrieve(ctx context.Context, request RetrieveRequest) (RetrieveResult, error)
}

// service 实现知识库模块的业务接口。
type service struct {
	store          Store
	indexer        Indexer
	embedder       Embedder
	retrieveConfig RetrieveConfig
}

// NewService 创建知识库服务。
func NewService(store Store, indexer Indexer, embedder Embedder, config RetrieveConfig) Service {
	if config.DefaultLimit <= 0 {
		config.DefaultLimit = 10
	}
	if config.MaxLimit <= 0 {
		config.MaxLimit = 100
	}
	config.DefaultLimit = min(config.DefaultLimit, config.MaxLimit)
	if config.CandidateMultiplier <= 0 {
		config.CandidateMultiplier = 3
	}
	if config.MinBranchCandidates <= 0 {
		config.MinBranchCandidates = 50
	}
	if config.RRFK <= 0 {
		config.RRFK = 60
	}
	return &service{store: store, indexer: indexer, embedder: embedder, retrieveConfig: config}
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

// Retrieve 检索知识文档，具体逻辑待实现。
func (s *service) Retrieve(ctx context.Context, request RetrieveRequest) (RetrieveResult, error) {
	// 1. 校验项目 ID、查询文本、过滤条件和结果数量，仅检索指定项目下的文档。
	// 项目、查询文本校验及其他过滤字段待补充。
	limit := request.Limit
	if limit == 0 {
		limit = s.retrieveConfig.DefaultLimit
	}
	if limit < 1 || limit > s.retrieveConfig.MaxLimit {
		return RetrieveResult{}, fmt.Errorf("retrieve limit must be between 1 and %d", s.retrieveConfig.MaxLimit)
	}
	branchLimit := max(limit*s.retrieveConfig.CandidateMultiplier, s.retrieveConfig.MinBranchCandidates)
	documents, err := s.store.ListByProject(ctx, request.ProjectID)
	if err != nil {
		return RetrieveResult{}, err
	}
	if len(documents) == 0 {
		return RetrieveResult{}, nil
	}
	documentIDs := make([]string, 0, len(documents))
	for _, doc := range documents {
		documentIDs = append(documentIDs, doc.ID)
	}

	// 2. 加载该项目文档的已发布全文索引和向量索引，合并为项目范围的运行时索引。
	//    合并使用已有倒排表、统计和向量，不重新生成文档向量。
	fulltextIndexes, err := s.store.GetFulltextIndexes(ctx, documentIDs)
	if err != nil {
		return RetrieveResult{}, err
	}
	vectorIndexes, err := s.store.GetVectorIndexes(ctx, documentIDs)
	if err != nil {
		return RetrieveResult{}, err
	}
	fulltextInputs := make([]FulltextIndex, 0, len(fulltextIndexes))
	vectorInputs := make([]VectorIndex, 0, len(vectorIndexes))
	// 按文档列表顺序收集索引，跳过缺失项，避免 map 遍历顺序影响合并。
	for _, documentID := range documentIDs {
		if index, ok := fulltextIndexes[documentID]; ok {
			fulltextInputs = append(fulltextInputs, index)
		}
		if index, ok := vectorIndexes[documentID]; ok {
			vectorInputs = append(vectorInputs, index)
		}
	}
	fulltextIndex, err := mergeFulltextIndexes(fulltextInputs)
	if err != nil {
		return RetrieveResult{}, err
	}
	vectorIndex, err := mergeVectorIndexes(vectorInputs)
	if err != nil {
		return RetrieveResult{}, err
	}

	// 3. 并行执行全文检索和向量检索，共用项目范围和每路候选数量上限。
	// 向量模型兼容性校验待补充，不设置额外超时，沿用调用方上下文。
	type branchResult struct {
		candidates []Candidate
		err        error
	}
	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	fulltextCh := make(chan branchResult, 1)
	vectorCh := make(chan branchResult, 1)
	go func() {
		candidates, err := fulltextIndex.Search(workCtx, request.Query, branchLimit)
		fulltextCh <- branchResult{candidates: candidates, err: err}
	}()
	go func() {
		vector, err := s.embedder.Embed(workCtx, request.Query)
		if err != nil {
			vectorCh <- branchResult{err: err}
			return
		}
		candidates, err := vectorIndex.Search(workCtx, vector, branchLimit)
		vectorCh <- branchResult{candidates: candidates, err: err}
	}()

	// 4. 全文失败直接返回；仅向量明确不可用时降级，其他错误及请求取消直接返回。
	var fulltextCandidates, vectorCandidates []Candidate
	var degraded []string
	for completed := 0; completed < 2; completed++ {
		select {
		case <-ctx.Done():
			return RetrieveResult{}, ctx.Err()
		case result := <-fulltextCh:
			if result.err != nil {
				return RetrieveResult{}, result.err
			}
			fulltextCandidates = result.candidates
		case result := <-vectorCh:
			if result.err != nil {
				if !errors.Is(result.err, ErrVectorUnavailable) {
					return RetrieveResult{}, result.err
				}
				degraded = append(degraded, "vector")
			} else {
				vectorCandidates = result.candidates
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return RetrieveResult{}, err
	}

	// 5. 使用 RRF 按两路排名融合候选，同一文本块合并计分，再排序并限制结果数量。
	//    不直接相加 BM25 分数与向量相似度；如需特定场景重排，在此阶段执行。
	candidates := fuseCandidates(fulltextCandidates, vectorCandidates, limit, s.retrieveConfig.RRFK)

	// 6. 根据命中文本块关联文档和章节，补充章节正文、命中片段及相邻章节上下文。
	hits, err := buildRetrieveHits(ctx, documents, fulltextIndex.Chunks, candidates)
	if err != nil {
		return RetrieveResult{}, err
	}

	// 后续流程待实现：
	// 7. 核对本次检索期间发布版本是否变化；变化时有限次重试，持续变化则返回错误。
	// 8. 返回检索结果、实际使用的范围或版本信息，以及是否发生向量降级。
	return RetrieveResult{Candidates: candidates, Hits: hits, Degraded: degraded}, nil
}

// buildRetrieveHits 使用索引中保存的文本块还原章节，不依赖当前切分配置。
func buildRetrieveHits(ctx context.Context, documents []Document, chunks []Chunk, candidates []Candidate) ([]RetrieveHit, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	hits := make([]RetrieveHit, 0, len(candidates))
	if len(candidates) == 0 {
		return hits, nil
	}
	documentsByID := make(map[string]Document, len(documents))
	for _, doc := range documents {
		documentsByID[doc.ID] = doc
	}
	chunksByDocument := make(map[string][]Chunk)
	chunksByID := make(map[string]Chunk, len(chunks))
	for _, chunk := range chunks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if _, ok := documentsByID[chunk.DocumentID]; !ok {
			return nil, fmt.Errorf("chunk %s references document outside retrieval scope: %s", chunk.ID, chunk.DocumentID)
		}
		if _, exists := chunksByID[chunk.ID]; exists {
			return nil, fmt.Errorf("duplicate chunk id: %s", chunk.ID)
		}
		chunk.HeadingPath = append([]string(nil), chunk.HeadingPath...)
		chunk.SectionPaths = append([]string(nil), chunk.SectionPaths...)
		chunksByID[chunk.ID] = chunk
		chunksByDocument[chunk.DocumentID] = append(chunksByDocument[chunk.DocumentID], chunk)
	}

	chaptersByDocument := make(map[string][]RetrieveChapter)
	chapterPositions := make(map[string]map[string]int)
	chunkPositions := make(map[string]int, len(chunks))
	for documentID, ordered := range chunksByDocument {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		sort.Slice(ordered, func(a, b int) bool { return ordered[a].Position < ordered[b].Position })
		positions := make(map[string]int)
		chapters := make([]RetrieveChapter, 0)
		for start := 0; start < len(ordered); {
			first := ordered[start]
			if _, exists := positions[first.ChapterID]; exists {
				return nil, fmt.Errorf("noncontiguous chapter chunks: %s", first.ChapterID)
			}
			chapter := RetrieveChapter{ID: first.ChapterID, HeadingPath: append([]string(nil), first.HeadingPath...)}
			var content strings.Builder
			end := start
			for end < len(ordered) && ordered[end].ChapterID == first.ChapterID {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				chunk := ordered[end]
				if end > 0 && chunk.Position == ordered[end-1].Position {
					return nil, fmt.Errorf("duplicate chunk position in document %s: %d", documentID, chunk.Position)
				}
				chunkPositions[chunk.ID] = end
				content.WriteString(chunk.Content)
				chapter.TokenCount += chunk.TokenCount
				common := 0
				for common < len(chapter.HeadingPath) && common < len(chunk.HeadingPath) && chapter.HeadingPath[common] == chunk.HeadingPath[common] {
					common++
				}
				chapter.HeadingPath = chapter.HeadingPath[:common]
				end++
			}
			chapter.Content = content.String()
			positions[chapter.ID] = len(chapters)
			chapters = append(chapters, chapter)
			start = end
		}
		chaptersByDocument[documentID] = chapters
		chapterPositions[documentID] = positions
	}

	// 按融合排名组装结果，相邻上下文只在所属文档内查找，不参与排序。
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		chunk, ok := chunksByID[candidate.ChunkID]
		if !ok {
			return nil, fmt.Errorf("candidate chunk not found: %s", candidate.ChunkID)
		}
		doc := documentsByID[chunk.DocumentID]
		ordered := chunksByDocument[doc.ID]
		chapters := chaptersByDocument[doc.ID]
		chapterPosition := chapterPositions[doc.ID][chunk.ChapterID]
		hit := RetrieveHit{
			DocumentID: doc.ID, DocumentFileName: doc.FileName, DocumentTitle: doc.Title,
			ProjectID: doc.ProjectID, ProjectName: doc.ProjectName,
			Chunk: chunk, Chapter: chapters[chapterPosition], Score: candidate.Score,
		}
		position := chunkPositions[chunk.ID]
		if position > 0 {
			previous := ordered[position-1]
			hit.PreviousChunk = &previous
		}
		if position+1 < len(ordered) {
			next := ordered[position+1]
			hit.NextChunk = &next
		}
		if chapterPosition > 0 {
			previous := chapters[chapterPosition-1]
			hit.PreviousChapter = &previous
		}
		if chapterPosition+1 < len(chapters) {
			next := chapters[chapterPosition+1]
			hit.NextChapter = &next
		}
		hits = append(hits, hit)
	}
	return hits, nil
}

// ErrVectorUnavailable 表示向量服务明确不可用，允许降级为全文检索。
var ErrVectorUnavailable = errors.New("vector retrieval unavailable")

// fuseCandidates 按文本块 ID 合并两路候选，计算 RRF 分数、排序并截取结果，具体逻辑待实现。
func fuseCandidates(fulltext, vector []Candidate, limit int, k float64) []Candidate {
	panic("not implemented")
}

// RetrieveConfig 表示检索的结果数量和排名融合配置。
type RetrieveConfig struct {
	// DefaultLimit 表示请求未指定数量时的返回上限，建议值为 10。
	DefaultLimit int
	// MaxLimit 表示请求允许指定的最大返回上限，建议值为 100。
	MaxLimit int
	// CandidateMultiplier 表示每路候选数量相对于返回上限的倍数，建议值为 3。
	CandidateMultiplier int
	// MinBranchCandidates 表示每路候选数量的下限，建议值为 50。
	MinBranchCandidates int
	// RRFK 表示倒数排名融合的平滑参数，建议值为 60。
	RRFK float64
}

// RetrieveRequest 表示知识文档检索请求。
type RetrieveRequest struct {
	// ProjectID 限定本次检索的项目范围。
	ProjectID string
	Query     string
	// Limit 表示最终返回数量上限，0 使用检索配置中的默认值。
	Limit int
}

// RetrieveResult 表示知识文档检索结果。
type RetrieveResult struct {
	Candidates []Candidate
	Hits       []RetrieveHit
	// Degraded 记录不可用的检索分支，目前仅包含 vector。
	Degraded []string
}

// RetrieveHit 表示按融合排名返回的命中文本块及文档、章节上下文。
type RetrieveHit struct {
	DocumentID       string
	DocumentFileName string
	DocumentTitle    string
	ProjectID        string
	ProjectName      string
	Chunk            Chunk
	Chapter          RetrieveChapter
	Score            float64
	// 相邻上下文仅来自同一文档，文档边界处对应字段为空。
	PreviousChunk   *Chunk
	NextChunk       *Chunk
	PreviousChapter *RetrieveChapter
	NextChapter     *RetrieveChapter
}

// RetrieveChapter 表示从已索引文本块还原的完整章节上下文。
type RetrieveChapter struct {
	ID          string
	HeadingPath []string
	Content     string
	TokenCount  int
}

// CreateDocumentRequest 表示创建知识文档的请求。
type CreateDocumentRequest struct {
	FileName    string
	Title       string
	Content     string
	ProjectID   string
	ProjectName string
}
