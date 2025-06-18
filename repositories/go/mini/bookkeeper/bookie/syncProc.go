package bookie

type SyncProc struct{}

func NewSyncProc() *SyncProc {
	return &SyncProc{}
}

func (s *SyncProc) RequestFlush() {
	s.Flush()
}

func (s *SyncProc) Start() {

}

func (s *SyncProc) Flush() {

}
