package server

type aeApiEpoll struct{}

type aeApiEpollState struct {
	event []epollEvent
	epFd  int
}

type epollEvent struct{}

func (a *aeApiEpoll) Create(eventLoop *aeEventLoop) int {
	state := &aeApiEpollState{}
	state.event = []epollEvent{}
	state.epFd = EpollCreate(1024)
	eventLoop.apiData = state
	return 0
}

func EpollCreate(size int) (fd int) {
	return 0
}
