package repository

import "context"

// Service 定义仓库模块对外提供的业务接口。
type Service interface {
	Create(ctx context.Context, request CreateRepositoryRequest) (Repository, error)

	// Search 按名称、服务归属和描述检索仓库，返回分页结果，具体逻辑待实现。
	// 服务启动时可预加载有效仓库及描述向量，查询时按以下流程处理：
	// 1. 校验查询文本和分页参数，默认第 1 页、每页 20 条，每页最多 100 条。
	//    查询为空时直接分页浏览仓库，跳过向量生成和相关性评分。
	// 2. 读取所有有效仓库的元数据，与缓存比较，增量加载新增或变化项，移除失效项。
	// 3. 复用描述、模型及维度一致的描述向量，补齐缺失或过期的向量；空描述跳过向量生成。
	// 4. 为查询文本生成一个向量；向量能力未接入或生成失败时降级为关键词匹配。
	//    请求取消、元数据或缓存读取失败时直接返回错误。
	// 5. 遍历全部有效仓库，分别计算名称、服务归属的文字匹配，以及描述向量余弦相似度。
	//    文字匹配统一大小写和分隔符，支持整体、空白拆词及名称拼写近似匹配，中文不逐字拆分。
	// 6. 名称或服务归属命中即可保留；仅描述语义命中时，相似度至少为 0.45。
	//    降级时使用描述关键词匹配，并在返回结果中标识降级状态。
	// 7. 按名称命中、服务归属命中、纯描述命中的优先级选择主评分，文字命中仍可叠加少量语义分。
	// 8. 所有仓库评估完成后，按得分降序排序，同分按名称、ID 排序，最后分页返回。
	// 上层入口可根据本页仓库补充近期版本，不参与本方法的候选评分和分页。
	Search(ctx context.Context, request SearchRepositoryRequest) (SearchRepositoryResult, error)
}

// CreateRepositoryRequest 表示创建代码仓库记录的请求。
type CreateRepositoryRequest struct {
	Name      string
	GitURL    string
	ProjectID string
}

// SearchRepositoryRequest 表示仓库搜索与分页请求，Query 为空时浏览仓库。
type SearchRepositoryRequest struct {
	Query   string
	Page    int
	PerPage int
}

// SearchRepositoryResult 表示仓库搜索的分页结果。
type SearchRepositoryResult struct {
	Items      []RepositorySearchHit
	Page       int
	PerPage    int
	Total      int
	TotalPages int
	// SearchMode 表示混合搜索或关键词降级，分别为 hybrid、keyword_fallback。
	SearchMode              string
	SemanticSearchAvailable bool
}

// RepositorySearchHit 表示命中的仓库、最终得分及命中原因。
type RepositorySearchHit struct {
	Repository Repository
	Score      float64
	MatchedBy  []string
}
