package ledger

type LedgerDirsMonitor struct{}

func NewLedgerDirsMonitor() *LedgerDirsMonitor {
	return &LedgerDirsMonitor{}
}

func (m *LedgerDirsMonitor) Start() {
	go func() {

	}()
}
