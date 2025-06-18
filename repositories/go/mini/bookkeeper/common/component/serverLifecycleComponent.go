package component

type ServerLifecycleComponent struct {
	Server Server
}

func NewServerLifecycleComponent(s Server) *ServerLifecycleComponent {
	return &ServerLifecycleComponent{
		Server: s,
	}
}

func (s *ServerLifecycleComponent) Start() {
	s.Server.DoStart()
}

type Server interface {
	DoStart()
}
