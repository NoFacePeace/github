package main

import (
	"github.com/NoFacePeace/github/repositories/go/mini/bookkeeper/common/component"
	embeddedserver "github.com/NoFacePeace/github/repositories/go/mini/bookkeeper/embeddedServer"
)

func main() {
	// 0. 解析命令行
	// 1. 构建组件栈
	server := buildBookieServer()
	// 2. 启动服务
	component.StartComponent(server)
}

func buildBookieServer() *component.LifecycleComponentStack {
	server := embeddedserver.New(embeddedserver.Config{})
	return server.GetLifecycleComponentStack()
}
