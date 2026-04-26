package core

import (
	"context"
	"time"
)

const (
	OperationTypeCheckCluster = "CheckCluster"
	OperationTypePreCheck     = "PreCheck"
	OperationTypePostCheck    = "PostCheck"
)

type OperationModel struct {
	Ctx         context.Context
	Time        time.Time
	Operation   string
	ClusterId   string
	ClusterName string
	Err         error
	Loggers     []string
	Extras      map[string]string
}

type OperationManager struct {
}

func (o *OperationManager) NewOperationBuilder(ctx context.Context, cls, op string) *OperationBuilder {
	return &OperationBuilder{}
}

type OperationBuilder struct {
}

func (o *OperationBuilder) WithError(err error) *OperationBuilder {
	return o
}

func (o *OperationBuilder) WithNodeName(name string) *OperationBuilder {
	return o
}

func (o *OperationBuilder) Report() {
}
