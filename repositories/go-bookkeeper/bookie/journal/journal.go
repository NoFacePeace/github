package journal

import (
	"os"
	"strconv"

	"github.com/NoFacePeace/github/repositories/go-bookkeeper/bookie/ledger"
	"github.com/NoFacePeace/github/repositories/go-bookkeeper/common/collections"
)

const (
	LAST_MARK_DEFAULT_NAME = "lastMark"
)

type Journal struct {
	Queue              collections.BatchedBlockingQueue
	ForceWriteRequests collections.BatchedBlockingQueue
	LastLogMark        *LastLogMark
	Directory          os.File
}

func NewJournal(idx int, file os.File, c *Config, ldm *ledger.LedgerDirsManager) *Journal {
	// 队列
	var queue, forceWriteRequests collections.BatchedBlockingQueue
	// 自旋
	if c.IsBusyWaitEnabled() {
		queue = collections.NewBlockingMpscQueue()
		forceWriteRequests = collections.NewBlockingMpscQueue()
	} else {
		queue = collections.NewBatchedArrayBlockingQueue()
		forceWriteRequests = collections.NewBatchedArrayBlockingQueue()
	}

	dir := file

	// log mark lastMark 文件, 格式 文件名 + offset
	lastLogMark := NewLogMark(0, 0)
	lastMarkFileName := LAST_MARK_DEFAULT_NAME
	if len(c.GetJournalDirs()) != 0 {
		lastMarkFileName += "." + strconv.Itoa(idx)
	}
	lastLogMark.ReadLog(ldm, lastMarkFileName)
	return &Journal{
		Queue:              queue,
		ForceWriteRequests: forceWriteRequests,
		LastLogMark:        lastLogMark,
		Directory:          dir,
	}
}

func (j *Journal) ListJournalIds(id int) []int {
	// 读取所有 journal 目录下文件，文件名大于 id
	return []int{}
}

func (j *Journal) Start() {
	go j.Run()
}

func (j *Journal) Run() {

}

func (j *Journal) LogAddEntry() {
	j.Queue.
}

type Config struct {
}

func (c *Config) GetJournalDirs() []os.File {
	return []os.File{}
}

// 自旋
func (c *Config) IsBusyWaitEnabled() bool {
	return false
}
