package bookie

type Service struct {
	Bookie Bookie
}

func NewService(b Bookie) *Service {
	return &Service{
		Bookie: b,
	}
}
func (s *Service) DoStart() {
	// 1. 启动 bookie
	s.Bookie.Start()

	// 2. 启动 netty
}
