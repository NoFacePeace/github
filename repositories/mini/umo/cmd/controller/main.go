package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"umo/controller/healthcheck"
	"umo/controller/watch"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	watchURL := os.Getenv("WATCH_URL")
	var healthConfig healthcheck.Config
	if watchURL != "" {
		kubeConfig, err := config.GetConfig()
		if err != nil {
			log.Fatalf("load Kubernetes configuration: %v", err)
		}
		healthConfig.K8sClient, err = client.New(kubeConfig, client.Options{})
		if err != nil {
			log.Fatalf("create Kubernetes client: %v", err)
		}
	}

	// 具体检查器与配置在对应功能实现时接入。
	manager, err := healthcheck.NewManager(nil, healthConfig)
	if err != nil {
		stop()
		log.Printf("create healthcheck manager failed: %v", err)
		os.Exit(1)
	}
	watchClient := watch.NewClient(watchURL, manager.SubmitEvent)

	var wg sync.WaitGroup
	wg.Go(func() {
		manager.Run(ctx)
	})
	wg.Go(func() {
		watchClient.Run(ctx)
	})
	wg.Wait()
	stop()
}
