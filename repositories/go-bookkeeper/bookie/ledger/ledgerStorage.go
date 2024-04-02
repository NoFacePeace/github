package ledger

import "os"

type LedgerStorage interface {
	Start()
}

type LedgerDirsManager struct {
}

func NewLedgerDirsManager() *LedgerDirsManager {
	return &LedgerDirsManager{}
}

func (m *LedgerDirsManager) GetAllLedgerDirs() []os.File {
	return []os.File{}
}
