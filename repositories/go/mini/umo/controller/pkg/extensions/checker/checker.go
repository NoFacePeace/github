package checker

import "nofacepeace.github.io/controller/pkg/model"

type Checker interface {
	Check(args ...any) (model.CheckerResult, string, error)
	GetName() string
}
