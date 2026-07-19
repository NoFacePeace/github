package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

// parseNodeIndex 提取节点名称中最后一个连字符后的数字索引。
func parseNodeIndex(name string) (int, error) {
	pos := strings.LastIndexByte(name, '-')
	if pos < 0 || pos == len(name)-1 {
		return 0, fmt.Errorf("invalid node name %q: missing index", name)
	}

	index, err := strconv.Atoi(name[pos+1:])
	if err != nil {
		return 0, fmt.Errorf("invalid node name %q: parse index: [%w]", name, err)
	}
	return index, nil
}

// parseNodeSetName 从节点名称中提取节点集名称。
func parseNodeSetName(cls, node string) (string, error) {
	prefix := cls + "-"
	if cls == "" || !strings.HasPrefix(node, prefix) {
		return "", fmt.Errorf("invalid node name %q: missing class prefix %q", node, prefix)
	}

	remaining := strings.TrimPrefix(node, prefix)
	pos := strings.LastIndexByte(remaining, '-')
	if pos <= 0 {
		return "", fmt.Errorf("invalid node name %q: missing node set name", node)
	}
	if pos == len(remaining)-1 {
		return "", fmt.Errorf("invalid node name %q: missing index", node)
	}
	if _, err := strconv.Atoi(remaining[pos+1:]); err != nil {
		return "", fmt.Errorf("invalid node name %q: parse index: [%w]", node, err)
	}
	return remaining[:pos], nil
}
