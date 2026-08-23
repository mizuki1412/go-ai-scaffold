package filekit

// 支持 "**"（globstar）的 glob 匹配（移植自 GoFrame os/gfile，MIT）。
// 在 path.Match 基础上扩展：
//   - "*"   匹配单段内任意字符；"?" 匹配单字符；[abc]/[a-z]/[^abc] 同 path.Match
//   - "**"  仅作为完整路径段时跨目录匹配（如 "src/**/*.go"、"**/a"）；
//     "a**b" 等场景退化为两个普通 "*"
//   - "/" 与 "\" 均视为路径分隔符；坏模式（如未闭合 "["）返回 error

import (
	"path"
	"strings"
)

func MatchGlob(pattern, name string) (bool, error) {
	if !strings.Contains(pattern, "**") {
		return path.Match(pattern, name)
	}
	return matchGlobstar(pattern, name)
}

func matchGlobstar(pattern, name string) (bool, error) {
	pattern = strings.ReplaceAll(pattern, "\\", "/")
	name = strings.ReplaceAll(name, "\\", "/")
	pattern = path.Clean(pattern)
	name = path.Clean(name)
	if !hasValidGlobstar(pattern) {
		return path.Match(strings.ReplaceAll(pattern, "**", "*"), name)
	}
	memo := make(map[string]bool)
	return doMatchGlobstarMemo(pattern, name, memo)
}

// hasValidGlobstar 是否存在作为完整路径段的 "**"。
func hasValidGlobstar(pattern string) bool {
	idx := 0
	for {
		pos := strings.Index(pattern[idx:], "**")
		if pos == -1 {
			return false
		}
		pos += idx
		if isValidGlobstarAt(pattern, pos) {
			return true
		}
		idx = pos + 2
		if idx >= len(pattern) {
			break
		}
	}
	return false
}

func isValidGlobstarAt(pattern string, pos int) bool {
	if pos > 0 && pattern[pos-1] != '/' {
		return false
	}
	endPos := pos + 2
	if endPos < len(pattern) && pattern[endPos] != '/' {
		return false
	}
	return true
}

// findValidGlobstar 返回第一个合法 globstar 的位置，无则 -1。
func findValidGlobstar(pattern string) int {
	idx := 0
	for {
		pos := strings.Index(pattern[idx:], "**")
		if pos == -1 {
			return -1
		}
		pos += idx
		if isValidGlobstarAt(pattern, pos) {
			return pos
		}
		idx = pos + 2
		if idx >= len(pattern) {
			break
		}
	}
	return -1
}

func doMatchGlobstarMemo(pattern, name string, memo map[string]bool) (bool, error) {
	cacheKey := pattern + "\x00" + name
	if cached, ok := memo[cacheKey]; ok {
		return cached, nil
	}
	result, err := doMatchGlobstarCore(pattern, name, memo)
	if err != nil {
		return false, err
	}
	memo[cacheKey] = result
	return result, nil
}

func doMatchGlobstarCore(pattern, name string, memo map[string]bool) (bool, error) {
	pos := findValidGlobstar(pattern)
	if pos == -1 {
		normalizedPattern := strings.ReplaceAll(pattern, "**", "*")
		return path.Match(normalizedPattern, name)
	}

	prefix := strings.TrimSuffix(pattern[:pos], "/")
	suffix := strings.TrimPrefix(pattern[pos+2:], "/")

	if prefix != "" {
		if !strings.ContainsAny(prefix, "*?[") {
			// 字面前缀：要求在分隔符边界处结束
			if !strings.HasPrefix(name, prefix) {
				return false, nil
			}
			if len(name) == len(prefix) {
				name = ""
			} else {
				if name[len(prefix)] != '/' {
					return false, nil
				}
				name = name[len(prefix)+1:]
			}
		} else {
			// 带通配符前缀：逐段匹配
			prefixParts := strings.Split(prefix, "/")
			nameParts := strings.Split(name, "/")
			if len(nameParts) < len(prefixParts) {
				return false, nil
			}
			for i, pp := range prefixParts {
				matched, err := path.Match(pp, nameParts[i])
				if err != nil {
					return false, err
				}
				if !matched {
					return false, nil
				}
			}
			name = strings.Join(nameParts[len(prefixParts):], "/")
		}
	}

	if suffix == "" {
		return true, nil
	}
	if name == "" {
		return doMatchGlobstarMemo(suffix, "", memo)
	}
	nameParts := strings.Split(name, "/")
	for i := 0; i <= len(nameParts); i++ {
		remaining := strings.Join(nameParts[i:], "/")
		matched, err := doMatchGlobstarMemo(suffix, remaining, memo)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}
