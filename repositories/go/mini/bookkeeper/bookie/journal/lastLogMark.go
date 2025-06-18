package journal

import "github.com/NoFacePeace/github/repositories/go/mini/bookkeeper/bookie/ledger"

type LastLogMark struct {
	LogFileId     int
	LogFileOffset int
}

func NewLogMark(id, offset int) *LastLogMark {
	return &LastLogMark{LogFileId: id, LogFileOffset: offset}
}

// 读取 ledger 目录下所有的 lastMark 文件, 设置 logFileId, logFileOffset
func (m *LastLogMark) ReadLog(ldm *ledger.LedgerDirsManager, name string) {
}
