package core

import (
	"regexp"
	"slices"

	corev1 "k8s.io/api/core/v1"
	umov1 "nofacepeace.github.io/controller/api/v1"
	"nofacepeace.github.io/controller/pkg/config"
)

const (
	GrayFilterTypePercent = "percent"
	GrayFilterTypeRegexp  = "regexp"
	GrayFilterTypeLabel   = "label"
)

func getFinalNodeSet(cls *umov1.Middleware, spec *umov1.NodeSetSpec, pod *corev1.Pod, idx int) (*umov1.NodeSetSpec, *umov1.GrayFilter) {
	specs := map[string]umov1.NodeSetSpec{}
	for _, gray := range cls.Spec.Gray {
		specs[gray.Name] = gray
	}
	if _, ok := specs[spec.Name]; !ok {
		return spec, nil
	}
	filter, ok := getGrayFilter(cls, spec, pod, idx)
	if !ok {
		return spec, nil
	}
	gray := specs[spec.Name]
	gray.NodeCounts = spec.NodeCounts
	gray.Annotations = spec.Annotations
	gray.Labels = spec.Labels
	return &gray, filter
}

func getGrayFilter(cls *umov1.Middleware, spec *umov1.NodeSetSpec, pod *corev1.Pod, idx int) (*umov1.GrayFilter, bool) {
	for _, filter := range cls.Spec.GrayFilters {
		if filter.NodeType != "" && filter.NodeType != spec.Type {
			continue
		}
		if filter.NodeSetName != "" && filter.NodeSetName != spec.Name {
			continue
		}
		switch filter.Type {
		case GrayFilterTypePercent:
			cnt := spec.NodeCounts[config.Get().Eks.Id].Count
			offset := spec.NodeCounts[config.Get().Eks.Id].Offset
			if cnt <= 0 || filter.Percent <= 0 {
				continue
			}
			end := offset + cnt*filter.Percent/100
			if idx < end {
				return &filter, true
			}
		case GrayFilterTypeRegexp:
			if filter.Regexp == "" {
				continue
			}
			if ok, _ := regexp.MatchString(filter.Regexp, spec.Name); ok {
				return &filter, true
			}
		case GrayFilterTypeLabel:
			if pod == nil || pod.Labels == nil || filter.MatchLabels == nil {
				continue
			}
			for k, vs := range filter.MatchLabels {
				if _, ok := pod.Labels[k]; !ok {
					continue
				}
				value := pod.Labels[k]
				if slices.Contains(vs, value) {
					return &filter, true
				}
			}
		}
	}
	return nil, false
}
