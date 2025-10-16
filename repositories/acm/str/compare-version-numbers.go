package str

import (
	"strconv"
	"strings"
)

func compareVersion(version1 string, version2 string) int {
	arr1 := strings.Split(version1, ".")
	arr2 := strings.Split(version2, ".")
	for i := len(arr1); i < len(arr2); i++ {
		arr1 = append(arr1, "0")
	}
	for i := len(arr2); i < len(arr1); i++ {
		arr2 = append(arr2, "0")
	}
	equal := func(s1, s2 string) int {
		for len(s1) > 0 {
			if s1[0] != '0' {
				break
			}
			s1 = s1[1:]
		}
		for len(s2) > 0 {
			if s2[0] != '0' {
				break
			}
			s2 = s2[1:]
		}
		if s1 == s2 {
			return 0
		}
		num1, _ := strconv.Atoi(s1)
		num2, _ := strconv.Atoi(s2)
		if num1 > num2 {
			return 1
		}
		return -1
	}
	n := len(arr1)
	for i := 0; i < n; i++ {
		val := equal(arr1[i], arr2[i])
		if val == 0 {
			continue
		}
		return val
	}
	return 0
}
