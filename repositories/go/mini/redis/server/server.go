package server

import (
	"os"
)

type redisServer struct {
	el *aeEventLoop

	port      int
	listeners [connTypeMax]connListener
}

var server redisServer

func Main() {
	// 初始化连接类型
	connTypeInitialize()
	// 初始化服务器
	initServer()

	// 初始化监听
	initListeners()

	initServerLast()

	aeMain(server.el)
}

func initServer() {
	// 创建事件循环器
	server.el = aeCreateEventLoop()
}

func initListeners() {
	if server.port != 0 {
		// conIdx := connectionIndexByType
	}
	for i := 0; i < connTypeMax; i++ {
		listener := &server.listeners[i]
		if listener.ct == nil {
			continue
		}
		if connListen(listener) != nil {
			os.Exit(1)
		}
		if createSocketAcceptHandler(listener, connAcceptHandler(listener.ct)) != nil {
			os.Exit(1)
		}
	}
}

func createSocketAcceptHandler(sfd *connListener, acceptHandler aeFileProc) error {
	for i := 0; i < sfd.count; i++ {

	}
	return nil
}

func initServerLast() {
	// 初始化线程 IO
	initThreadedIO()
}
