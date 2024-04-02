package embeddedserver

import (
	"github.com/NoFacePeace/github/repositories/go-bookkeeper/bookie"
	"github.com/NoFacePeace/github/repositories/go-bookkeeper/bookie/ledger"
	"github.com/NoFacePeace/github/repositories/go-bookkeeper/common/component"
)

type EmbeddedServer struct {
	Config                  Config
	LifecycleComponentStack *component.LifecycleComponentStack
}

func New(c Config) *EmbeddedServer {
	stack := &component.LifecycleComponentStack{}

	// 1. 构建统计提供者
	// 2. 构建元数据驱动
	// 3. 构建 ledger 管理者
	storage := ledger.NewDbLedgerStorage()

	// 4. 构建 bookie
	ldm := &ledger.LedgerDirsManager{}
	var b bookie.Bookie
	if c.IsForceReadOnlyBookie() {

	} else {
		b = bookie.NewBookieImpl(&bookie.Config{}, storage, ldm)
	}

	// 5. 构建 bookie 服务端
	bs := bookie.NewService(b)
	slc := component.NewServerLifecycleComponent(bs)
	stack.AddComponent(slc)
	// 6. 构建自动恢复器
	// 7. 构建数据完整性校验服务
	// 8. 构建 http 服务
	// 9. 构建扩展服务
	return &EmbeddedServer{
		Config:                  c,
		LifecycleComponentStack: stack,
	}
}
func (s *EmbeddedServer) GetLifecycleComponentStack() *component.LifecycleComponentStack {
	return s.LifecycleComponentStack
}

type Config struct {
}

// IsForceReadOnlyBookie bookie 是否强制只读启动，默认 false，强制只读启动不会回放 journal
// https://bookkeeper.apache.org/docs/reference/config#read-only-mode-support
func (c *Config) IsForceReadOnlyBookie() bool {
	return false
}
