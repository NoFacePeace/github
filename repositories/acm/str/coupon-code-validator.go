package str

import "sort"

func validateCoupons(code []string, businessLine []string, isActive []bool) []string {
	busi := map[string][]string{
		"electronics": []string{},
		"grocery":     []string{},
		"pharmacy":    []string{},
		"restaurant":  []string{},
	}
	n := len(code)
	for i := 0; i < n; i++ {
		if !isActive[i] {
			continue
		}
		if _, ok := busi[businessLine[i]]; !ok {
			continue
		}
		if len(code[i]) == 0 {
			continue
		}
		match := true
		for _, c := range code[i] {
			if c >= 'A' && c <= 'Z' {
				continue
			}
			if c >= 'a' && c <= 'z' {
				continue
			}
			if c >= '0' && c <= '9' {
				continue
			}
			if c == '_' {
				continue
			}
			match = false
			break
		}
		if !match {
			continue
		}
		busi[businessLine[i]] = append(busi[businessLine[i]], code[i])
	}
	ans := []string{}
	for k, v := range busi {
		sort.Strings(v)
		busi[k] = v
	}
	ans = append(ans, busi["electronics"]...)
	ans = append(ans, busi["grocery"]...)
	ans = append(ans, busi["pharmacy"]...)
	ans = append(ans, busi["restaurant"]...)
	return ans
}
