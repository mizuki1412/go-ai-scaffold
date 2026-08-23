package filekit

// 源码定位与文件查找（移植自 GoFrame os/gfile，MIT）。

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

var (
	mainPkgPathOnce sync.Mutex
	mainPkgPath     string
	packageMainRe   = regexp.MustCompile(`package\s+main\s+`)
)

// MainPkgPath 返回 package main 所在目录的绝对路径（仅开发期有效，结果缓存）。
// 注意：首次调用若在异步 goroutine 中可能取不到，取不到时会重试。
func MainPkgPath() string {
	mainPkgPathOnce.Lock()
	defer mainPkgPathOnce.Unlock()
	if mainPkgPath != "" {
		return mainPkgPath
	}
	goRoot := runtime.GOROOT()
	var lastFile string
	for i := 1; i < 10000; i++ {
		pc, file, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if goRoot != "" && len(file) >= len(goRoot) && file[:len(goRoot)] == goRoot {
			continue
		}
		if filepath.Ext(file) != ".go" {
			continue
		}
		lastFile = file
		if fn := runtime.FuncForPC(pc); fn != nil && strings.Split(fn.Name(), ".")[0] != "main" {
			continue
		}
		if content, err := os.ReadFile(file); err == nil && packageMainRe.Match(content) {
			mainPkgPath = filepath.Dir(file)
			return mainPkgPath
		}
	}
	// 兜底：从最后一个 .go 文件逐级向上找 package main（单测场景常用）
	if lastFile != "" {
		for dir := filepath.Dir(lastFile); len(dir) > 1; dir = filepath.Dir(dir) {
			files, _ := filepath.Glob(filepath.ToSlash(dir) + "/*.go")
			for _, v := range files {
				if content, err := os.ReadFile(v); err == nil && packageMainRe.Match(content) {
					mainPkgPath = dir
					return mainPkgPath
				}
			}
		}
	}
	return ""
}

// Search 按 优先路径 → Pwd → 本包目录 → MainPkgPath 的顺序查找文件，
// 找到返回其绝对路径；找不到返回列出全部搜索路径的错误。
func Search(name string, prioritySearchPaths ...string) (realPath string, err error) {
	if realPath = RealPath(name); realPath != "" {
		return realPath, nil
	}
	paths := append([]string{}, prioritySearchPaths...)
	pwd, _ := os.Getwd()
	if pwd != "" {
		paths = append(paths, pwd)
	}
	selfDir := selfFileDir()
	if selfDir != "" {
		paths = append(paths, selfDir)
	}
	if p := MainPkgPath(); p != "" {
		paths = append(paths, p)
	}
	// 去重保序
	seen := make(map[string]struct{}, len(paths))
	unique := paths[:0]
	for _, v := range paths {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		unique = append(unique, v)
	}
	for _, p := range unique {
		if realPath = RealPath(filepath.Join(p, name)); realPath != "" {
			return realPath, nil
		}
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "cannot find %q in following paths:", name)
	for k, v := range unique {
		fmt.Fprintf(&sb, "\n%d. %s", k+1, v)
	}
	return "", errors.New(sb.String())
}

// ReadLines 流式逐行读取大文件，行内容交给回调；回调返回 error 时中断。
// 注意 bufio.Scanner 默认单行上限 64KB，超长行会以错误终止读取。
func ReadLines(file string, callback func(line string) error) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err = callback(scanner.Text()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func selfFileDir() string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Dir(file)
}

// RealPath 返回存在的文件的绝对路径，不存在返回空串。
func RealPath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		if fi, err := os.Stat(abs); err == nil && fi != nil {
			return abs
		}
	}
	return ""
}
