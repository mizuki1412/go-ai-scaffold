package storagekit

import (
	"github.com/example/go-ai-scaffold/pkg/class"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/cli/configkey"
	"github.com/example/go-ai-scaffold/pkg/library/filekit"
	"github.com/example/go-ai-scaffold/pkg/service/configkit"
)

// GetFullPath 将path或filename用项目目录包裹
func GetFullPath(path string) string {
	// P1 修复：空路径原会直接 index out of range panic，报错信息不可读
	if path == "" {
		panic(exception.New("storagekit: path 不能为空"))
	}
	if path[0] != '/' {
		path = "/" + path
	}
	p := configkit.GetString(configkey.ProjectDir, ".") + path
	return p
}

// checkAndCreateDir 确保父目录存在。
// P1 修复：原 `_ = CheckFilePath` 吞掉建目录错误，目录创建失败无感知
func checkAndCreateDir(path string) {
	if err := filekit.CheckFilePath(path); err != nil {
		panic(exception.New("创建文件目录失败: " + err.Error()))
	}
}

// SaveInHome 存入项目目录下, path是全路径（含文件名）
func SaveInHome(file *class.File, path string) {
	path = GetFullPath(path)
	checkAndCreateDir(path)
	// WriteClassFile 内部对打开/写入错误已 panic
	filekit.WriteClassFile(path, file)
}

func SaveBytesInHome(data []byte, path string) {
	path = GetFullPath(path)
	checkAndCreateDir(path)
	// P1 修复：原 `_ =` 吞掉写入错误，落盘失败无感知
	if err := filekit.WriteFile(path, data); err != nil {
		panic(exception.New("文件写入失败: " + err.Error()))
	}
}

func SaveBytesAppendInHome(data []byte, path string) {
	path = GetFullPath(path)
	checkAndCreateDir(path)
	if err := filekit.WriteFileAppend(path, data); err != nil {
		panic(exception.New("文件追加写入失败: " + err.Error()))
	}
}

func GetInHome(path string) []byte {
	path = GetFullPath(path)
	// ReadBytes 内部对读取错误记日志并返回 nil
	return filekit.ReadBytes(path)
}
