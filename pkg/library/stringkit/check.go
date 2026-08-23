package stringkit

import "strings"

// 字符串判定与切分（移植自 GoFrame internal/utils，MIT）。

// IsNumeric 判断字符串是否为数字（含符号与至多一个小数点，如 "-12.34"）。
// 手写字符循环零分配。
func IsNumeric(s string) bool {
	var (
		dotCount = 0
		length   = len(s)
	)
	if length == 0 {
		return false
	}
	for i := range length {
		if (s[i] == '-' || s[i] == '+') && i == 0 {
			if length == 1 {
				return false
			}
			continue
		}
		if s[i] == '.' {
			dotCount++
			if i > 0 && i < length-1 && s[i-1] >= '0' && s[i-1] <= '9' {
				continue
			}
			return false
		}
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return dotCount <= 1
}

// RemoveSymbols 去掉所有符号，仅保留字母、数字与非 ASCII 字符。
func RemoveSymbols(s string) string {
	b := make([]rune, 0, len(s))
	for _, c := range s {
		if c > 127 {
			b = append(b, c)
		} else if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			b = append(b, c)
		}
	}
	return string(b)
}

// EqualFoldWithoutChars 忽略大小写并忽略 '-'/'_'/'.'/' ' 后比较两串是否相等。
// 典型场景：配置键/字段名的模糊匹配（"user_name" 与 "userName" 视为相等）。
func EqualFoldWithoutChars(s1, s2 string) bool {
	return strings.EqualFold(RemoveSymbols(s1), RemoveSymbols(s2))
}

// SplitAndTrim 按分隔符切割并 Trim 每段，忽略 Trim 后为空的段。
func SplitAndTrim(str, delimiter string) []string {
	array := make([]string, 0)
	for _, v := range strings.Split(str, delimiter) {
		v = strings.TrimSpace(v)
		if v != "" {
			array = append(array, v)
		}
	}
	return array
}
