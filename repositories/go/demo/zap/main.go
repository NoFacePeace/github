package main

import (
	"fmt"

	"go.uber.org/zap"
)

type T struct {
	Name string
}

func main() {
	t := T{}
	t.Name = "name"
	fmt.Printf("%+v", t)

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger.Info("hh", zap.Any("t", t))
}
