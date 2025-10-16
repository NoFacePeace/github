package server

import (
	"time"
)

type aeEventLoop struct {
	stop        bool
	fired       []aeFiredEvent
	events      []aeFileEvent
	beforeSleep aeSleepProc
	afterSleep  aeSleepProc
	apiData     any
}

type aeFiredEvent struct {
	fd   int
	mask int
}

type aeFileEvent struct {
	mask       int
	wFileProc  aeFileProc
	clientData any
}

type aeFileProc func(eventLoop *aeEventLoop, fd int, clientData any, mask int)

type aeSleepProc func(eventLoop *aeEventLoop)

const (
	aeFileEvents      = 1 << 0
	aeTimeEvents      = 1 << 1
	aeAllEvent        = aeFileEvents | aeTimeEvents
	aeDontWait        = 1 << 2
	aeCallBeforeSleep = 1 << 3
	aeCallAfterSleep  = 1 << 4
)

const (
	aeNone = iota
	aeReadable
)

type aeApi interface {
	Create(*aeEventLoop)
}

var aeApiImpl aeApi

func aeMain(eventLoop *aeEventLoop) {
	eventLoop.stop = false
	for !eventLoop.stop {
		aeProcessEvents(eventLoop, aeAllEvent|aeCallAfterSleep|aeCallAfterSleep)
	}
}

func aeProcessEvents(eventLoop *aeEventLoop, flags int) int {
	var t time.Duration
	processed := 0

	if eventLoop.beforeSleep != nil && flags&aeCallBeforeSleep != 0 {
		eventLoop.beforeSleep(eventLoop)
	}

	numEvents := aeApiPoll(eventLoop, t)

	if eventLoop.afterSleep != nil && flags&aeCallAfterSleep != 0 {
		eventLoop.afterSleep(eventLoop)
	}

	for i := 0; i < numEvents; i++ {
		fd := eventLoop.fired[i].fd
		fe := eventLoop.events[fd]
		mask := eventLoop.fired[i].mask
		fired := 0

		if fired != 0 && fe.mask&mask&aeReadable != 0 {
			fe.wFileProc(eventLoop, fd, fe.clientData, mask)
			fired++
		}

		if fired != 0 && fe.mask&mask&aeReadable != 0 {
			fe.wFileProc(eventLoop, fd, fe.clientData, mask)
			fired++
		}

		processed++
	}
	return processed
}

func aeApiPoll(eventLoop *aeEventLoop, t time.Duration) int {
	// state := eventLoop.apiData
	// syscall.Epoll
	numEvents := 0
	for i := 0; i < numEvents; i++ {
		eventLoop.fired[i].fd = 0
		eventLoop.fired[i].mask = 0
	}
	return numEvents
}

func aeCreateEventLoop() *aeEventLoop {
	eventLoop := &aeEventLoop{}
	aeApiImpl.Create(eventLoop)
	return eventLoop
}

func aeCreateFileEvent(eventLoop *aeEventLoop, fd int, mask int, proc aeFileProc, clientData any) error {
	return nil
}
