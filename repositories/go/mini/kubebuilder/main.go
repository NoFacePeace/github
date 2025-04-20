package main

import (
	"log/slog"

	"github.com/NoFacePeace/github/repositories/go/mini/kubebuilder/controller"
	"github.com/NoFacePeace/github/repositories/go/mini/kubebuilder/manager"
	"github.com/NoFacePeace/github/repositories/go/utils/log"
	"github.com/NoFacePeace/github/repositories/go/utils/signal"
)

var (
	NewManager = manager.New
)

func main() {
	// 初始化日志
	log.Init()

	// 创建管理器
	mgr, err := NewManager()
	if err != nil {
		slog.Error("new manager error", "error", err)
		return
	}

	// 创建 reconciler，绑定 manager
	reconciler := &controller.ExampleReconciler{}
	if err := reconciler.SetupWithManager(mgr); err != nil {
		slog.Error("reconciler setup with manager error", "error", err)
		return
	}

	// 启动管理器
	if err := mgr.Start(signal.SetupSignalHandler()); err != nil {
		slog.Error("manager start error", "error", err)
		return
	}
}
