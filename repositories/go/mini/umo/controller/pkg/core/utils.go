package core

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	umov1 "nofacepeace.github.io/controller/api/v1"
)

func generateNodeName(cls, nodeset string, idx int) string {
	return fmt.Sprintf("%s-%s-%d", cls, nodeset, idx)
}

func generateGroupName(filter *umov1.GrayFilter) string {
	return fmt.Sprintf("stage-%s-%s-%d", filter.NodeType, filter.NodeSetName, filter.Stage)
}

func errorsToError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	var ret error
	for _, err := range errs {
		ret = errors.Join(ret, err)
	}
	return ret
}

func generateTplName(middleware, tplVersion string) string {
	return fmt.Sprintf("%s_%s", middleware, tplVersion)
}

func generateVersion() string {
	return fmt.Sprint(time.Now().Unix())
}

func isTrue(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}
