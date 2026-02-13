package core

import (
	"context"
	"time"
)

type Model struct {
	Ctx         context.Context
	Time        time.Time
	Operation   string
	ClusterId   string
	ClusterName string
	Err         error
	Loggers     []string
	Extras      map[string]string
}

type Builder struct {
	Model *Model
	
}
