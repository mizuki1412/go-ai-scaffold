package stringkit

import (
	"strconv"
	"strings"
)

// CompareVersion 按 GNU 版本号比较 a 与 b（移植自 GoFrame text/gstr，MIT）。
// 返回 1(a>b) / -1(a<b) / 0(相等)；支持 v1.0、1、1.0.0 等格式，v/V 前缀可省略。
func CompareVersion(a, b string) int {
	if a != "" && (a[0] == 'v' || a[0] == 'V') {
		a = a[1:]
	}
	if b != "" && (b[0] == 'v' || b[0] == 'V') {
		b = b[1:]
	}
	array1 := strings.Split(a, ".")
	array2 := strings.Split(b, ".")
	for len(array2) > len(array1) {
		array1 = append(array1, "0")
	}
	for len(array1) > len(array2) {
		array2 = append(array2, "0")
	}
	for i := 0; i < len(array1); i++ {
		v1 := versionSegmentToInt(array1[i])
		v2 := versionSegmentToInt(array2[i])
		if v1 > v2 {
			return 1
		}
		if v1 < v2 {
			return -1
		}
	}
	return 0
}

// 非法段（空/非数字）按 0 处理。
func versionSegmentToInt(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}
