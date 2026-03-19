package core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	umov1 "nofacepeace.github.io/controller/api/v1"
)

const (
	GrayFilterTypePercent = "percent"
	GrayFilterTypeRegexp  = "regexp"
	GrayFilterTypeLabel   = "label"
)

func GetFinalNodeSet(ctx context.Context, cls *umov1.Middleware, pod *corev1.Pod, idx int) (umov1.NodeSetSpec, umov1.GrayFilter) {

	return umov1.NodeSetSpec{}, umov1.GrayFilter{}
}

func GetGrayFilter(cls *umov1.Middleware, spec *umov1.NodeSetSpec, pod *corev1.Pod, idx int) (umov1.GrayFilter, bool) {
	return umov1.GrayFilter{}, false
	// for _, filter := range cls.Spec.GrayFilters {
	// 	if filter.NodeSetName != "" && filter.NodeSetName != spec.Name {
	// 		continue
	// 	}
	// 	if filter.NodeType != "" && filter.NodeType != spec.Type {
	// 		continue
	// 	}
	// 	switch filter.Type {
	// 	case GrayFilterTypePercent:
	// 		cnts := spec.NodeCounts[config.Get().Eks.Id]
	// 		end := cnts.Offset + cnts.Count*filter.Percent/100
	// 		if idx < end {
	// 			return filter, true
	// 		}
	// 	case GrayFilterTypeRegexp:
	// 		if ok, _ := regexp.MatchString(filter.Regexp, pod.Name); ok {
	// 			return filter, true
	// 		}
	// 	case GrayFilterTypeLabel:

	// 	}
	// }
	return umov1.GrayFilter{}, false
}
