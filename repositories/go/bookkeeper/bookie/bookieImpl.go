package bookie

import (
	"github.com/NoFacePeace/github/repositories/go/bookkeeper/bookie/journal"
	"github.com/NoFacePeace/github/repositories/go/bookkeeper/bookie/ledger"
)

type BookieImpl struct {
	DirsMonitor   *ledger.LedgerDirsMonitor
	SyncProc      *SyncProc
	LedgerStorage ledger.LedgerStorage
	Journals      []*journal.Journal
}

func NewBookieImpl(c *Config, l ledger.LedgerStorage, ldm *ledger.LedgerDirsManager) *BookieImpl {
	// Journal
	journals := []*journal.Journal{}
	journalDirectories := c.GetJournalDirs()
	for k := range journalDirectories {
		journals = append(journals, journal.NewJournal(k, journalDirectories[k], &journal.Config{}, ldm))
	}
	// SyncThread
	// isDbLedgerStorage := fmt.Sprintf("%T", l) == ""
	return &BookieImpl{
		LedgerStorage: l,
		Journals:      journals,
	}
}

func (b *BookieImpl) Start() {
	// 1. 启动磁盘检查器 goroutine
	b.DirsMonitor.Start()

	// 2. 回放 journals
	b.ReadJournal()

	//3. 同步刷盘
	b.SyncProc.RequestFlush()

	// 4. 启动 sync thread
	b.SyncProc.Start()
	// 5. 启动 bookie thread，本质是启动 journal thread
	go b.Run()
	// 6. 启动 ledger storage
	b.LedgerStorage.Start()
	// 7. 注册 zookeeper
}

func (b *BookieImpl) Run() {
	for _, j := range b.Journals {
		j.Start()
	}
}

// 将 Journal 写到 ledgerStorage
func (b *BookieImpl) ReadJournal() {
	for _, j := range b.Journals {
		b.replay(j)
	}
}

func (b *BookieImpl) replay(j *journal.Journal) {
	// 1. 获取最新 lastMark
	// 2. 遍历大于 lastMark 的 txn 日志
	// 3. 写入到 ledgerStorage
	// 4. 更新 lastMark
}

func (b *BookieImpl) AddEntry() {
	b.addEntryInternal()
}

func (b *BookieImpl) addEntryInternal() {
	// 1. 先写 ledger storage
	// 2. 后写 journal
	b.getJournal(0).LogAddEntry()
}

// getJournal 根据 ledger id 获得 journal
func (b *BookieImpl) getJournal(id int) *journal.Journal {
	return b.Journals[id]
}
